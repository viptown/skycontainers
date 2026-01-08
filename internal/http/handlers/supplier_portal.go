package handlers

import (
	"net/http"
	"strings"

	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
)

func ShowSupplierPortal(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceSupplierPortal, 0, "업체 전용 조회"); !ok {
		return
	}
	hblNo := strings.TrimSpace(r.URL.Query().Get("hbl_no"))
	var items []repo.SupplierPortalItem
	if hblNo != "" {
		repoItem := repo.SupplierPortalItem{}
		list, err := repoItem.ListByHBLNo(r.Context(), hblNo)
		if err != nil {
			http.Error(w, "조회 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
			return
		}
		items = list
	}
	view.Render(w, r, "supplier_portal.html", view.PageData{
		Title: "업체 전용 조회",
		Data: map[string]interface{}{
			"Items": items,
			"HBLNo": hblNo,
		},
	})
}
