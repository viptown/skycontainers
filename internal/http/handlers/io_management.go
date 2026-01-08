package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"

	"github.com/go-chi/chi/v5"
)

func ShowIOManagement(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceContainers, 0, "입출고관리"); !ok {
		return
	}

	tab := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tab")))
	if tab == "" {
		tab = "inbound"
	}
	if tab != "inbound" && tab != "work" && tab != "outbound" {
		tab = "inbound"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pager := pagination.NewPager(0, page, 20)
	repoItem := repo.Container{}
	list, total, err := repoItem.ListForIOManagement(r.Context(), pager, tab)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pager = pagination.NewPager(total, page, 20)

	view.Render(w, r, "io_management.html", view.PageData{
		Title: "입출고관리",
		Data: map[string]interface{}{
			"Tab":   tab,
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowInboundModal(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceContainers, 0, "입고등록"); !ok {
		return
	}

	typeRepo := repo.ContainerType{}
	types, err := typeRepo.ListAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to load container types", http.StatusInternalServerError)
		return
	}

	view.Render(w, r, "io_inbound_modal.html", view.PageData{
		Data: map[string]interface{}{
			"ContainerID": id,
			"Types":       types,
		},
	})
}

// Re-write PostIOInbound
func PostIOInbound(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		redirectWithError(w, r, "/admin/io_management?tab=inbound", "잘못된 요청입니다.")
		return
	}
	repoItem := repo.Container{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=inbound", "대상 컨테이너를 찾을 수 없습니다.")
		return
	}

	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainers, item.UserID, "입고등록"); !ok {
		return
	}

	// Handle type update if provided
	typeID, _ := strconv.ParseInt(r.FormValue("type_id"), 10, 64)
	if typeID > 0 {
		item.ContainerTypeID = typeID
		if err := item.Update(r.Context()); err != nil {
			redirectWithError(w, r, "/admin/io_management?tab=inbound", "타입 업데이트 실패: "+err.Error())
			return
		}
	}

	if err := repoItem.MarkInboundToday(r.Context(), id); err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=inbound", err.Error())
		return
	}
	redirectWithSuccess(w, r, "/admin/io_management?tab=inbound", "입고등록이 완료되었습니다.")
}

func PostIOProcessing(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		redirectWithError(w, r, "/admin/io_management?tab=work", "잘못된 요청입니다.")
		return
	}
	repoItem := repo.Container{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=work", "대상 컨테이너를 찾을 수 없습니다.")
		return
	}

	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainers, item.UserID, "작업확인"); !ok {
		return
	}

	if err := repoItem.MarkProcessingToday(r.Context(), id); err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=work", err.Error())
		return
	}
	redirectWithSuccess(w, r, "/admin/io_management?tab=work", "작업확인이 완료되었습니다.")
}

func PostIOOutbound(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id <= 0 {
		redirectWithError(w, r, "/admin/io_management?tab=outbound", "잘못된 요청입니다.")
		return
	}
	repoItem := repo.Container{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=outbound", "대상 컨테이너를 찾을 수 없습니다.")
		return
	}

	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainers, item.UserID, "출고등록"); !ok {
		return
	}

	if err := repoItem.MarkOutboundToday(r.Context(), id); err != nil {
		redirectWithError(w, r, "/admin/io_management?tab=outbound", err.Error())
		return
	}
	redirectWithSuccess(w, r, "/admin/io_management?tab=outbound", "출고등록이 완료되었습니다.")
}
