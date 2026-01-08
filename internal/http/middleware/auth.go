package middleware

import (
	"context"
	"net/http"
	"skycontainers/internal/auth"
)

type contextKey string

const UserKey contextKey = "user"

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, _ := auth.Store.Get(r, "session-name")
		user := &auth.User{
			ID:   session.Values["user_id"].(int64),
			Name: session.Values["user_name"].(string),
			Role: session.Values["user_role"].(string),
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
