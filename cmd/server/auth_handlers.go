package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "os"

    "cups-web/internal/auth"
)

type loginReq struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func writeJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
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
    envUser := os.Getenv("AUTH_USER")
    envPass := os.Getenv("AUTH_PASS")
    if req.Username == "" || req.Password == "" || req.Username != envUser || req.Password != envPass {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    sess := auth.Session{Username: req.Username}
    _ = auth.SetSession(w, sess)
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
        http.Error(w, "unauthorized", http.StatusUnauthorized)
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
