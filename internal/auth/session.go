package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"
)

var s *securecookie.SecureCookie

const sessionCookieName = "session"
const csrfCookieName = "csrf_token"

func SetupSecureCookie(hashKeyEnv, blockKeyEnv string) {
	hashKey := decodeKey(hashKeyEnv, 32)
	blockKey := decodeKey(blockKeyEnv, 32)
	s = securecookie.New(hashKey, blockKey)
}

func decodeKey(env string, size int) []byte {
	if env == "" {
		return securecookie.GenerateRandomKey(size)
	}
	key, err := base64.StdEncoding.DecodeString(env)
	if err == nil && len(key) > 0 {
		return key
	}
	return []byte(env)
}

type Session struct {
	UserID   int64     `json:"userId"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Expires  time.Time `json:"expires"`
}

func SetSession(w http.ResponseWriter, sess Session) error {
	if s == nil {
		return errors.New("securecookie not initialized")
	}
	encoded, err := s.Encode(sessionCookieName, sess)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("SESSION_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}
	http.SetCookie(w, cookie)
	return nil
}

func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("SESSION_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	// clear csrf cookie too
	csrf := &http.Cookie{
		Name:   csrfCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, csrf)
}

func GetSession(r *http.Request) (Session, error) {
	var sess Session
	if s == nil {
		return sess, errors.New("securecookie not initialized")
	}
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return sess, err
	}
	err = s.Decode(sessionCookieName, c.Value, &sess)
	if err != nil {
		return sess, err
	}
	return sess, nil
}
