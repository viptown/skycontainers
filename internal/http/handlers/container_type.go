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

func ListContainerTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceContainerTypes, 0, "컨테이너 규격 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoCT := repo.ContainerType{}
	list, total, err := repoCT.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)

	view.Render(w, r, "container_types_list.html", view.PageData{
		Title: "컨테이너 규격 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateContainerType(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceContainerTypes, 0, "컨테이너 규격 등록"); !ok {
		return
	}
	view.Render(w, r, "container_types_form.html", view.PageData{
		Title: "컨테이너 규격 등록",
	})
}

func PostCreateContainerType(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceContainerTypes, 0, "컨테이너 규격 등록"); !ok {
		return
	}
	length, _ := strconv.Atoi(r.FormValue("length_ft"))
	softOrder, _ := strconv.Atoi(r.FormValue("soft_order"))

	item := repo.ContainerType{
		Code:      r.FormValue("code"),
		LengthFT:  int16(length),
		Name:      r.FormValue("name"),
		SoftOrder: int32(softOrder),
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "container_types_form.html", view.PageData{
			Title: "컨테이너 규격 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/container_types", "등록이 완료되었습니다.")
}

func ShowEditContainerType(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoCT := repo.ContainerType{}
	item, err := repoCT.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainerTypes, 0, "컨테이너 규격 수정"); !ok {
		return
	}

	view.Render(w, r, "container_types_form.html", view.PageData{
		Title: "컨테이너 규격 수정",
		Data:  item,
	})
}

func PostUpdateContainerType(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainerTypes, 0, "컨테이너 규격 수정"); !ok {
		return
	}
	length, _ := strconv.Atoi(r.FormValue("length_ft"))
	softOrder, _ := strconv.Atoi(r.FormValue("soft_order"))

	item := repo.ContainerType{
		ID:        id,
		Code:      r.FormValue("code"),
		LengthFT:  int16(length),
		Name:      r.FormValue("name"),
		SoftOrder: int32(softOrder),
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "container_types_form.html", view.PageData{
			Title: "컨테이너 규격 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/container_types", "수정이 완료되었습니다.")
}

func DeleteContainerType(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceContainerTypes, 0, "컨테이너 규격 삭제"); !ok {
		return
	}
	repoCT := repo.ContainerType{}
	if err := repoCT.Delete(r.Context(), id); err != nil {
		w.Header().Set("HX-Reswap", "none") // Don't swap if error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// For HTMX delete, we often return empty or just a success indicator
	// but here we might want to tell HTMX to remove the row
	w.WriteHeader(http.StatusOK)
}
