package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Dilgo-dev/ssm-sync/internal/db"
	"github.com/Dilgo-dev/ssm-sync/internal/email"
	"github.com/Dilgo-dev/ssm-sync/internal/handler"
)

var version = "dev"

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := env("PORT", "8080")
	dataDir := env("DATA_DIR", "./data")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := env("SMTP_PORT", "587")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	apiURL := env("API_URL", "http://localhost:"+port)

	smtpEnabled := smtpHost != ""

	if err := db.Init(dataDir); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}

	var mailer *email.Mailer
	if smtpEnabled {
		mailer = email.NewMailer(smtpHost, smtpPort, smtpUser, smtpPass, apiURL)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	mux.Handle("POST /auth/register", &handler.RegisterHandler{
		Secret:      jwtSecret,
		SMTPEnabled: smtpEnabled,
		Mailer:      mailer,
	})
	mux.Handle("POST /auth/login", &handler.LoginHandler{
		Secret:      jwtSecret,
		SMTPEnabled: smtpEnabled,
	})

	verifyHandler := &handler.VerifyHandler{Mailer: mailer}
	mux.HandleFunc("GET /auth/verify", verifyHandler.Verify)
	mux.HandleFunc("GET /auth/status", handler.AuthMiddleware(jwtSecret, handler.Status(smtpEnabled)))
	mux.HandleFunc("POST /auth/resend-verification", verifyHandler.Resend)

	authVerified := func(h http.HandlerFunc) http.HandlerFunc {
		return handler.AuthMiddleware(jwtSecret, handler.VerifiedMiddleware(smtpEnabled, h))
	}
	mux.HandleFunc("GET /sync", authVerified(handler.SyncGet))
	mux.HandleFunc("PUT /sync", authVerified(handler.SyncPut))

	srv := handler.CORSMiddleware(mux)

	if smtpEnabled {
		log.Printf("ssm-sync %s starting on :%s (SMTP enabled)", version, port)
	} else {
		log.Printf("ssm-sync %s starting on :%s (no SMTP, accounts auto-verified)", version, port)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), srv); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
