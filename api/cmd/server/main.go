package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type app struct {
	db                  *pgxpool.Pool
	adminSecret         string
	botSecret           string
	tokenPepper         string
	legacyTokenFallback bool
}

type privateCreateUserRequest struct {
	DisplayName string `json:"displayName"`
	ColorHex    string `json:"colorHex"`
	Goal        string `json:"goal"`
}

type createUserResponse struct {
	MemberID     string `json:"memberId"`
	DisplayName  string `json:"displayName"`
	ColorHex     string `json:"colorHex"`
	Role         string `json:"role"`
	MCPToken     string `json:"mcpToken"`
	MCPServerURL string `json:"mcpServerUrl"`
}

type vacationDTO struct {
	ID          string `json:"id"`
	MemberID    string `json:"memberId"`
	DisplayName string `json:"displayName"`
	ColorHex    string `json:"colorHex"`
	FromDate    string `json:"fromDate"`
	ToDate      string `json:"toDate"`
	Reason      string `json:"reason"`
	Status      string `json:"status"`
}

type createVacationRequest struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
	Reason   string `json:"reason"`
}

type changeVacationRequest struct {
	NewFrom   string `json:"newFrom"`
	NewTo     string `json:"newTo"`
	NewReason string `json:"newReason"`
}

type changeColorRequest struct {
	NewColor string `json:"newColor"`
}

type changeNameRequest struct {
	NewName string `json:"newName"`
}

type issueTokenRequest struct {
	DisplayName string `json:"displayName"`
	ColorHex    string `json:"colorHex"`
	Goal        string `json:"goal"`
}

type issueTokenResponse struct {
	MemberID    string `json:"memberId"`
	DisplayName string `json:"displayName"`
	ColorHex    string `json:"colorHex"`
	MCPToken    string `json:"mcpToken"`
}

type privateTelegramWhitelistRequest struct {
	TelegramUsername string `json:"telegramUsername"`
}

type privateTelegramWhitelistResponse struct {
	TelegramUsername string `json:"telegramUsername"`
}

type privateTelegramExchangeRequest struct {
	TelegramID       int64  `json:"telegramId"`
	TelegramUsername string `json:"telegramUsername"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
}

type privateTelegramAccessResponse struct {
	Allowed          bool   `json:"allowed"`
	TelegramUsername string `json:"telegramUsername"`
}

type privateTelegramExchangeResponse struct {
	MemberID         string `json:"memberId"`
	DisplayName      string `json:"displayName"`
	ColorHex         string `json:"colorHex"`
	MCPToken         string `json:"mcpToken"`
	MCPServerURL     string `json:"mcpServerUrl"`
	WasCreated       bool   `json:"wasCreated"`
	TelegramUsername string `json:"telegramUsername"`
}

type telegramWhitelistRecord struct {
	ID               uuid.UUID
	MemberID         *uuid.UUID
	TelegramID       *int64
	TelegramUsername string
}

type revokeTokenRequest struct {
	Token string `json:"token"`
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

const (
	memberTokenPrefix = "dof"
	memberTokenV1     = "v1"
	defaultGoal       = "other"
	defaultRole       = "member"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	handler := &app{
		db:                  db,
		adminSecret:         os.Getenv("ADMIN_SECRET"),
		botSecret:           strings.TrimSpace(os.Getenv("BOT_SECRET")),
		tokenPepper:         strings.TrimSpace(os.Getenv("TOKEN_PEPPER")),
		legacyTokenFallback: true,
	}
	if handler.tokenPepper == "" {
		log.Fatal("TOKEN_PEPPER is required")
	}
	if flagValue := strings.TrimSpace(os.Getenv("LEGACY_TOKEN_FALLBACK")); flagValue != "" {
		enabled, err := strconv.ParseBool(flagValue)
		if err != nil {
			log.Fatalf("LEGACY_TOKEN_FALLBACK must be a boolean: %v", err)
		}
		handler.legacyTokenFallback = enabled
	}
	if handler.legacyTokenFallback {
		exists, err := legacyMCPTokenColumnExists(ctx, db)
		if err != nil {
			log.Fatalf("check legacy token column: %v", err)
		}
		handler.legacyTokenFallback = exists
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.healthz)
	mux.HandleFunc("/v1/private/users", handler.auth(handler.createPrivateUser))
	mux.HandleFunc("/v1/private/telegram/whitelist", handler.privateTelegramWhitelist)
	mux.HandleFunc("/v1/private/telegram/access", handler.privateTelegramAccess)
	mux.HandleFunc("/v1/private/telegram/exchange", handler.privateTelegramExchange)
	mux.HandleFunc("/v1/vacations", handler.listVacations)
	mux.HandleFunc("/v1/mcp/createVacation", handler.auth(handler.createVacation))
	mux.HandleFunc("/v1/mcp/changeVacation", handler.auth(handler.changeVacation))
	mux.HandleFunc("/v1/mcp/removeVacation", handler.auth(handler.removeVacation))
	mux.HandleFunc("/v1/mcp/changeColor", handler.auth(handler.changeColor))
	mux.HandleFunc("/v1/mcp/changeName", handler.auth(handler.changeName))
	mux.HandleFunc("/v1/mcp/approveVacation", handler.auth(handler.approveVacation))
	mux.HandleFunc("/v1/mcp/issueToken", handler.auth(handler.issueToken))
	mux.HandleFunc("/v1/mcp/revokeToken", handler.auth(handler.revokeToken))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("api listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-MCP-Token, X-Profile-Name, X-Profile-Color")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *app) createPrivateUser(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "env_admin_only")
		return
	}

	var req privateCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = fmt.Sprintf("member-%s", randomSuffix(6))
	}

	goal := strings.TrimSpace(req.Goal)
	if goal == "" {
		goal = defaultGoal
	}

	colorHex := strings.ToLower(strings.TrimSpace(req.ColorHex))
	if colorHex == "" {
		colorHex = randomColorHex()
	}
	if !colorPattern.MatchString(colorHex) {
		writeError(w, http.StatusBadRequest, "invalid_color")
		return
	}

	result, statusCode, errorCode := a.createMemberConnection(r.Context(), displayName, colorHex, goal, defaultRole)
	if errorCode != "" {
		writeError(w, statusCode, errorCode)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, result)
}

func (a *app) privateTelegramWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !a.isEnvAdminRequest(r) {
		writeError(w, http.StatusForbidden, "env_admin_only")
		return
	}

	var req privateTelegramWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	telegramUsername := normalizeTelegramUsername(req.TelegramUsername)
	if telegramUsername == "" {
		writeError(w, http.StatusBadRequest, "telegram_username_required")
		return
	}

	result, statusCode, errorCode := a.upsertTelegramWhitelist(r.Context(), telegramUsername)
	if errorCode != "" {
		writeError(w, statusCode, errorCode)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *app) privateTelegramExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !a.isBotOrEnvAdminRequest(r) {
		writeError(w, http.StatusUnauthorized, "bot_or_env_admin_required")
		return
	}

	var req privateTelegramExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	result, statusCode, errorCode := a.exchangeTelegramForToken(r.Context(), req)
	if errorCode != "" {
		writeError(w, statusCode, errorCode)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, result)
}

func (a *app) privateTelegramAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !a.isBotOrEnvAdminRequest(r) {
		writeError(w, http.StatusUnauthorized, "bot_or_env_admin_required")
		return
	}

	var req privateTelegramExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	result, statusCode, errorCode := a.checkTelegramAccess(r.Context(), req)
	if errorCode != "" {
		writeError(w, statusCode, errorCode)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *app) isEnvAdminRequest(r *http.Request) bool {
	if a.adminSecret == "" {
		return false
	}
	token := extractRequestToken(r)
	return token != "" && token == a.adminSecret
}

func (a *app) isBotOrEnvAdminRequest(r *http.Request) bool {
	if a.isEnvAdminRequest(r) {
		return true
	}
	if a.botSecret == "" {
		return false
	}
	botSecret := strings.TrimSpace(r.Header.Get("X-Bot-Secret"))
	if botSecret == "" {
		return false
	}
	return hmac.Equal([]byte(botSecret), []byte(a.botSecret))
}

func extractRequestToken(r *http.Request) string {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-MCP-Token"))
	}
	return token
}

func normalizeTelegramUsername(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	value = strings.TrimPrefix(value, "@")
	return value
}

func (a *app) listVacations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	yearValue := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_year")
		return
	}

	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	rows, err := a.db.Query(
		r.Context(),
		`select v.id, v.member_id, m.display_name, m.color_hex, v.from_date, v.to_date, v.reason, v.status
		 from vacations v
		 join members m on m.id = v.member_id
		 join connections c on c.member_id = m.id and c.active = true
		 where v.to_date >= $1 and v.from_date <= $2
		 order by v.from_date asc`,
		from, to,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vacation_query_failed")
		return
	}
	defer rows.Close()

	result := make([]vacationDTO, 0)
	for rows.Next() {
		var (
			item             vacationDTO
			id               uuid.UUID
			memberID         uuid.UUID
			fromDate, toDate time.Time
		)
		if err := rows.Scan(&id, &memberID, &item.DisplayName, &item.ColorHex, &fromDate, &toDate, &item.Reason, &item.Status); err != nil {
			writeError(w, http.StatusInternalServerError, "vacation_scan_failed")
			return
		}
		item.ID = id.String()
		item.MemberID = memberID.String()
		item.FromDate = fromDate.Format("2006-01-02")
		item.ToDate = toDate.Format("2006-01-02")
		result = append(result, item)
	}

	writeJSON(w, http.StatusOK, result)
}

type authContext struct {
	MemberID     uuid.UUID
	ConnectionID uuid.UUID
	Role         string
	IsEnvAdmin   bool
	CurrentName  string
	CurrentColor string
	AuthTokenRaw string
}

func (a *app) auth(next func(http.ResponseWriter, *http.Request, authContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = r.Header.Get("X-MCP-Token")
		}
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing_token")
			return
		}

		if a.adminSecret != "" && token == a.adminSecret {
			next(w, r, authContext{
				Role:         "admin",
				IsEnvAdmin:   true,
				AuthTokenRaw: token,
			})
			return
		}

		var ctx authContext
		if tokenID, tokenSecret, ok := parseMemberToken(token); ok {
			var storedHash *string
			err := a.db.QueryRow(
				r.Context(),
				`select m.id, c.id, m.role, m.display_name, m.color_hex, c.token_hash
				 from connections c
				 join members m on m.id = c.member_id
				 where c.token_id = $1 and c.active = true`,
				tokenID,
			).Scan(&ctx.MemberID, &ctx.ConnectionID, &ctx.Role, &ctx.CurrentName, &ctx.CurrentColor, &storedHash)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					writeError(w, http.StatusUnauthorized, "invalid_token")
					return
				}
				writeError(w, http.StatusInternalServerError, "auth_failed")
				return
			}
			if storedHash == nil || !verifyMemberTokenHash(a.tokenPepper, tokenID, tokenSecret, *storedHash) {
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}
		} else {
			if !a.legacyTokenFallback {
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}
			err := a.db.QueryRow(
				r.Context(),
				`select m.id, c.id, m.role, m.display_name, m.color_hex
				 from connections c
				 join members m on m.id = c.member_id
				 where c.mcp_token = $1 and c.active = true`,
				token,
			).Scan(&ctx.MemberID, &ctx.ConnectionID, &ctx.Role, &ctx.CurrentName, &ctx.CurrentColor)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					writeError(w, http.StatusUnauthorized, "invalid_token")
					return
				}
				writeError(w, http.StatusInternalServerError, "auth_failed")
				return
			}
		}
		ctx.AuthTokenRaw = token

		next(w, r, ctx)
	}
}

func parseProfileHeaders(r *http.Request) (string, string) {
	name := strings.TrimSpace(r.Header.Get("X-Profile-Name"))
	color := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Profile-Color")))
	return name, color
}

func (a *app) syncProfileFromHeaders(w http.ResponseWriter, r *http.Request, auth authContext) bool {
	if auth.IsEnvAdmin {
		return true
	}

	nextName, nextColor := parseProfileHeaders(r)
	if nextName == "" && nextColor == "" {
		return true
	}

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_begin_failed")
		return false
	}
	defer tx.Rollback(r.Context())

	if nextName != "" && nextName != auth.CurrentName {
		if _, err := tx.Exec(
			r.Context(),
			`update members set display_name = $1 where id = $2`,
			nextName, auth.MemberID,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "profile_name_update_failed")
			return false
		}
	}

	if nextColor != "" {
		if !colorPattern.MatchString(nextColor) {
			writeError(w, http.StatusBadRequest, "invalid_profile_color")
			return false
		}
		if nextColor != auth.CurrentColor {
			if _, err := tx.Exec(
				r.Context(),
				`update members set color_hex = $1 where id = $2`,
				nextColor, auth.MemberID,
			); err != nil {
				writeError(w, http.StatusInternalServerError, "profile_color_update_failed")
				return false
			}
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "db_commit_failed")
		return false
	}
	return true
}

func (a *app) createVacation(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "member_token_required")
		return
	}
	if !a.syncProfileFromHeaders(w, r, auth) {
		return
	}

	var req createVacationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	fromDate, toDate, ok := parseDateRange(req.FromDate, req.ToDate)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date_range")
		return
	}

	var id uuid.UUID
	err := a.db.QueryRow(
		r.Context(),
		`insert into vacations (id, member_id, from_date, to_date, reason, status, created_at, updated_at)
		 values ($1, $2, $3, $4, $5, 'pending', now(), now())
		 returning id`,
		uuid.New(), auth.MemberID, fromDate, toDate, strings.TrimSpace(req.Reason),
	).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vacation_create_failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"vacationId": id.String()})
}

func (a *app) changeVacation(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "member_token_required")
		return
	}
	if !a.syncProfileFromHeaders(w, r, auth) {
		return
	}

	vacationIDRaw := strings.TrimSpace(r.URL.Query().Get("vacationId"))
	vacationID, err := uuid.Parse(vacationIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_vacation_id")
		return
	}

	var req changeVacationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	fromDate, toDate, ok := parseDateRange(req.NewFrom, req.NewTo)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date_range")
		return
	}

	commandTag, err := a.db.Exec(
		r.Context(),
		`update vacations
		 set from_date = $1, to_date = $2, reason = $3, status = 'pending', updated_at = now()
		 where id = $4 and member_id = $5`,
		fromDate, toDate, strings.TrimSpace(req.NewReason), vacationID, auth.MemberID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vacation_change_failed")
		return
	}

	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "vacation_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"vacationId": vacationID.String()})
}

func (a *app) removeVacation(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "member_token_required")
		return
	}
	if !a.syncProfileFromHeaders(w, r, auth) {
		return
	}

	vacationIDRaw := strings.TrimSpace(r.URL.Query().Get("vacationId"))
	vacationID, err := uuid.Parse(vacationIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_vacation_id")
		return
	}

	commandTag, err := a.db.Exec(
		r.Context(),
		`delete from vacations where id = $1 and member_id = $2`,
		vacationID, auth.MemberID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vacation_remove_failed")
		return
	}

	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "vacation_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"vacationId": vacationID.String()})
}

func (a *app) changeColor(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "member_token_required")
		return
	}

	var req changeColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	newColor := strings.ToLower(strings.TrimSpace(req.NewColor))
	if !colorPattern.MatchString(newColor) {
		writeError(w, http.StatusBadRequest, "invalid_color")
		return
	}

	commandTag, err := a.db.Exec(
		r.Context(),
		`update members set color_hex = $1 where id = $2`,
		newColor, auth.MemberID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_color_update_failed")
		return
	}
	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "member_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"colorHex": newColor})
}

func (a *app) changeName(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "member_token_required")
		return
	}

	var req changeNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	newName := strings.TrimSpace(req.NewName)
	if newName == "" {
		writeError(w, http.StatusBadRequest, "invalid_name")
		return
	}

	commandTag, err := a.db.Exec(
		r.Context(),
		`update members set display_name = $1 where id = $2`,
		newName, auth.MemberID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_name_update_failed")
		return
	}
	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "member_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"displayName": newName})
}

func (a *app) approveVacation(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if auth.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin_only")
		return
	}

	vacationIDRaw := strings.TrimSpace(r.URL.Query().Get("vacationId"))
	vacationID, err := uuid.Parse(vacationIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_vacation_id")
		return
	}

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_begin_failed")
		return
	}
	defer tx.Rollback(r.Context())

	commandTag, err := tx.Exec(
		r.Context(),
		`update vacations set status = 'approved', updated_at = now() where id = $1`,
		vacationID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vacation_approve_failed")
		return
	}
	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "vacation_not_found")
		return
	}

	if _, err := tx.Exec(
		r.Context(),
		`insert into approvals_audit (id, vacation_id, approved_by_member_id, approved_at)
		 values ($1, $2, $3, now())`,
		uuid.New(), vacationID, auth.MemberID,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "audit_write_failed")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "db_commit_failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"vacationId": vacationID.String(), "status": "approved"})
}

func (a *app) issueToken(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "env_admin_only")
		return
	}

	var req issueTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = fmt.Sprintf("member-%s", randomSuffix(6))
	}

	goal := strings.TrimSpace(req.Goal)
	if goal == "" {
		goal = defaultGoal
	}

	colorHex := strings.ToLower(strings.TrimSpace(req.ColorHex))
	if colorHex == "" {
		colorHex = randomColorHex()
	}
	if !colorPattern.MatchString(colorHex) {
		writeError(w, http.StatusBadRequest, "invalid_color")
		return
	}

	result, statusCode, errorCode := a.createMemberConnection(r.Context(), displayName, colorHex, goal, defaultRole)
	if errorCode != "" {
		writeError(w, statusCode, errorCode)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, issueTokenResponse{
		MemberID:    result.MemberID,
		DisplayName: result.DisplayName,
		ColorHex:    result.ColorHex,
		MCPToken:    result.MCPToken,
	})
}

func (a *app) revokeToken(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !auth.IsEnvAdmin {
		writeError(w, http.StatusForbidden, "env_admin_only")
		return
	}

	var req revokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		writeError(w, http.StatusBadRequest, "token_required")
		return
	}
	if token == a.adminSecret {
		writeError(w, http.StatusBadRequest, "cannot_revoke_admin_secret")
		return
	}

	var (
		commandTag pgconn.CommandTag
		err        error
	)
	if tokenID, tokenSecret, ok := parseMemberToken(token); ok {
		commandTag, err = a.db.Exec(
			r.Context(),
			`update connections
			 set active = false, revoked_at = now()
			 where token_id = $1 and token_hash = $2 and active = true`,
			tokenID,
			hashMemberToken(a.tokenPepper, tokenID, tokenSecret),
		)
	} else {
		if !a.legacyTokenFallback {
			writeError(w, http.StatusNotFound, "token_not_found")
			return
		}
		commandTag, err = a.db.Exec(
			r.Context(),
			`update connections
			 set active = false, revoked_at = now()
			 where mcp_token = $1 and active = true`,
			token,
		)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_revoke_failed")
		return
	}
	if commandTag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "token_not_found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func randomSuffix(length int) string {
	if length <= 0 {
		return ""
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "member"
	}
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(out)
}

func randomColorHex() string {
	raw := make([]byte, 3)
	if _, err := rand.Read(raw); err != nil {
		return "#3b82f6"
	}
	return fmt.Sprintf("#%02x%02x%02x", raw[0], raw[1], raw[2])
}

func generateMemberToken(pepper string) (string, string, string, string, error) {
	tokenID, err := randomHex(8)
	if err != nil {
		return "", "", "", "", err
	}
	secret, err := randomURLSafe(32)
	if err != nil {
		return "", "", "", "", err
	}
	token := fmt.Sprintf("%s_%s_%s", memberTokenPrefix, tokenID, secret)
	return token, tokenID, hashMemberToken(pepper, tokenID, secret), memberTokenV1, nil
}

func randomURLSafe(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func randomHex(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func parseMemberToken(token string) (string, string, bool) {
	parts := strings.SplitN(token, "_", 3)
	if len(parts) != 3 {
		return "", "", false
	}
	if parts[0] != memberTokenPrefix || parts[1] == "" || parts[2] == "" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func hashMemberToken(pepper, tokenID, secret string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(tokenID))
	mac.Write([]byte("."))
	mac.Write([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifyMemberTokenHash(pepper, tokenID, secret, storedHash string) bool {
	expected := hashMemberToken(pepper, tokenID, secret)
	return hmac.Equal([]byte(expected), []byte(storedHash))
}

func legacyMCPTokenColumnExists(ctx context.Context, db *pgxpool.Pool) (bool, error) {
	var exists bool
	err := db.QueryRow(
		ctx,
		`select exists (
			select 1
			from information_schema.columns
			where table_schema = 'public'
			  and table_name = 'connections'
			  and column_name = 'mcp_token'
		)`,
	).Scan(&exists)
	return exists, err
}

func (a *app) createMemberConnection(ctx context.Context, displayName, colorHex, goal, role string) (createUserResponse, int, string) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return createUserResponse{}, http.StatusInternalServerError, "db_begin_failed"
	}
	defer tx.Rollback(ctx)

	result, statusCode, errorCode := a.createMemberConnectionInTx(ctx, tx, displayName, colorHex, goal, role)
	if errorCode != "" {
		return createUserResponse{}, statusCode, errorCode
	}

	if err := tx.Commit(ctx); err != nil {
		return createUserResponse{}, http.StatusInternalServerError, "db_commit_failed"
	}

	return result, 0, ""
}

func (a *app) upsertTelegramWhitelist(ctx context.Context, telegramUsername string) (privateTelegramWhitelistResponse, int, string) {
	_, err := a.db.Exec(
		ctx,
		`insert into telegram_whitelist (id, telegram_username, created_at, updated_at)
		 values ($1, $2, now(), now())
		 on conflict (telegram_username)
		 do update set
		   updated_at = now()`,
		uuid.New(), telegramUsername,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return privateTelegramWhitelistResponse{}, http.StatusConflict, "telegram_identity_conflict"
		}
		return privateTelegramWhitelistResponse{}, http.StatusInternalServerError, "telegram_whitelist_upsert_failed"
	}

	return privateTelegramWhitelistResponse{
		TelegramUsername: telegramUsername,
	}, 0, ""
}

func (a *app) exchangeTelegramForToken(ctx context.Context, req privateTelegramExchangeRequest) (privateTelegramExchangeResponse, int, string) {
	if req.TelegramID <= 0 {
		return privateTelegramExchangeResponse{}, http.StatusBadRequest, "telegram_id_required"
	}

	normalizedUsername := normalizeTelegramUsername(req.TelegramUsername)

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "db_begin_failed"
	}
	defer tx.Rollback(ctx)

	record, statusCode, errorCode := loadTelegramWhitelistRecord(ctx, tx, req.TelegramID, normalizedUsername)
	if errorCode != "" {
		return privateTelegramExchangeResponse{}, statusCode, errorCode
	}

	nextUsername := record.TelegramUsername
	if normalizedUsername != "" {
		nextUsername = normalizedUsername
	}
	if nextUsername == "" {
		return privateTelegramExchangeResponse{}, http.StatusForbidden, "telegram_username_required"
	}

	if statusCode, errorCode := updateTelegramIdentityInTx(ctx, tx, record.ID, req.TelegramID, nextUsername, req.FirstName, req.LastName); errorCode != "" {
		return privateTelegramExchangeResponse{}, statusCode, errorCode
	}

	if record.MemberID == nil {
		displayName := nextUsername
		if displayName == "" {
			displayName = fmt.Sprintf("member-%s", randomSuffix(6))
		}

		goal := defaultGoal
		colorHex := randomColorHex()
		if !colorPattern.MatchString(colorHex) {
			return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "invalid_default_color"
		}

		memberResult, statusCode, errorCode := a.createMemberConnectionInTx(ctx, tx, displayName, colorHex, goal, defaultRole)
		if errorCode != "" {
			return privateTelegramExchangeResponse{}, statusCode, errorCode
		}

		memberUUID, err := uuid.Parse(memberResult.MemberID)
		if err != nil {
			return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "member_id_parse_failed"
		}
		if _, err := tx.Exec(
			ctx,
			`update telegram_whitelist set member_id = $1, updated_at = now() where id = $2`,
			memberUUID, record.ID,
		); err != nil {
			return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "telegram_whitelist_link_failed"
		}
		if err := tx.Commit(ctx); err != nil {
			return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "db_commit_failed"
		}
		return privateTelegramExchangeResponse{
			MemberID:         memberResult.MemberID,
			DisplayName:      memberResult.DisplayName,
			ColorHex:         memberResult.ColorHex,
			MCPToken:         memberResult.MCPToken,
			MCPServerURL:     memberResult.MCPServerURL,
			WasCreated:       true,
			TelegramUsername: nextUsername,
		}, 0, ""
	}

	memberResult, statusCode, errorCode := a.rotateMemberTokenInTx(ctx, tx, *record.MemberID, defaultGoal)
	if errorCode != "" {
		return privateTelegramExchangeResponse{}, statusCode, errorCode
	}
	if err := tx.Commit(ctx); err != nil {
		return privateTelegramExchangeResponse{}, http.StatusInternalServerError, "db_commit_failed"
	}
	return privateTelegramExchangeResponse{
		MemberID:         memberResult.MemberID,
		DisplayName:      memberResult.DisplayName,
		ColorHex:         memberResult.ColorHex,
		MCPToken:         memberResult.MCPToken,
		MCPServerURL:     memberResult.MCPServerURL,
		WasCreated:       false,
		TelegramUsername: nextUsername,
	}, 0, ""
}

func (a *app) checkTelegramAccess(ctx context.Context, req privateTelegramExchangeRequest) (privateTelegramAccessResponse, int, string) {
	if req.TelegramID <= 0 {
		return privateTelegramAccessResponse{}, http.StatusBadRequest, "telegram_id_required"
	}

	normalizedUsername := normalizeTelegramUsername(req.TelegramUsername)

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return privateTelegramAccessResponse{}, http.StatusInternalServerError, "db_begin_failed"
	}
	defer tx.Rollback(ctx)

	record, statusCode, errorCode := loadTelegramWhitelistRecord(ctx, tx, req.TelegramID, normalizedUsername)
	if errorCode != "" {
		return privateTelegramAccessResponse{}, statusCode, errorCode
	}

	nextUsername := record.TelegramUsername
	if normalizedUsername != "" {
		nextUsername = normalizedUsername
	}
	if nextUsername == "" {
		return privateTelegramAccessResponse{}, http.StatusForbidden, "telegram_username_required"
	}

	if statusCode, errorCode := updateTelegramIdentityInTx(ctx, tx, record.ID, req.TelegramID, nextUsername, req.FirstName, req.LastName); errorCode != "" {
		return privateTelegramAccessResponse{}, statusCode, errorCode
	}
	if err := tx.Commit(ctx); err != nil {
		return privateTelegramAccessResponse{}, http.StatusInternalServerError, "db_commit_failed"
	}

	return privateTelegramAccessResponse{
		Allowed:          true,
		TelegramUsername: nextUsername,
	}, 0, ""
}

func updateTelegramIdentityInTx(ctx context.Context, tx pgx.Tx, whitelistID uuid.UUID, telegramID int64, telegramUsername, firstName, lastName string) (int, string) {
	if _, err := tx.Exec(
		ctx,
		`update telegram_whitelist
		 set telegram_id = $1,
		     telegram_username = $2,
		     first_name = nullif($3, ''),
		     last_name = nullif($4, ''),
		     last_seen_at = now(),
		     updated_at = now()
		 where id = $5`,
		telegramID,
		telegramUsername,
		strings.TrimSpace(firstName),
		strings.TrimSpace(lastName),
		whitelistID,
	); err != nil {
		if isUniqueViolation(err) {
			return http.StatusConflict, "telegram_identity_conflict"
		}
		return http.StatusInternalServerError, "telegram_whitelist_update_failed"
	}
	return 0, ""
}

func loadTelegramWhitelistRecord(ctx context.Context, tx pgx.Tx, telegramID int64, normalizedUsername string) (telegramWhitelistRecord, int, string) {
	var record telegramWhitelistRecord
	err := tx.QueryRow(
		ctx,
		`select id, member_id, telegram_id, telegram_username
		 from telegram_whitelist
		 where telegram_id = $1`,
		telegramID,
	).Scan(
		&record.ID,
		&record.MemberID,
		&record.TelegramID,
		&record.TelegramUsername,
	)
	if err == nil {
		return record, 0, ""
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return telegramWhitelistRecord{}, http.StatusInternalServerError, "telegram_whitelist_lookup_failed"
	}
	if normalizedUsername == "" {
		return telegramWhitelistRecord{}, http.StatusForbidden, "not_whitelisted"
	}

	err = tx.QueryRow(
		ctx,
		`select id, member_id, telegram_id, telegram_username
		 from telegram_whitelist
		 where telegram_username = $1`,
		normalizedUsername,
	).Scan(
		&record.ID,
		&record.MemberID,
		&record.TelegramID,
		&record.TelegramUsername,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return telegramWhitelistRecord{}, http.StatusForbidden, "not_whitelisted"
		}
		return telegramWhitelistRecord{}, http.StatusInternalServerError, "telegram_whitelist_lookup_failed"
	}
	return record, 0, ""
}

func (a *app) createMemberConnectionInTx(ctx context.Context, tx pgx.Tx, displayName, colorHex, goal, role string) (createUserResponse, int, string) {
	memberID := uuid.New()
	token, tokenID, tokenHash, tokenVersion, err := generateMemberToken(a.tokenPepper)
	if err != nil {
		return createUserResponse{}, http.StatusInternalServerError, "token_generation_failed"
	}

	if _, err := tx.Exec(
		ctx,
		`insert into members (id, display_name, color_hex, role, created_at)
		 values ($1, $2, $3, $4, now())`,
		memberID, displayName, colorHex, role,
	); err != nil {
		return createUserResponse{}, http.StatusInternalServerError, "member_create_failed"
	}

	if _, err := tx.Exec(
		ctx,
		`insert into connections (id, member_id, goal, token_id, token_hash, token_version, active, created_at, revoked_at)
		 values ($1, $2, $3, $4, $5, $6, true, now(), null)`,
		uuid.New(), memberID, goal, tokenID, tokenHash, tokenVersion,
	); err != nil {
		if isUniqueViolation(err) {
			return createUserResponse{}, http.StatusConflict, "duplicate_connection_or_token"
		}
		return createUserResponse{}, http.StatusInternalServerError, "connection_create_failed"
	}

	return createUserResponse{
		MemberID:     memberID.String(),
		DisplayName:  displayName,
		ColorHex:     colorHex,
		Role:         role,
		MCPToken:     token,
		MCPServerURL: os.Getenv("MCP_SERVER_URL"),
	}, 0, ""
}

func (a *app) rotateMemberTokenInTx(ctx context.Context, tx pgx.Tx, memberID uuid.UUID, goal string) (createUserResponse, int, string) {
	token, tokenID, tokenHash, tokenVersion, err := generateMemberToken(a.tokenPepper)
	if err != nil {
		return createUserResponse{}, http.StatusInternalServerError, "token_generation_failed"
	}

	commandTag, err := tx.Exec(
		ctx,
		`update connections
		 set goal = $1,
		     token_id = $2,
		     token_hash = $3,
		     token_version = $4,
		     active = true,
		     revoked_at = null,
		     created_at = now()
		 where member_id = $5`,
		goal, tokenID, tokenHash, tokenVersion, memberID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return createUserResponse{}, http.StatusConflict, "duplicate_connection_or_token"
		}
		return createUserResponse{}, http.StatusInternalServerError, "connection_rotate_failed"
	}
	if commandTag.RowsAffected() == 0 {
		if _, err := tx.Exec(
			ctx,
			`insert into connections (id, member_id, goal, token_id, token_hash, token_version, active, created_at, revoked_at)
			 values ($1, $2, $3, $4, $5, $6, true, now(), null)`,
			uuid.New(), memberID, goal, tokenID, tokenHash, tokenVersion,
		); err != nil {
			if isUniqueViolation(err) {
				return createUserResponse{}, http.StatusConflict, "duplicate_connection_or_token"
			}
			return createUserResponse{}, http.StatusInternalServerError, "connection_create_failed"
		}
	}

	var (
		displayName string
		colorHex    string
		role        string
	)
	if err := tx.QueryRow(
		ctx,
		`select display_name, color_hex, role from members where id = $1`,
		memberID,
	).Scan(&displayName, &colorHex, &role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return createUserResponse{}, http.StatusNotFound, "member_not_found"
		}
		return createUserResponse{}, http.StatusInternalServerError, "member_lookup_failed"
	}

	return createUserResponse{
		MemberID:     memberID.String(),
		DisplayName:  displayName,
		ColorHex:     colorHex,
		Role:         role,
		MCPToken:     token,
		MCPServerURL: os.Getenv("MCP_SERVER_URL"),
	}, 0, ""
}

func parseDateRange(fromRaw, toRaw string) (time.Time, time.Time, bool) {
	fromDate, err := time.Parse("2006-01-02", strings.TrimSpace(fromRaw))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	toDate, err := time.Parse("2006-01-02", strings.TrimSpace(toRaw))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	if toDate.Before(fromDate) {
		return time.Time{}, time.Time{}, false
	}
	return fromDate.UTC(), toDate.UTC(), true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}
