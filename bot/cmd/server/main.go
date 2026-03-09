package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type app struct {
	apiBaseURL  string
	calendarURL string
	botToken    string
	botSecret   string
	client      *http.Client
}

type telegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *telegramMessage       `json:"message"`
	CallbackQuery *telegramCallbackQuery `json:"callback_query"`
}

type telegramMessage struct {
	Chat telegramChat  `json:"chat"`
	From *telegramUser `json:"from"`
	Text string        `json:"text"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type telegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    *telegramUser    `json:"from"`
	Message *telegramMessage `json:"message"`
	Data    string           `json:"data"`
}

type exchangeRequest struct {
	TelegramID       int64  `json:"telegramId"`
	TelegramUsername string `json:"telegramUsername"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
}

type exchangeSuccessResponse struct {
	MCPToken   string `json:"mcpToken"`
	WasCreated bool   `json:"wasCreated"`
}

type accessSuccessResponse struct {
	Allowed          bool   `json:"allowed"`
	HasConnection    bool   `json:"hasConnection"`
	TelegramUsername string `json:"telegramUsername"`
}

type telegramGetUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Result      []telegramUpdate `json:"result"`
	Description string           `json:"description"`
}

type telegramBaseResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

const (
	callbackIssueToken = "issue_token"
	supportContact     = "@xigax"
)

func main() {
	apiBaseURL := strings.TrimRight(os.Getenv("API_BASE_URL"), "/")
	if apiBaseURL == "" {
		log.Fatal("API_BASE_URL is required")
	}
	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}
	botSecret := strings.TrimSpace(os.Getenv("BOT_SECRET"))
	if botSecret == "" {
		log.Fatal("BOT_SECRET is required")
	}
	calendarURL := strings.TrimSpace(os.Getenv("OPEN_CALENDAR_REDIRECT_URL"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	application := &app{
		apiBaseURL:  apiBaseURL,
		calendarURL: calendarURL,
		botToken:    botToken,
		botSecret:   botSecret,
		client:      &http.Client{Timeout: 35 * time.Second},
	}

	ctx := context.Background()
	if err := application.deleteWebhook(ctx); err != nil {
		log.Fatalf("delete webhook: %v", err)
	}
	go application.pollUpdates(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", application.healthz)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("telegram bot webhook server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func (a *app) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *app) handleMessage(ctx context.Context, message telegramMessage) {
	if message.From == nil {
		return
	}
	chatID := message.Chat.ID
	payload := requestFromUser(message.From)

	statusCode, body, err := a.callAPIAccess(ctx, payload)
	if err != nil {
		log.Printf("access failed: %v", err)
		_ = a.sendTelegramMessage(ctx, chatID, textServiceUnavailable)
		return
	}

	switch {
	case statusCode == http.StatusForbidden:
		_ = a.sendTelegramMessage(ctx, chatID, textNotAllowlisted(supportContact))
	case statusCode >= http.StatusBadRequest:
		log.Printf("access error status=%d body=%s", statusCode, strings.TrimSpace(string(body)))
		_ = a.sendTelegramMessage(ctx, chatID, textAllowlistCheckFail)
	default:
		var result accessSuccessResponse
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("access parse failed: %v", err)
			_ = a.sendTelegramMessage(ctx, chatID, textAllowlistParseFail)
			return
		}
		_ = a.sendTelegramMessageWithActionButtons(ctx, chatID, textAllowlisted, result.HasConnection)
	}
}

func (a *app) handleCallbackQuery(ctx context.Context, callback telegramCallbackQuery) {
	if strings.TrimSpace(callback.ID) != "" {
		_ = a.answerCallbackQuery(ctx, callback.ID)
	}
	if callback.Message == nil || callback.From == nil {
		return
	}
	chatID := callback.Message.Chat.ID
	switch strings.TrimSpace(callback.Data) {
	case callbackIssueToken:
		a.handleIssueTokenCallback(ctx, chatID, callback.From)
	}
}

func (a *app) handleIssueTokenCallback(ctx context.Context, chatID int64, user *telegramUser) {
	payload := requestFromUser(user)
	statusCode, body, err := a.callAPIExchange(ctx, payload)
	if err != nil {
		log.Printf("exchange failed: %v", err)
		_ = a.sendTelegramMessage(ctx, chatID, textServiceUnavailable)
		return
	}

	switch {
	case statusCode == http.StatusForbidden:
		_ = a.sendTelegramMessage(ctx, chatID, textNotAllowlisted(supportContact))
	case statusCode >= http.StatusBadRequest:
		log.Printf("exchange error status=%d body=%s", statusCode, strings.TrimSpace(string(body)))
		_ = a.sendTelegramMessage(ctx, chatID, textIssueTokenFail)
	default:
		var result exchangeSuccessResponse
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("exchange parse failed: %v", err)
			_ = a.sendTelegramMessage(ctx, chatID, textTokenParseFail)
			return
		}
		if strings.TrimSpace(result.MCPToken) == "" {
			_ = a.sendTelegramMessage(ctx, chatID, textTokenEmpty)
			return
		}
		_ = a.sendTelegramMessage(ctx, chatID, textTokenIssued(result.MCPToken, onboardingStepURL(a.calendarURL, 4)))
		_ = a.sendTelegramMessageWithActionButtons(ctx, chatID, textAllowlisted, true)
	}
}

func (a *app) pollUpdates(ctx context.Context) {
	var offset int64
	for {
		updates, err := a.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("getUpdates failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if update.Message != nil {
				a.handleMessage(ctx, *update.Message)
			}
			if update.CallbackQuery != nil {
				a.handleCallbackQuery(ctx, *update.CallbackQuery)
			}
		}
	}
}

func (a *app) getUpdates(ctx context.Context, offset int64) ([]telegramUpdate, error) {
	payload := map[string]any{
		"timeout": 30,
		"offset":  offset,
	}
	body, err := a.callTelegram(ctx, "getUpdates", payload)
	if err != nil {
		return nil, err
	}
	var response telegramGetUpdatesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse getUpdates response: %w", err)
	}
	if !response.OK {
		return nil, fmt.Errorf("getUpdates not ok: %s", strings.TrimSpace(response.Description))
	}
	return response.Result, nil
}

func (a *app) deleteWebhook(ctx context.Context) error {
	body, err := a.callTelegram(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": false})
	if err != nil {
		return err
	}
	var response telegramBaseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("parse deleteWebhook response: %w", err)
	}
	if !response.OK {
		return fmt.Errorf("deleteWebhook not ok: %s", strings.TrimSpace(response.Description))
	}
	return nil
}

func (a *app) callAPIExchange(ctx context.Context, payload exchangeRequest) (int, []byte, error) {
	return a.callAPI(ctx, "/v1/private/telegram/exchange", payload)
}

func (a *app) callAPIAccess(ctx context.Context, payload exchangeRequest) (int, []byte, error) {
	return a.callAPI(ctx, "/v1/private/telegram/access", payload)
}

func (a *app) callAPI(ctx context.Context, path string, payload exchangeRequest) (int, []byte, error) {
	rawBody, _ := json.Marshal(payload)
	targetURL := a.apiBaseURL + path
	outReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(rawBody))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	outReq.Header.Set("Content-Type", "application/json")
	outReq.Header.Set("X-Bot-Secret", a.botSecret)

	response, err := a.client.Do(outReq)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer response.Body.Close()

	respBody, _ := io.ReadAll(response.Body)
	return response.StatusCode, respBody, nil
}

func requestFromUser(user *telegramUser) exchangeRequest {
	return exchangeRequest{
		TelegramID:       user.ID,
		TelegramUsername: user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
	}
}

func onboardingStepURL(baseURL string, step int) string {
	trimmedBaseURL := strings.TrimSpace(baseURL)
	if trimmedBaseURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(trimmedBaseURL)
	if err != nil {
		return ""
	}

	query := parsedURL.Query()
	query.Set("onboarding", "1")
	query.Set("step", strconv.Itoa(step))
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

func (a *app) sendTelegramMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]any{
		"chat_id":    strconv.FormatInt(chatID, 10),
		"text":       text,
		"parse_mode": "Markdown",
	}
	_, err := a.callTelegram(ctx, "sendMessage", payload)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) sendTelegramMessageWithActionButtons(ctx context.Context, chatID int64, text string, hasConnection bool) error {
	tokenButtonText := buttonGetToken
	if hasConnection {
		tokenButtonText = buttonReissueToken
	}

	buttonRows := [][]map[string]string{
		{
			{"text": tokenButtonText, "callback_data": callbackIssueToken},
		},
	}
	if strings.TrimSpace(a.calendarURL) != "" {
		buttonRows = append(buttonRows, []map[string]string{
			{"text": buttonOpenCalendar, "url": a.calendarURL},
		})
	}

	payload := map[string]any{
		"chat_id": strconv.FormatInt(chatID, 10),
		"text":    text,
		"reply_markup": map[string]any{
			"inline_keyboard": buttonRows,
		},
	}
	_, err := a.callTelegram(ctx, "sendMessage", payload)
	return err
}

func (a *app) answerCallbackQuery(ctx context.Context, callbackID string) error {
	_, err := a.callTelegram(ctx, "answerCallbackQuery", map[string]any{
		"callback_query_id": callbackID,
	})
	return err
}

func (a *app) callTelegram(ctx context.Context, method string, payload map[string]any) ([]byte, error) {
	targetURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", a.botToken, method)
	rawBody, _ := json.Marshal(payload)
	outReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	outReq.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(outReq)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("telegram %s failed: status=%d body=%s", method, response.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
