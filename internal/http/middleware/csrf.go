package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"skycontainers/internal/auth"
	"strings"
)

type csrfKey string

const csrfContextKey csrfKey = "csrf_token"
const csrfSessionKey = "csrf_token"

func CSRFTokenFromContext(r *http.Request) string {
	if token, ok := r.Context().Value(csrfContextKey).(string); ok {
		return token
	}
	return ""
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := auth.Store.Get(r, "session-name")
		token, ok := session.Values[csrfSessionKey].(string)
		if !ok || token == "" {
			token = generateCSRFToken()
			session.Values[csrfSessionKey] = token
			_ = session.Save(r, w)
		}

		if isUnsafeMethod(r.Method) {
			reqToken := r.Header.Get("X-CSRF-Token")
			if reqToken == "" {
				reqToken = r.FormValue("csrf_token")
			}
			if reqToken == "" && r.PostForm != nil {
				reqToken = r.PostForm.Get("csrf_token")
			}
			if reqToken == "" && r.MultipartForm != nil {
				if values := r.MultipartForm.Value["csrf_token"]; len(values) > 0 {
					reqToken = values[0]
				}
			}
			if reqToken == "" && strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
				_ = r.ParseMultipartForm(10 << 20)
				if r.MultipartForm != nil {
					if values := r.MultipartForm.Value["csrf_token"]; len(values) > 0 {
						reqToken = values[0]
					}
				}
			}
			if reqToken == "" || subtle.ConstantTimeCompare([]byte(reqToken), []byte(token)) != 1 {
				http.Error(w, "CSRF 검증에 실패했습니다.", http.StatusForbidden)
				return
			}
		}

		ctx := r.Context()
		ctx = contextWithCSRFToken(ctx, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func generateCSRFToken() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return base64.RawStdEncoding.EncodeToString(buf)
}

func contextWithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfContextKey, token)
}
