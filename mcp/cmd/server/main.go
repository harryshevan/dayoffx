package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type app struct {
	apiBaseURL string
	client     *http.Client
}

func main() {
	apiBaseURL := strings.TrimRight(os.Getenv("API_BASE_URL"), "/")
	if apiBaseURL == "" {
		log.Fatal("API_BASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	application := &app{
		apiBaseURL: apiBaseURL,
		client:     &http.Client{Timeout: 15 * time.Second},
	}

	mcpServer := server.NewMCPServer(
		"dayoffs-mcp",
		"2.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	application.registerTools(mcpServer)

	streamable := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath("/mcp"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/mcp", requireBearer(streamable))
	mux.Handle("/mcp/", requireBearer(streamable))

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("mcp server listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func requireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if extractBearer(r.Header) == "" {
			writeUnauthorized(w, "missing_or_invalid_authorization")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) registerTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool(
			"createVacation",
			mcp.WithDescription("Create a vacation request"),
			mcp.WithString("from", mcp.Required(), mcp.Description("Start date YYYY-MM-DD")),
			mcp.WithString("to", mcp.Required(), mcp.Description("End date YYYY-MM-DD")),
			mcp.WithString("reason", mcp.Required(), mcp.Description("Vacation reason")),
			mcp.WithString("displayName", mcp.Description("Optional new display name")),
			mcp.WithString("colorHex", mcp.Description("Optional profile color #RRGGBB")),
		),
		a.createVacation,
	)
	s.AddTool(
		mcp.NewTool(
			"changeVacation",
			mcp.WithDescription("Change existing vacation by id"),
			mcp.WithString("vacationId", mcp.Required(), mcp.Description("Vacation id (uuid)")),
			mcp.WithString("newFrom", mcp.Required(), mcp.Description("New start date YYYY-MM-DD")),
			mcp.WithString("newTo", mcp.Required(), mcp.Description("New end date YYYY-MM-DD")),
			mcp.WithString("newReason", mcp.Required(), mcp.Description("New vacation reason")),
			mcp.WithString("displayName", mcp.Description("Optional new display name")),
			mcp.WithString("colorHex", mcp.Description("Optional profile color #RRGGBB")),
		),
		a.changeVacation,
	)
	s.AddTool(
		mcp.NewTool(
			"removeVacation",
			mcp.WithDescription("Remove existing vacation by id"),
			mcp.WithString("vacationId", mcp.Required(), mcp.Description("Vacation id (uuid)")),
			mcp.WithString("displayName", mcp.Description("Optional new display name")),
			mcp.WithString("colorHex", mcp.Description("Optional profile color #RRGGBB")),
		),
		a.removeVacation,
	)
	s.AddTool(
		mcp.NewTool(
			"approveVacation",
			mcp.WithDescription("Approve vacation by id, admin only"),
			mcp.WithString("vacationId", mcp.Required(), mcp.Description("Vacation id (uuid)")),
		),
		a.approveVacation,
	)
	s.AddTool(
		mcp.NewTool(
			"issueToken",
			mcp.WithDescription("Issue new member token, env-admin only"),
			mcp.WithString("displayName", mcp.Description("Optional member display name")),
			mcp.WithString("colorHex", mcp.Description("Optional member color #RRGGBB")),
			mcp.WithString("goal", mcp.Description("Optional goal: cursor|claude_desktop|other")),
		),
		a.issueToken,
	)
	s.AddTool(
		mcp.NewTool(
			"revokeToken",
			mcp.WithDescription("Revoke member token, env-admin only"),
			mcp.WithString("token", mcp.Required(), mcp.Description("Member token to revoke")),
		),
		a.revokeToken,
	)
}

func (a *app) createVacation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	from, err := req.RequireString("from")
	if err != nil || strings.TrimSpace(from) == "" {
		return mcp.NewToolResultError("from_required"), nil
	}
	to, err := req.RequireString("to")
	if err != nil || strings.TrimSpace(to) == "" {
		return mcp.NewToolResultError("to_required"), nil
	}
	reason, err := req.RequireString("reason")
	if err != nil || strings.TrimSpace(reason) == "" {
		return mcp.NewToolResultError("reason_required"), nil
	}

	payload := map[string]any{
		"fromDate": from,
		"toDate":   to,
		"reason":   reason,
	}

	headers := profileHeaders(req)
	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodPost, "/v1/mcp/createVacation", url.Values{}, payload, headers)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) changeVacation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vacationID, err := req.RequireString("vacationId")
	if err != nil || strings.TrimSpace(vacationID) == "" {
		return mcp.NewToolResultError("vacationId_required"), nil
	}
	newFrom, err := req.RequireString("newFrom")
	if err != nil || strings.TrimSpace(newFrom) == "" {
		return mcp.NewToolResultError("newFrom_required"), nil
	}
	newTo, err := req.RequireString("newTo")
	if err != nil || strings.TrimSpace(newTo) == "" {
		return mcp.NewToolResultError("newTo_required"), nil
	}
	newReason, err := req.RequireString("newReason")
	if err != nil || strings.TrimSpace(newReason) == "" {
		return mcp.NewToolResultError("newReason_required"), nil
	}

	payload := map[string]any{
		"newFrom":   newFrom,
		"newTo":     newTo,
		"newReason": newReason,
	}
	query := url.Values{}
	query.Set("vacationId", vacationID)

	headers := profileHeaders(req)
	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodPatch, "/v1/mcp/changeVacation", query, payload, headers)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) removeVacation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vacationID, err := req.RequireString("vacationId")
	if err != nil || strings.TrimSpace(vacationID) == "" {
		return mcp.NewToolResultError("vacationId_required"), nil
	}

	query := url.Values{}
	query.Set("vacationId", vacationID)

	headers := profileHeaders(req)
	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodDelete, "/v1/mcp/removeVacation", query, map[string]any{}, headers)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) approveVacation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vacationID, err := req.RequireString("vacationId")
	if err != nil || strings.TrimSpace(vacationID) == "" {
		return mcp.NewToolResultError("vacationId_required"), nil
	}
	query := url.Values{}
	query.Set("vacationId", vacationID)
	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodPost, "/v1/mcp/approveVacation", query, map[string]any{}, nil)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) issueToken(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]any{}
	if value := strings.TrimSpace(req.GetString("displayName", "")); value != "" {
		payload["displayName"] = value
	}
	if value := strings.TrimSpace(req.GetString("colorHex", "")); value != "" {
		payload["colorHex"] = value
	}
	if value := strings.TrimSpace(req.GetString("goal", "")); value != "" {
		payload["goal"] = value
	}

	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodPost, "/v1/mcp/issueToken", url.Values{}, payload, nil)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) revokeToken(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token, err := req.RequireString("token")
	if err != nil || strings.TrimSpace(token) == "" {
		return mcp.NewToolResultError("token_required"), nil
	}
	payload := map[string]any{"token": token}

	statusCode, body, callErr := a.callAPI(ctx, req, http.MethodPost, "/v1/mcp/revokeToken", url.Values{}, payload, nil)
	return apiResult(statusCode, body, callErr), nil
}

func (a *app) callAPI(
	ctx context.Context,
	req mcp.CallToolRequest,
	method string,
	path string,
	query url.Values,
	payload map[string]any,
	extraHeaders map[string]string,
) (int, []byte, error) {
	authValue := req.Header.Get("Authorization")
	if extractBearer(req.Header) == "" {
		return http.StatusUnauthorized, nil, errors.New("missing_or_invalid_authorization")
	}

	targetURL := a.apiBaseURL + path
	if encoded := query.Encode(); encoded != "" {
		targetURL += "?" + encoded
	}

	body, _ := json.Marshal(payload)
	outReq, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	outReq.Header.Set("Content-Type", "application/json")
	outReq.Header.Set("Authorization", authValue)
	for k, v := range extraHeaders {
		outReq.Header.Set(k, v)
	}

	response, err := a.client.Do(outReq)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer response.Body.Close()

	respBody, _ := io.ReadAll(response.Body)
	return response.StatusCode, respBody, nil
}

func profileHeaders(req mcp.CallToolRequest) map[string]string {
	headers := map[string]string{}
	if value := strings.TrimSpace(req.GetString("displayName", "")); value != "" {
		headers["X-Profile-Name"] = value
	}
	if value := strings.TrimSpace(req.GetString("colorHex", "")); value != "" {
		headers["X-Profile-Color"] = strings.ToLower(value)
	}
	return headers
}

func extractBearer(headers http.Header) string {
	value := strings.TrimSpace(headers.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
	if token == "" {
		return ""
	}
	return token
}

func apiResult(statusCode int, body []byte, err error) *mcp.CallToolResult {
	if err != nil {
		return mcp.NewToolResultError(err.Error())
	}
	if statusCode >= http.StatusBadRequest {
		text := strings.TrimSpace(string(body))
		if text == "" {
			text = http.StatusText(statusCode)
		}
		return mcp.NewToolResultError(text)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return mcp.NewToolResultText("{}")
	}
	return mcp.NewToolResultText(string(body))
}

func writeUnauthorized(w http.ResponseWriter, errorCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": errorCode})
}
