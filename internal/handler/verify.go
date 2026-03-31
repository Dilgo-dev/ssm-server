package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dilgo-dev/ssm-server/internal/db"
	"github.com/Dilgo-dev/ssm-server/internal/email"
)

type VerifyHandler struct {
	Mailer *email.Mailer
}

func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(400)
		fmt.Fprint(w, htmlPage("Invalid link", "The verification link is invalid."))
		return
	}

	v, err := db.GetVerification(token)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(400)
		fmt.Fprint(w, htmlPage("Invalid link", "This verification link is invalid or has already been used."))
		return
	}
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(500)
		fmt.Fprint(w, htmlPage("Error", "Something went wrong. Please try again."))
		return
	}

	if time.Now().After(v.ExpiresAt) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(400)
		fmt.Fprint(w, htmlPage("Link expired", "This verification link has expired. Please request a new one."))
		return
	}

	_ = db.SetEmailVerified(v.UserID)
	_ = db.DeleteVerificationsByUser(v.UserID)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlPage("Email verified", "Your email has been verified. You can go back to ssm."))
}

func Status(smtpEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !smtpEnabled {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"verified": true})
			return
		}

		user, err := db.GetUserByID(UserID(r))
		if err != nil {
			jsonError(w, "User not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"verified": user.EmailVerified})
	}
}

func (h *VerifyHandler) Resend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", 400)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		jsonError(w, "Email is required", 400)
		return
	}

	user, err := db.GetUserByEmail(req.Email)
	if err != nil || user.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a verification link has been sent."})
		return
	}

	_ = db.DeleteVerificationsByUser(user.ID)

	token, err := email.GenerateToken()
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	_ = db.CreateVerification(user.ID, token, time.Now().Add(24*time.Hour))
	go h.Mailer.SendVerification(user.Email, token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a verification link has been sent."})
}

func htmlPage(title, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>%s</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#09090b;color:#d4d4d8}
.box{text-align:center;max-width:400px;padding:2rem}h1{color:#fff;margin-bottom:0.5rem}p{color:#a1a1aa}</style>
</head><body><div class="box"><h1>%s</h1><p>%s</p></div></body></html>`, title, title, message)
}
