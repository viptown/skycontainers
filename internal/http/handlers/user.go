package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceUsers, 0, "사용자 관리"); !ok {
		return
	}

	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.User{}
	list, total, err := repoItem.List(r.Context(), pager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)
	view.Render(w, r, "users_list.html", view.PageData{
		Title: "사용자 관리",
		Data: map[string]interface{}{
			"Items": list,
			"Pager": pager,
		},
	})
}

func ShowCreateUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceUsers, 0, "사용자 등록"); !ok {
		return
	}
	renderUserForm(w, r, "사용자 등록", "", repo.User{})
}

func PostCreateUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceUsers, 0, "사용자 등록"); !ok {
		return
	}
	supplierID, err := parseOptionalInt64(r.FormValue("supplier_id"))
	if err != nil {
		renderUserForm(w, r, "사용자 등록", err.Error(), userFromForm(r, supplierID, "", time.Time{}))
		return
	}

	password := r.FormValue("password")
	if password == "" {
		renderUserForm(w, r, "사용자 등록", "비밀번호를 입력해 주세요.", userFromForm(r, supplierID, "", time.Time{}))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderUserForm(w, r, "사용자 등록", "비밀번호 처리 중 오류가 발생했습니다.", userFromForm(r, supplierID, "", time.Time{}))
		return
	}

	item := userFromForm(r, supplierID, string(hash), time.Time{})

	if err := item.Create(r.Context()); err != nil {
		renderUserForm(w, r, "사용자 등록", "등록 중 오류가 발생했습니다: "+err.Error(), item)
		return
	}

	redirectWithSuccess(w, r, "/admin/users", "등록이 완료되었습니다.")
}

func ShowEditUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.User{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceUsers, 0, "사용자 수정"); !ok {
		return
	}

	renderUserForm(w, r, "사용자 수정", "", *item)
}

func PostUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceUsers, 0, "사용자 수정"); !ok {
		return
	}
	supplierID, err := parseOptionalInt64(r.FormValue("supplier_id"))
	if err != nil {
		item := userFromForm(r, supplierID, "", time.Time{})
		item.ID = id
		renderUserForm(w, r, "사용자 수정", err.Error(), item)
		return
	}

	repoItem := repo.User{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}

	password := r.FormValue("password")
	hash := existing.PasswordHash
	if password != "" {
		newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			item := userFromForm(r, supplierID, hash, existing.LastLoginAt)
			item.ID = id
			renderUserForm(w, r, "사용자 수정", "비밀번호 처리 중 오류가 발생했습니다.", item)
			return
		}
		hash = string(newHash)
	}

	item := userFromForm(r, supplierID, hash, existing.LastLoginAt)
	item.ID = id

	if err := item.Update(r.Context()); err != nil {
		renderUserForm(w, r, "사용자 수정", "수정 중 오류가 발생했습니다: "+err.Error(), item)
		return
	}

	redirectWithSuccess(w, r, "/admin/users", "수정이 완료되었습니다.")
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceUsers, 0, "사용자 삭제"); !ok {
		return
	}
	repoItem := repo.User{}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PostUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceUsers, 0, "사용자 상태 변경"); !ok {
		return
	}
	status := "suspened"
	if strings.EqualFold(r.FormValue("is_active"), "true") {
		status = "active"
	}
	repoItem := repo.User{}
	if err := repoItem.UpdateStatus(r.Context(), id, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		label := "비활성"
		className := "badge badge-danger"
		if status == "active" {
			label = "활성"
			className = "badge badge-success"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fmt.Sprintf(`<td id="user-status-%d"><span class="%s">%s</span></td>`, id, className, label)))
		return
	}
	redirectWithSuccess(w, r, "/admin/users", "상태가 변경되었습니다.")
}

func renderUserForm(w http.ResponseWriter, r *http.Request, title, errMsg string, user repo.User) {
	data, err := userFormData(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.Render(w, r, "users_form.html", view.PageData{
		Title: title,
		Error: errMsg,
		Data:  data,
	})
}

func userFormData(ctx context.Context, user repo.User) (map[string]interface{}, error) {
	repoSup := repo.Supplier{}
	suppliers, err := repoSup.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"User":      user,
		"Suppliers": suppliers,
	}, nil
}

func userFromForm(r *http.Request, supplierID *int64, hash string, lastLoginAt time.Time) repo.User {
	return repo.User{
		SupplierID:   supplierID,
		UID:          r.FormValue("uid"),
		PasswordHash: hash,
		Name:         r.FormValue("name"),
		Email:        r.FormValue("email"),
		Duty:         r.FormValue("duty"),
		Phone:        r.FormValue("phone"),
		Role:         r.FormValue("role"),
		Status:       r.FormValue("status"),
		LastLoginAt:  lastLoginAt,
	}
}
