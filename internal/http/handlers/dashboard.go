package handlers

import (
	"net/http"
	"skycontainers/internal/policy"
	"skycontainers/internal/view"
)

func ShowDashboard(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceDashboard, 0, "대시보드"); !ok {
		return
	}
	view.Render(w, r, "dashboard.html", view.PageData{
		Title: "대시보드",
	})
}
