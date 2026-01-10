package handlers

import (
	"net/http"
	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func ListBLPositions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceBLPositions, 0, "BL 포지션 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.BLPosition{}
	list, total, err := repoItem.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)
	view.Render(w, r, "bl_positions_list.html", view.PageData{
		Title: "BL 포지션 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateBLPosition(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceBLPositions, 0, "BL 포지션 등록"); !ok {
		return
	}
	view.Render(w, r, "bl_positions_form.html", view.PageData{
		Title: "BL 포지션 등록",
	})
}

func PostCreateBLPosition(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceBLPositions, 0, "BL 포지션 등록"); !ok {
		return
	}
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 등록",
			Error: "이름을 입력해 주세요.",
			Data:  repo.BLPosition{Name: name},
		})
		return
	}
	repoItem := repo.BLPosition{}
	exists, err := repoItem.ExistsByName(r.Context(), name, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 등록",
			Error: "같은 이름의 BL 포지션이 이미 있습니다.",
			Data:  repo.BLPosition{Name: name},
		})
		return
	}

	item := repo.BLPosition{
		Name:     name,
		IsActive: true,
		UserID:   userID,
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/bl_positions", "등록이 완료되었습니다.")
}

func ShowEditBLPosition(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.BLPosition{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLPositions, 0, "BL 포지션 수정"); !ok {
		return
	}

	view.Render(w, r, "bl_positions_form.html", view.PageData{
		Title: "BL 포지션 수정",
		Data:  item,
	})
}

func PostUpdateBLPosition(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLPositions, 0, "BL 포지션 수정"); !ok {
		return
	}
	repoItem := repo.BLPosition{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 수정",
			Error: "이름을 입력해 주세요.",
			Data:  repo.BLPosition{ID: id, Name: name, IsActive: existing.IsActive},
		})
		return
	}
	exists, err := repoItem.ExistsByName(r.Context(), name, &id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 수정",
			Error: "같은 이름의 BL 포지션이 이미 있습니다.",
			Data:  repo.BLPosition{ID: id, Name: name, IsActive: existing.IsActive},
		})
		return
	}

	item := repo.BLPosition{
		ID:       id,
		Name:     name,
		IsActive: existing.IsActive,
		UserID:   userID,
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "bl_positions_form.html", view.PageData{
			Title: "BL 포지션 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/bl_positions", "수정이 완료되었습니다.")
}

func PostUpdateBLPositionStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLPositions, 0, "BL 포지션 상태 변경"); !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "요청을 처리할 수 없습니다.", http.StatusBadRequest)
		return
	}
	isActive := strings.EqualFold(r.FormValue("is_active"), "true")
	repoItem := repo.BLPosition{}
	if err := repoItem.UpdateStatus(r.Context(), id, isActive); err != nil {
		http.Error(w, "상태 변경 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	redirectWithSuccess(w, r, "/admin/bl_positions", "상태가 변경되었습니다.")
}

func DeleteBLPosition(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceBLPositions, 0, "BL 포지션 삭제"); !ok {
		return
	}
	repoItem := repo.BLPosition{}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
