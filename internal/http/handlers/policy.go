package handlers

import (
	"net/http"
	"strings"

	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
)

type policyRole struct {
	Key       string
	Label     string
	Summary   string
	Detail    string
	Exception string
	Locked    bool
}

type policyResource struct {
	Key   string
	Label string
}

type policyAction struct {
	Key   string
	Label string
}

func policyRoles() []policyRole {
	return []policyRole{
		{
			Key:       "internal_super_admin",
			Label:     "총관리자",
			Summary:   "모든 권한 보유 (권한관리 포함)",
			Detail:    "총관리자는 모든 메뉴/데이터에 대한 권한을 항상 보유합니다. 권한관리 화면에서 체크를 변경해도 총관리자 권한은 제한되지 않습니다.",
			Exception: "총관리자는 모든 권한이 항상 허용되며, 권한관리 체크 변경의 영향을 받지 않습니다.",
			Locked:    true,
		},
		{
			Key:       "admin",
			Label:     "관리자",
			Summary:   "설정(규격/업체/BL포지션/차량번호) CRUD + 사용자 조회",
			Detail:    "관리자는 설정 메뉴의 주요 마스터 데이터를 등록/수정/삭제할 수 있으며, 사용자 관리는 조회만 가능합니다. 기타 메뉴는 정책에서 개별적으로 허용해야 합니다.",
			Exception: "기본 운영 규칙: 사용자관리는 조회만 허용됩니다.",
		},
		{
			Key:       "staff",
			Label:     "사원",
			Summary:   "입출고/BL마킹/휴가신청서 등록 및 본인만 수정/삭제",
			Detail:    "사원은 입출고/BL 마킹/휴가신청서를 조회/등록할 수 있으며, 수정/삭제는 본인이 등록한 데이터에 한해 허용됩니다.",
			Exception: "사원의 수정/삭제는 정책 체크와 무관하게 본인 데이터만 허용됩니다.",
		},
		{
			Key:       "supplier",
			Label:     "업체 사용자",
			Summary:   "업체 전용 조회만 가능",
			Detail:    "업체 사용자는 내부 관리자 메뉴에 접근할 수 없으며, 업체 전용 조회 페이지에서만 정보를 확인합니다.",
			Exception: "업체 사용자는 전용 조회 페이지 외 기능을 사용하지 않도록 운영합니다.",
		},
	}
}

func policyResources() []policyResource {
	return []policyResource{
		{Key: string(policy.ResourceDashboard), Label: "대시보드"},
                {Key: string(policy.ResourceContainers), Label: "컨테이너관리"},
		{Key: string(policy.ResourceBLMarkings), Label: "BL 마킹"},
		{Key: string(policy.ResourceReports), Label: "휴가신청서"},
		{Key: string(policy.ResourceContainerTypes), Label: "규격관리"},
		{Key: string(policy.ResourceSuppliers), Label: "업체관리"},
		{Key: string(policy.ResourceBLPositions), Label: "BL 포지션"},
		{Key: string(policy.ResourceCarNumbers), Label: "차량번호"},
		{Key: string(policy.ResourceUsers), Label: "사용자관리"},
		{Key: string(policy.ResourceSupplierPortal), Label: "업체 전용 조회"},
		{Key: string(policy.ResourcePolicies), Label: "권한관리"},
	}
}

func policyActions() []policyAction {
	return []policyAction{
		{Key: string(policy.ActionRead), Label: "조회"},
		{Key: string(policy.ActionCreate), Label: "등록"},
		{Key: string(policy.ActionUpdate), Label: "수정"},
		{Key: string(policy.ActionDelete), Label: "삭제"},
	}
}

func policyKey(role, resource, action string) string {
	return role + "::" + resource + "::" + action
}

func ShowPolicySettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourcePolicies, 0, "권한 관리"); !ok {
		return
	}
	repoItem := repo.PolicyPermission{}
	if err := repoItem.EnsureDefaults(r.Context(), policy.DefaultPermissions()); err != nil {
		http.Error(w, "권한 정보를 불러올 수 없습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	list, err := repoItem.List(r.Context())
	if err != nil {
		http.Error(w, "권한 정보를 불러올 수 없습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}
	allowed := make(map[string]bool, len(list))
	for _, item := range list {
		key := policyKey(item.Role, item.Resource, item.Action)
		allowed[key] = item.Allowed
	}

	view.Render(w, r, "policies_list.html", view.PageData{
		Title: "권한 관리",
		Data: map[string]interface{}{
			"Roles":     policyRoles(),
			"Resources": policyResources(),
			"Actions":   policyActions(),
			"Allowed":   allowed,
		},
	})
}

func PostUpdatePolicySettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourcePolicies, 0, "권한 관리"); !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "요청을 처리할 수 없습니다.", http.StatusBadRequest)
		return
	}
	repoItem := repo.PolicyPermission{}
	roles := policyRoles()
	resources := policyResources()
	actions := policyActions()

	for _, role := range roles {
		for _, resource := range resources {
			for _, action := range actions {
				key := policyKey(role.Key, resource.Key, action.Key)
				allowed := strings.EqualFold(r.FormValue(key), "1")
				if role.Locked {
					allowed = true
				}
				if err := repoItem.Upsert(r.Context(), role.Key, resource.Key, action.Key, allowed); err != nil {
					http.Error(w, "권한 저장 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	redirectWithSuccess(w, r, "/admin/policies", "권한이 저장되었습니다.")
}
