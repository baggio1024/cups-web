package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"cups-web/internal/auth"
	"cups-web/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeJSONStatus(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSONStatus(w, status, map[string]string{"error": msg})
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// LoginHandler handles POST /api/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "missing credentials")
		return
	}

	var user store.User
	err := appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		found, err := store.GetUserByUsername(r.Context(), tx, req.Username)
		if err != nil {
			return err
		}
		user = found
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "login failed")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	sess := auth.Session{UserID: user.ID, Username: user.Username, Role: user.Role}
	if err := auth.SetSession(w, sess); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "session error")
		return
	}
	// set csrf token cookie (readable by JS)
	token := randomToken()
	csrfCookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   os.Getenv("SESSION_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}
	http.SetCookie(w, csrfCookie)
	writeJSON(w, map[string]bool{"ok": true})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	writeJSON(w, map[string]bool{"ok": true})
}

// SessionHandler handles GET /api/session and returns session info if present
func SessionHandler(w http.ResponseWriter, r *http.Request) {
	sess, err := auth.GetSession(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, sess)
}

func CSRFHandler(w http.ResponseWriter, r *http.Request) {
	// Not used: CSRF token is set on login; provide endpoint if needed
	token := randomToken()
	csrfCookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   os.Getenv("SESSION_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}
	http.SetCookie(w, csrfCookie)
	writeJSON(w, map[string]string{"csrfToken": token})
}
