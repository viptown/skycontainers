package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"skycontainers/internal/auth"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
)

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	if auth.IsAuthenticated(r) {
		session, _ := auth.Store.Get(r, "session-name")
		role, _ := session.Values["user_role"].(string)
		redirectAfterLogin(w, r, &auth.User{Role: role})
		return
	}
	view.Render(w, r, "login.html", view.PageData{Title: "로그인"})
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	password := r.FormValue("password")

	user, err := auth.Authenticate(r.Context(), uid, password)
	if err != nil {
		if errors.Is(err, auth.ErrUserInactive) {
			view.Render(w, r, "login.html", view.PageData{
				Title: "로그인",
				Error: "비활성화된 계정입니다.",
			})
			return
		}
		view.Render(w, r, "login.html", view.PageData{
			Title: "로그인",
			Error: "아이디 또는 비밀번호가 올바르지 않습니다.",
		})
		return
	}

	if err := auth.SetSession(w, r, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoItem := repo.User{}
	if err := repoItem.UpdateLastLogin(r.Context(), user.ID, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectAfterLogin(w, r, user)
}

func PostLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func redirectAfterLogin(w http.ResponseWriter, r *http.Request, user *auth.User) {
	// Detect mobile devices
	userAgent := r.Header.Get("User-Agent")
	isMobile := strings.Contains(strings.ToLower(userAgent), "mobile") ||
		strings.Contains(strings.ToLower(userAgent), "android") ||
		strings.Contains(strings.ToLower(userAgent), "iphone")

	if strings.EqualFold(strings.TrimSpace(user.Role), "supplier") {
		if isMobile {
			// For now, suppliers might not have a mobile page, but let's send them to portal or mobile home
			http.Redirect(w, r, "/supplier/portal", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/supplier/portal", http.StatusSeeOther)
		}
		return
	}

	if isMobile {
		http.Redirect(w, r, "/mobile/scan", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
