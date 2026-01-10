package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"strings"
	"time"
)

type BLPosition struct {
	ID        int64
	Name      string
	IsActive  bool
	UserID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *BLPosition) List(ctx context.Context, p pagination.Pager) ([]BLPosition, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM bl_positions").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT id, name, is_active, user_id
                FROM bl_positions
                ORDER BY id DESC
                LIMIT $1 OFFSET $2`,
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []BLPosition
	for rows.Next() {
		var item BLPosition
		err := rows.Scan(&item.ID, &item.Name, &item.IsActive, &item.UserID)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *BLPosition) ListAll(ctx context.Context) ([]BLPosition, error) {
	rows, err := DB.Query(ctx,
		`SELECT id, name
                FROM bl_positions
                WHERE is_active = true
                ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BLPosition
	for rows.Next() {
		var item BLPosition
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *BLPosition) GetByID(ctx context.Context, id int64) (*BLPosition, error) {
	var item BLPosition
	err := DB.QueryRow(ctx,
		`SELECT id, name, is_active, user_id, created_at, updated_at
                FROM bl_positions WHERE id = $1`, id).
		Scan(
			&item.ID,
			&item.Name,
			&item.IsActive,
			&item.UserID,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BLPosition) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	r.IsActive = true
	_, err := DB.Exec(ctx,
		`INSERT INTO bl_positions
                 (name, is_active, created_at, updated_at, user_id)
                 VALUES ($1, $2, $3, $4, $5)`,
		r.Name,
		r.IsActive,
		r.CreatedAt,
		r.UpdatedAt,
		r.UserID,
	)
	return err
}

func (r *BLPosition) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`UPDATE bl_positions SET
                 name = $1,
                 is_active = $2,
                 updated_at = $3,
                 user_id = $4
                 WHERE id = $5`,
		r.Name,
		r.IsActive,
		r.UpdatedAt,
		r.UserID,
		r.ID,
	)
	return err
}

func (r *BLPosition) ExistsByName(ctx context.Context, name string, excludeID *int64) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var count int
	if excludeID != nil {
		err := DB.QueryRow(ctx,
			`SELECT count(*) FROM bl_positions WHERE lower(name) = lower($1) AND id <> $2`,
			name, *excludeID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
	err := DB.QueryRow(ctx,
		`SELECT count(*) FROM bl_positions WHERE lower(name) = lower($1)`,
		name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BLPosition) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM bl_positions WHERE id = $1", id)
	return err
}

func (r *BLPosition) UpdateStatus(ctx context.Context, id int64, isActive bool) error {
	_, err := DB.Exec(ctx,
		"UPDATE bl_positions SET is_active = $1, updated_at = $2 WHERE id = $3",
		isActive, time.Now(), id)
	return err
}
