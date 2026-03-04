package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type toolRequest struct {
	Token     string         `json:"token"`
	Arguments map[string]any `json:"arguments"`
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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/tools", toolsList)
	mux.HandleFunc("/tools/createVacation", forward(apiBaseURL, "/v1/mcp/createVacation", http.MethodPost))
	mux.HandleFunc("/tools/changeVacation", forward(apiBaseURL, "/v1/mcp/changeVacation", http.MethodPatch))
	mux.HandleFunc("/tools/removeVacation", forward(apiBaseURL, "/v1/mcp/removeVacation", http.MethodDelete))
	mux.HandleFunc("/tools/approveVacation", forward(apiBaseURL, "/v1/mcp/approveVacation", http.MethodPost))
	mux.HandleFunc("/tools/issueToken", forward(apiBaseURL, "/v1/mcp/issueToken", http.MethodPost))
	mux.HandleFunc("/tools/revokeToken", forward(apiBaseURL, "/v1/mcp/revokeToken", http.MethodPost))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("mcp adapter listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func toolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tools": []map[string]any{
			{
				"name":        "createVacation",
				"description": "Create a vacation request",
				"arguments":   []string{"from", "to", "reason", "displayName(optional)", "colorHex(optional)"},
			},
			{
				"name":        "changeVacation",
				"description": "Change existing vacation by id",
				"arguments":   []string{"vacationId", "newFrom", "newTo", "newReason", "displayName(optional)", "colorHex(optional)"},
			},
			{
				"name":        "removeVacation",
				"description": "Remove existing vacation by id",
				"arguments":   []string{"vacationId", "displayName(optional)", "colorHex(optional)"},
			},
			{
				"name":        "approveVacation",
				"description": "Approve vacation by id, admin only",
				"arguments":   []string{"vacationId"},
			},
			{
				"name":        "issueToken",
				"description": "Issue new member token, env-admin only",
				"arguments":   []string{"displayName(optional)", "colorHex(optional)", "goal(optional)"},
			},
			{
				"name":        "revokeToken",
				"description": "Revoke member token, env-admin only",
				"arguments":   []string{"token"},
			},
		},
	})
}

func forward(apiBaseURL, targetPath, expectedMethod string) http.HandlerFunc {
	client := &http.Client{Timeout: 15 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
			return
		}

		var req toolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		if strings.TrimSpace(req.Token) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing_token"})
			return
		}

		targetURL := apiBaseURL + targetPath
		queryParts := make([]string, 0)
		switch targetPath {
		case "/v1/mcp/changeVacation", "/v1/mcp/removeVacation", "/v1/mcp/approveVacation":
			vacationID, _ := req.Arguments["vacationId"].(string)
			if strings.TrimSpace(vacationID) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "vacationId_required"})
				return
			}
			queryParts = append(queryParts, "vacationId="+url.QueryEscape(vacationID))
		}
		if len(queryParts) > 0 {
			targetURL += "?" + strings.Join(queryParts, "&")
		}

		payload := map[string]any{}
		switch targetPath {
		case "/v1/mcp/createVacation":
			if !requireStringArg(w, req.Arguments, "from") ||
				!requireStringArg(w, req.Arguments, "to") ||
				!requireStringArg(w, req.Arguments, "reason") {
				return
			}
			payload["fromDate"] = req.Arguments["from"]
			payload["toDate"] = req.Arguments["to"]
			payload["reason"] = req.Arguments["reason"]
		case "/v1/mcp/changeVacation":
			if !requireStringArg(w, req.Arguments, "newFrom") ||
				!requireStringArg(w, req.Arguments, "newTo") ||
				!requireStringArg(w, req.Arguments, "newReason") {
				return
			}
			payload["newFrom"] = req.Arguments["newFrom"]
			payload["newTo"] = req.Arguments["newTo"]
			payload["newReason"] = req.Arguments["newReason"]
		case "/v1/mcp/issueToken":
			if value, ok := optionalStringArg(req.Arguments, "displayName"); ok {
				payload["displayName"] = value
			}
			if value, ok := optionalStringArg(req.Arguments, "colorHex"); ok {
				payload["colorHex"] = value
			}
			if value, ok := optionalStringArg(req.Arguments, "goal"); ok {
				payload["goal"] = value
			}
		case "/v1/mcp/revokeToken":
			if !requireStringArg(w, req.Arguments, "token") {
				return
			}
			payload["token"] = req.Arguments["token"]
		}

		body, _ := json.Marshal(payload)
		outReq, err := http.NewRequest(expectedMethod, targetURL, bytes.NewReader(body))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "request_build_failed"})
			return
		}
		outReq.Header.Set("Content-Type", "application/json")
		outReq.Header.Set("Authorization", "Bearer "+req.Token)
		if value, ok := optionalStringArg(req.Arguments, "displayName"); ok {
			outReq.Header.Set("X-Profile-Name", value)
		}
		if value, ok := optionalStringArg(req.Arguments, "colorHex"); ok {
			outReq.Header.Set("X-Profile-Color", strings.ToLower(value))
		}

		response, err := client.Do(outReq)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "api_unreachable"})
			return
		}
		defer response.Body.Close()

		respBody, _ := io.ReadAll(response.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write(respBody)
	}
}

func optionalStringArg(args map[string]any, key string) (string, bool) {
	value, ok := args[key].(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func requireStringArg(w http.ResponseWriter, args map[string]any, key string) bool {
	value, ok := args[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": key + "_required"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
