package auth

import (
	"context"
	"errors"
	"net/http"
	"os"

	"skycontainers/internal/repo"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var Store *sessions.CookieStore
var ErrUserInactive = errors.New("inactive user")

func InitAuth() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "default-secret-very-weak"
	}
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
}

type User struct {
	ID     int64
	UID    string
	Name   string
	Role   string
	Status string
}

func Authenticate(ctx context.Context, uid, password string) (*User, error) {
	var user User
	var hash string

	err := repo.DB.QueryRow(ctx, "SELECT id, uid, name, role, status, password_hash FROM users WHERE uid = $1", uid).
		Scan(&user.ID, &user.UID, &user.Name, &user.Role, &user.Status, &hash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func SetSession(w http.ResponseWriter, r *http.Request, user *User) error {
	session, _ := Store.Get(r, "session-name")
	session.Values["user_id"] = user.ID
	session.Values["user_name"] = user.Name
	session.Values["user_role"] = user.Role
	return session.Save(r, w)
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, _ := Store.Get(r, "session-name")
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func IsAuthenticated(r *http.Request) bool {
	session, _ := Store.Get(r, "session-name")
	_, ok := session.Values["user_id"]
	return ok
}
