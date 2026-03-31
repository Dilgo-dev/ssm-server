package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Dilgo-dev/ssm-sync/internal/auth"
	"github.com/Dilgo-dev/ssm-sync/internal/db"
	"github.com/Dilgo-dev/ssm-sync/internal/email"
)

type RegisterHandler struct {
	Secret      string
	SMTPEnabled bool
	Mailer      *email.Mailer
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if len(req.Password) < 8 {
		jsonError(w, "Password must be at least 8 characters", 400)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	autoVerify := !h.SMTPEnabled
	userID, err := db.CreateUser(req.Email, hash, autoVerify)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			jsonError(w, "Email already registered", 409)
			return
		}
		jsonError(w, "Internal server error", 500)
		return
	}

	if h.SMTPEnabled && h.Mailer != nil {
		token, err := email.GenerateToken()
		if err == nil {
			_ = db.CreateVerification(userID, token, time.Now().Add(24*time.Hour))
			go h.Mailer.SendVerification(req.Email, token)
		}
	}

	tokenStr, err := auth.SignToken(userID, h.Secret)
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}
