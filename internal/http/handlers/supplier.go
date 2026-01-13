package handlers

import (
	"net/http"
	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func ListSuppliers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceSuppliers, 0, "업체 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoSup := repo.Supplier{}
	list, total, err := repoSup.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)

	view.Render(w, r, "suppliers_list.html", view.PageData{
		Title: "업체 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateSupplier(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceSuppliers, 0, "업체 등록"); !ok {
		return
	}
	view.Render(w, r, "suppliers_form.html", view.PageData{
		Title: "업체 등록",
	})
}

func PostCreateSupplier(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceSuppliers, 0, "업체 등록"); !ok {
		return
	}
	isActive := r.FormValue("is_active") == "true"

		item := repo.Supplier{
			Name:     r.FormValue("name"),
			ShortName: r.FormValue("short_name"),
			Tel:      r.FormValue("tel"),
			Email:    r.FormValue("email"),
			Color:    r.FormValue("color"),
		IsActive: isActive,
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "suppliers_form.html", view.PageData{
			Title: "업체 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/suppliers", "등록이 완료되었습니다.")
}

func ShowEditSupplier(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoSup := repo.Supplier{}
	item, err := repoSup.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceSuppliers, 0, "업체 정보 수정"); !ok {
		return
	}

	view.Render(w, r, "suppliers_form.html", view.PageData{
		Title: "업체 정보 수정",
		Data:  item,
	})
}

func PostUpdateSupplier(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceSuppliers, 0, "업체 정보 수정"); !ok {
		return
	}
	isActive := r.FormValue("is_active") == "true"

		item := repo.Supplier{
			ID:       id,
			Name:     r.FormValue("name"),
			ShortName: r.FormValue("short_name"),
			Tel:      r.FormValue("tel"),
			Email:    r.FormValue("email"),
			Color:    r.FormValue("color"),
		IsActive: isActive,
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "suppliers_form.html", view.PageData{
			Title: "업체 정보 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/suppliers", "수정이 완료되었습니다.")
}

func DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceSuppliers, 0, "업체 삭제"); !ok {
		return
	}
	repoSup := repo.Supplier{}
	if err := repoSup.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
