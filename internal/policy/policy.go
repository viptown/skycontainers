package policy

import (
	"context"
	"strings"
	"sync"

	"skycontainers/internal/auth"
	"skycontainers/internal/repo"
)

type Action string

const (
	ActionRead   Action = "read"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Resource string

const (
	ResourceDashboard      Resource = "dashboard"
	ResourceContainers     Resource = "containers"
	ResourceBLMarkings     Resource = "bl_markings"
	ResourceReports        Resource = "reports"
	ResourceContainerTypes Resource = "container_types"
	ResourceSuppliers      Resource = "suppliers"
	ResourceBLPositions    Resource = "bl_positions"
	ResourceCarNumbers     Resource = "carnumbers"
	ResourceUsers          Resource = "users"
	ResourceSupplierPortal Resource = "supplier_portal"
	ResourcePolicies       Resource = "policies"
)

const (
	roleSuperAdmin = "internal_super_admin"
	roleAdmin      = "admin"
	roleStaff      = "staff"
	roleSupplier   = "supplier"
)

var defaultsMu sync.Mutex
var defaultsReady bool

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func IsSuperAdmin(user *auth.User) bool {
	if user == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(user.Role), "INTERNAL_SUPER_ADMIN")
}

func Allow(user *auth.User, action Action, resource Resource, ownerID int64) bool {
	if user == nil {
		return false
	}

	if IsSuperAdmin(user) {
		return true
	}

	role := normalizeRole(user.Role)
	actionKey := Action(strings.ToLower(strings.TrimSpace(string(action))))
	resourceKey := Resource(strings.ToLower(strings.TrimSpace(string(resource))))

	allowed, found, err := lookupPermission(role, resourceKey, actionKey)
	if err != nil {
		return defaultAllow(user, actionKey, resourceKey, ownerID)
	}
	if !found {
		return false
	}
	if !allowed {
		return false
	}
	if role == roleStaff && (actionKey == ActionUpdate || actionKey == ActionDelete) && isOwnerScoped(resourceKey) {
		return ownerID > 0 && ownerID == user.ID
	}
	return true
}

func Actions() []Action {
	return []Action{ActionRead, ActionCreate, ActionUpdate, ActionDelete}
}

func Resources() []Resource {
	return []Resource{
		ResourceDashboard,
		ResourceContainers,
		ResourceBLMarkings,
		ResourceReports,
		ResourceContainerTypes,
		ResourceSuppliers,
		ResourceBLPositions,
		ResourceCarNumbers,
		ResourceUsers,
		ResourceSupplierPortal,
		ResourcePolicies,
	}
}

func DefaultPermissions() []repo.PolicyPermission {
	var defaults []repo.PolicyPermission
	for _, role := range []string{roleSuperAdmin, roleAdmin, roleStaff, roleSupplier} {
		for _, resource := range Resources() {
			for _, action := range Actions() {
				defaults = append(defaults, repo.PolicyPermission{
					Role:     role,
					Resource: string(resource),
					Action:   string(action),
					Allowed:  defaultAllowByRole(role, action, resource),
				})
			}
		}
	}
	return defaults
}

func lookupPermission(role string, resource Resource, action Action) (bool, bool, error) {
	if err := ensureDefaults(); err != nil {
		return false, false, err
	}
	repoItem := repo.PolicyPermission{}
	return repoItem.GetPermission(context.Background(), role, string(resource), string(action))
}

func ensureDefaults() error {
	defaultsMu.Lock()
	defer defaultsMu.Unlock()
	if defaultsReady {
		return nil
	}
	repoItem := repo.PolicyPermission{}
	if err := repoItem.EnsureDefaults(context.Background(), DefaultPermissions()); err != nil {
		return err
	}
	defaultsReady = true
	return nil
}

func defaultAllow(user *auth.User, action Action, resource Resource, ownerID int64) bool {
	role := normalizeRole(user.Role)
	if role == roleSuperAdmin {
		return true
	}
	if role == roleStaff && (action == ActionUpdate || action == ActionDelete) && isOwnerScoped(resource) {
		return ownerID > 0 && ownerID == user.ID
	}
	return defaultAllowByRole(role, action, resource)
}

func defaultAllowByRole(role string, action Action, resource Resource) bool {
	switch role {
	case roleSupplier:
		return resource == ResourceSupplierPortal && action == ActionRead
	case roleAdmin:
		switch resource {
		case ResourceDashboard:
			return action == ActionRead
		case ResourceUsers:
			return action == ActionRead
		case ResourceContainerTypes, ResourceSuppliers, ResourceBLPositions, ResourceCarNumbers:
			return action == ActionRead || action == ActionCreate || action == ActionUpdate || action == ActionDelete
		case ResourcePolicies:
			return false
		default:
			return false
		}
	case roleStaff:
		switch resource {
		case ResourceDashboard:
			return action == ActionRead
		case ResourceContainers, ResourceBLMarkings, ResourceReports:
			return action == ActionRead || action == ActionCreate || action == ActionUpdate || action == ActionDelete
		default:
			return false
		}
	case roleSuperAdmin:
		return true
	default:
		return false
	}
}

func isOwnerScoped(resource Resource) bool {
	switch resource {
	case ResourceContainers, ResourceBLMarkings, ResourceReports:
		return true
	default:
		return false
	}
}
