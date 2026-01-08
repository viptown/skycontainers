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

func ListCarNumbers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceCarNumbers, 0, "차량번호 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.CarNumber{}
	list, total, err := repoItem.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)
	view.Render(w, r, "carnumbers_list.html", view.PageData{
		Title: "차량번호 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateCarNumber(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceCarNumbers, 0, "차량번호 등록"); !ok {
		return
	}
	view.Render(w, r, "carnumbers_form.html", view.PageData{
		Title: "차량번호 등록",
	})
}

func PostCreateCarNumber(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceCarNumbers, 0, "차량번호 등록"); !ok {
		return
	}
	item := repo.CarNumber{
		LogDate: r.FormValue("log_date"),
		CarNo:   r.FormValue("car_no"),
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "carnumbers_form.html", view.PageData{
			Title: "차량번호 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/carnumbers", "등록이 완료되었습니다.")
}

func ShowEditCarNumber(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.CarNumber{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceCarNumbers, 0, "차량번호 수정"); !ok {
		return
	}

	view.Render(w, r, "carnumbers_form.html", view.PageData{
		Title: "차량번호 수정",
		Data:  item,
	})
}

func PostUpdateCarNumber(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceCarNumbers, 0, "차량번호 수정"); !ok {
		return
	}

	item := repo.CarNumber{
		ID:      id,
		LogDate: r.FormValue("log_date"),
		CarNo:   r.FormValue("car_no"),
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "carnumbers_form.html", view.PageData{
			Title: "차량번호 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/carnumbers", "수정이 완료되었습니다.")
}

func DeleteCarNumber(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceCarNumbers, 0, "차량번호 삭제"); !ok {
		return
	}
	repoItem := repo.CarNumber{}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
