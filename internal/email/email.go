package email

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"

	mail "github.com/wneessen/go-mail"
)

type Mailer struct {
	host   string
	port   int
	user   string
	pass   string
	apiURL string
}

func NewMailer(host, port, user, pass, apiURL string) *Mailer {
	p, _ := strconv.Atoi(port)
	if p == 0 {
		p = 587
	}
	return &Mailer{host: host, port: p, user: user, pass: pass, apiURL: apiURL}
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Mailer) SendVerification(to, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", m.apiURL, token)

	msg := mail.NewMsg()
	if err := msg.From(m.user); err != nil {
		return err
	}
	if err := msg.To(to); err != nil {
		return err
	}
	msg.Subject("Verify your email")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(
		"Click the link below to verify your email:\n\n%s\n\nThis link expires in 24 hours.",
		verifyURL,
	))

	client, err := mail.NewClient(m.host,
		mail.WithPort(m.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.user),
		mail.WithPassword(m.pass),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	)
	if err != nil {
		return err
	}
	return client.DialAndSend(msg)
}
