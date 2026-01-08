package handlers

import (
	"net/http"
	"strconv"

	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"

	"github.com/go-chi/chi/v5"
)

func ListReports(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceReports, 0, "휴가신청서 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.Report{}
	list, total, err := repoItem.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)
	view.Render(w, r, "reports_list.html", view.PageData{
		Title: "휴가신청서 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateReport(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceReports, 0, "휴가신청서 등록"); !ok {
		return
	}
	view.Render(w, r, "reports_form.html", view.PageData{
		Title: "휴가신청서 등록",
	})
}

func PostCreateReport(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceReports, 0, "휴가신청서 등록"); !ok {
		return
	}
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}
	isActive := r.FormValue("is_active") == "true"

	periodStart, err := parseDate(r.FormValue("period_start"))
	if err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 등록",
			Error: err.Error(),
		})
		return
	}
	periodEnd, err := parseDate(r.FormValue("period_end"))
	if err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 등록",
			Error: err.Error(),
		})
		return
	}

	item := repo.Report{
		UserID:      userID,
		Subject:     r.FormValue("subject"),
		Contents:    r.FormValue("contents"),
		Types:       r.FormValue("types"),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		IsActive:    isActive,
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/reports", "등록이 완료되었습니다.")
}

func ShowEditReport(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Report{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		renderReportModalError(w, r, "휴가신청서 수정", "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceReports, item.UserID, "휴가신청서 수정"); !ok {
		return
	}

	view.Render(w, r, "reports_form.html", view.PageData{
		Title: "휴가신청서 수정",
		Data:  item,
	})
}

func ShowReport(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Report{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		renderReportModalError(w, r, "휴가신청서 보기", "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceReports, 0, "휴가신청서 보기"); !ok {
		return
	}

	view.Render(w, r, "reports_view.html", view.PageData{
		Title: "휴가신청서 보기",
		Data:  item,
	})
}

func PostUpdateReport(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Report{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceReports, existing.UserID, "휴가신청서 수정"); !ok {
		return
	}
	userID := existing.UserID
	isActive := r.FormValue("is_active") == "true"

	periodStart, err := parseDate(r.FormValue("period_start"))
	if err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 수정",
			Error: err.Error(),
		})
		return
	}
	periodEnd, err := parseDate(r.FormValue("period_end"))
	if err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 수정",
			Error: err.Error(),
		})
		return
	}

	item := repo.Report{
		ID:          id,
		UserID:      userID,
		Subject:     r.FormValue("subject"),
		Contents:    r.FormValue("contents"),
		Types:       r.FormValue("types"),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		IsActive:    isActive,
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "reports_form.html", view.PageData{
			Title: "휴가신청서 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/reports", "수정이 완료되었습니다.")
}

func DeleteReport(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Report{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceReports, item.UserID, "휴가신청서 삭제"); !ok {
		return
	}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func renderReportModalError(w http.ResponseWriter, r *http.Request, title, message string, statusCode int) {
	if renderModalMessage(w, r, title, message) {
		return
	}
	http.Error(w, message, statusCode)
}
