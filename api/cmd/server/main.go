package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type app struct {
	db          *pgxpool.Pool
	adminSecret string
}

type connectRequest struct {
	DisplayName string `json:"displayName"`
	Goal        string `json:"goal"`
	ColorHex    string `json:"colorHex"`
	AdminSecret string `json:"adminSecret"`
}

type connectResponse struct {
	MemberID    string `json:"memberId"`
	DisplayName string `json:"displayName"`
	ColorHex    string `json:"colorHex"`
	Role        string `json:"role"`
	MCPToken    string `json:"mcpToken"`
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

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

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
		db:          db,
		adminSecret: os.Getenv("ADMIN_SECRET"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.healthz)
	mux.HandleFunc("/v1/connect", handler.connect)
	mux.HandleFunc("/v1/vacations", handler.listVacations)
	mux.HandleFunc("/v1/mcp/createVacation", handler.auth(handler.createVacation))
	mux.HandleFunc("/v1/mcp/changeVacation", handler.auth(handler.changeVacation))
	mux.HandleFunc("/v1/mcp/removeVacation", handler.auth(handler.removeVacation))
	mux.HandleFunc("/v1/mcp/approveVacation", handler.auth(handler.approveVacation))

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-MCP-Token")
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

func (a *app) connect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Goal = strings.TrimSpace(req.Goal)
	req.ColorHex = strings.TrimSpace(req.ColorHex)

	if req.DisplayName == "" || req.Goal == "" || !colorPattern.MatchString(req.ColorHex) {
		writeError(w, http.StatusBadRequest, "invalid_connect_payload")
		return
	}

	role := "member"
	if a.adminSecret != "" && req.AdminSecret != "" && req.AdminSecret == a.adminSecret {
		role = "admin"
	}

	memberID := uuid.New()
	token := uuid.NewString()

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_begin_failed")
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(
		r.Context(),
		`insert into members (id, display_name, role, created_at)
		 values ($1, $2, $3, now())`,
		memberID, req.DisplayName, role,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "member_create_failed")
		return
	}

	if _, err := tx.Exec(
		r.Context(),
		`insert into connections (id, member_id, goal, color_hex, mcp_token, active, created_at)
		 values ($1, $2, $3, $4, $5, true, now())`,
		uuid.New(), memberID, req.Goal, strings.ToLower(req.ColorHex), token,
	); err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "color_taken_or_duplicate_connection")
			return
		}
		writeError(w, http.StatusInternalServerError, "connection_create_failed")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "db_commit_failed")
		return
	}

	writeJSON(w, http.StatusCreated, connectResponse{
		MemberID:    memberID.String(),
		DisplayName: req.DisplayName,
		ColorHex:    strings.ToLower(req.ColorHex),
		Role:        role,
		MCPToken:    token,
		MCPServerURL: os.Getenv("MCP_SERVER_URL"),
	})
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
		`select v.id, v.member_id, m.display_name, c.color_hex, v.from_date, v.to_date, v.reason, v.status
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
			item            vacationDTO
			id              uuid.UUID
			memberID        uuid.UUID
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
	MemberID uuid.UUID
	Role     string
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

		var ctx authContext
		err := a.db.QueryRow(
			r.Context(),
			`select m.id, m.role
			 from connections c
			 join members m on m.id = c.member_id
			 where c.mcp_token = $1 and c.active = true`,
			token,
		).Scan(&ctx.MemberID, &ctx.Role)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}
			writeError(w, http.StatusInternalServerError, "auth_failed")
			return
		}

		next(w, r, ctx)
	}
}

func (a *app) createVacation(w http.ResponseWriter, r *http.Request, auth authContext) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
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
