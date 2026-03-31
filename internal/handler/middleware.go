package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Dilgo-dev/ssm-sync/internal/auth"
	"github.com/Dilgo-dev/ssm-sync/internal/db"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserID(r *http.Request) int64 {
	return r.Context().Value(userIDKey).(int64)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			jsonError(w, "Missing or invalid authorization header", 401)
			return
		}

		userID, err := auth.VerifyToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			jsonError(w, "Invalid or expired token", 401)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func VerifiedMiddleware(smtpEnabled bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !smtpEnabled {
			next(w, r)
			return
		}

		user, err := db.GetUserByID(UserID(r))
		if err != nil {
			jsonError(w, "User not found", 404)
			return
		}
		if !user.EmailVerified {
			jsonError(w, "Email not verified. Check your inbox.", 403)
			return
		}
		next(w, r)
	}
}
