package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type User struct {
	ID            int64
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     string
}

type Verification struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", filepath.Join(dataDir, "ssm.db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("exec %s: %w", pragma, err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email_verified INTEGER NOT NULL DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS vaults (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
		data BLOB NOT NULL,
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS email_verifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		token TEXT UNIQUE NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT DEFAULT (datetime('now'))
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	return nil
}

func CreateUser(email, passwordHash string, verified bool) (int64, error) {
	v := 0
	if verified {
		v = 1
	}
	res, err := db.Exec("INSERT INTO users (email, password_hash, email_verified) VALUES (?, ?, ?)", email, passwordHash, v)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetUserByEmail(email string) (*User, error) {
	u := &User{}
	var verified int
	err := db.QueryRow("SELECT id, email, password_hash, email_verified, created_at FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &verified, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.EmailVerified = verified == 1
	return u, nil
}

func GetUserByID(id int64) (*User, error) {
	u := &User{}
	var verified int
	err := db.QueryRow("SELECT id, email, password_hash, email_verified, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &verified, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.EmailVerified = verified == 1
	return u, nil
}

func SetEmailVerified(userID int64) error {
	_, err := db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", userID)
	return err
}

func GetVault(userID int64) ([]byte, error) {
	var data []byte
	err := db.QueryRow("SELECT data FROM vaults WHERE user_id = ?", userID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func UpsertVault(userID int64, data []byte) error {
	_, err := db.Exec(`
		INSERT INTO vaults (user_id, data, updated_at) VALUES (?, ?, datetime('now'))
		ON CONFLICT(user_id) DO UPDATE SET data = excluded.data, updated_at = datetime('now')
	`, userID, data)
	return err
}

func CreateVerification(userID int64, token string, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO email_verifications (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt.UTC().Format(time.RFC3339))
	return err
}

func GetVerification(token string) (*Verification, error) {
	v := &Verification{}
	var expiresStr string
	err := db.QueryRow("SELECT id, user_id, token, expires_at FROM email_verifications WHERE token = ?", token).
		Scan(&v.ID, &v.UserID, &v.Token, &expiresStr)
	if err != nil {
		return nil, err
	}
	v.ExpiresAt, _ = time.Parse(time.RFC3339, expiresStr)
	return v, nil
}

func DeleteVerificationsByUser(userID int64) error {
	_, err := db.Exec("DELETE FROM email_verifications WHERE user_id = ?", userID)
	return err
}
