package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type PolicyPermission struct {
	Role     string
	Resource string
	Action   string
	Allowed  bool
}

func (r *PolicyPermission) List(ctx context.Context) ([]PolicyPermission, error) {
	rows, err := DB.Query(ctx,
		`SELECT role, resource, action, allowed
		 FROM policy_permissions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PolicyPermission
	for rows.Next() {
		var item PolicyPermission
		if err := rows.Scan(&item.Role, &item.Resource, &item.Action, &item.Allowed); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *PolicyPermission) GetPermission(ctx context.Context, role, resource, action string) (bool, bool, error) {
	var allowed bool
	err := DB.QueryRow(ctx,
		`SELECT allowed
		 FROM policy_permissions
		 WHERE role = $1 AND resource = $2 AND action = $3`,
		role, resource, action).Scan(&allowed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return allowed, true, nil
}

func (r *PolicyPermission) Upsert(ctx context.Context, role, resource, action string, allowed bool) error {
	now := time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO policy_permissions (role, resource, action, allowed, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (role, resource, action)
		 DO UPDATE SET allowed = $4, updated_at = $6`,
		role, resource, action, allowed, now, now)
	return err
}

func (r *PolicyPermission) EnsureDefaults(ctx context.Context, defaults []PolicyPermission) error {
	now := time.Now()
	for _, item := range defaults {
		_, err := DB.Exec(ctx,
			`INSERT INTO policy_permissions (role, resource, action, allowed, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (role, resource, action) DO NOTHING`,
			item.Role, item.Resource, item.Action, item.Allowed, now, now)
		if err != nil {
			return err
		}
	}
	return nil
}
