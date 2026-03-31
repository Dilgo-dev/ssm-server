package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Dilgo-dev/ssm-sync/internal/auth"
	"github.com/Dilgo-dev/ssm-sync/internal/db"
)

type LoginHandler struct {
	Secret      string
	SMTPEnabled bool
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", 400)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		jsonError(w, "Email and password are required", 400)
		return
	}

	user, err := db.GetUserByEmail(req.Email)
	if err == sql.ErrNoRows {
		jsonError(w, "Invalid email or password", 401)
		return
	}
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		jsonError(w, "Invalid email or password", 401)
		return
	}

	if h.SMTPEnabled && !user.EmailVerified {
		jsonError(w, "Email not verified. Check your inbox.", 403)
		return
	}

	token, err := auth.SignToken(user.ID, h.Secret)
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
