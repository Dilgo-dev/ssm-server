package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Dilgo-dev/ssm-sync/internal/db"
)

func SyncGet(w http.ResponseWriter, r *http.Request) {
	data, err := db.GetVault(UserID(r))
	if err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}
	if data == nil {
		jsonError(w, "No vault found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func SyncPut(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB

	data, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "Request body too large (max 10MB)", 413)
		return
	}
	if len(data) == 0 {
		jsonError(w, "Empty body", 400)
		return
	}

	if err := db.UpsertVault(UserID(r), data); err != nil {
		jsonError(w, "Internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
