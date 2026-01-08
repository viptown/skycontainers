package handlers

import (
	"context"
	"net/http"

	"skycontainers/internal/auth"
	"skycontainers/internal/http/middleware"
	"skycontainers/internal/policy"
	"skycontainers/internal/view"
)

func currentUser(ctx context.Context) (*auth.User, bool) {
	user, ok := ctx.Value(middleware.UserKey).(*auth.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}

func currentUserID(ctx context.Context) (int64, bool) {
	user, ok := currentUser(ctx)
	if !ok {
		return 0, false
	}
	return user.ID, true
}

func requirePermission(w http.ResponseWriter, r *http.Request, action policy.Action, resource policy.Resource, ownerID int64, title string) (*auth.User, bool) {
	user, ok := currentUser(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return nil, false
	}
	if !policy.Allow(user, action, resource, ownerID) {
		renderPermissionDenied(w, r, title)
		return user, false
	}
	return user, true
}

func renderPermissionDenied(w http.ResponseWriter, r *http.Request, title string) {
	if renderModalMessage(w, r, title, "권한이 없습니다.") {
		return
	}
	http.Error(w, "권한이 없습니다.", http.StatusForbidden)
}

func renderModalMessage(w http.ResponseWriter, r *http.Request, title, message string) bool {
	if isModalRequest(r) {
		view.Render(w, r, "modal_message.html", view.PageData{
			Title: title,
			Error: message,
		})
		return true
	}
	return false
}

func isModalRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "global-modal-body"
}
