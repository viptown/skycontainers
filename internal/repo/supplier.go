package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"time"
)

type Supplier struct {
	ID        int64
	Name      string
	ShortName string
	Tel       string
	Email     string
	Color     string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *Supplier) List(ctx context.Context, p pagination.Pager) ([]Supplier, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM suppliers").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		"SELECT id, name, COALESCE(short_name, ''), tel, email, COALESCE(color, ''), is_active, created_at, updated_at FROM suppliers ORDER BY id DESC LIMIT $1 OFFSET $2",
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Supplier
	for rows.Next() {
		var item Supplier
		err := rows.Scan(&item.ID, &item.Name, &item.ShortName, &item.Tel, &item.Email, &item.Color, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *Supplier) GetByID(ctx context.Context, id int64) (*Supplier, error) {
	var item Supplier
	err := DB.QueryRow(ctx, "SELECT id, name, COALESCE(short_name, ''), tel, email, COALESCE(color, ''), is_active, created_at, updated_at FROM suppliers WHERE id = $1", id).
		Scan(&item.ID, &item.Name, &item.ShortName, &item.Tel, &item.Email, &item.Color, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Supplier) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		"INSERT INTO suppliers (name, short_name, tel, email, color, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		r.Name, r.ShortName, r.Tel, r.Email, r.Color, r.IsActive, r.CreatedAt, r.UpdatedAt)
	return err
}

func (r *Supplier) ListAll(ctx context.Context) ([]Supplier, error) {
	rows, err := DB.Query(ctx,
		"SELECT id, name, COALESCE(short_name, ''), tel, email, COALESCE(color, ''), is_active, created_at, updated_at FROM suppliers WHERE is_active = true ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Supplier
	for rows.Next() {
		var item Supplier
		err := rows.Scan(&item.ID, &item.Name, &item.ShortName, &item.Tel, &item.Email, &item.Color, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *Supplier) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		"UPDATE suppliers SET name = $1, short_name = $2, tel = $3, email = $4, color = $5, is_active = $6, updated_at = $7 WHERE id = $8",
		r.Name, r.ShortName, r.Tel, r.Email, r.Color, r.IsActive, r.UpdatedAt, r.ID)
	return err
}

func (r *Supplier) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM suppliers WHERE id = $1", id)
	return err
}
