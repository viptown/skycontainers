package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"time"
)

type ContainerType struct {
	ID        int64
	Code      string
	LengthFT  int16
	Name      string
	SoftOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *ContainerType) List(ctx context.Context, p pagination.Pager) ([]ContainerType, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM container_types").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		"SELECT id, code, length_ft, name, soft_order, created_at, updated_at FROM container_types ORDER BY soft_order ASC LIMIT $1 OFFSET $2",
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []ContainerType
	for rows.Next() {
		var item ContainerType
		err := rows.Scan(&item.ID, &item.Code, &item.LengthFT, &item.Name, &item.SoftOrder, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *ContainerType) GetByID(ctx context.Context, id int64) (*ContainerType, error) {
	var item ContainerType
	err := DB.QueryRow(ctx, "SELECT id, code, length_ft, name, soft_order, created_at, updated_at FROM container_types WHERE id = $1", id).
		Scan(&item.ID, &item.Code, &item.LengthFT, &item.Name, &item.SoftOrder, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ContainerType) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		"INSERT INTO container_types (code, length_ft, name, soft_order, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		r.Code, r.LengthFT, r.Name, r.SoftOrder, r.CreatedAt, r.UpdatedAt)
	return err
}

func (r *ContainerType) ListAll(ctx context.Context) ([]ContainerType, error) {
	rows, err := DB.Query(ctx,
		"SELECT id, code, length_ft, name, soft_order, created_at, updated_at FROM container_types ORDER BY soft_order ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ContainerType
	for rows.Next() {
		var item ContainerType
		err := rows.Scan(&item.ID, &item.Code, &item.LengthFT, &item.Name, &item.SoftOrder, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *ContainerType) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		"UPDATE container_types SET code = $1, length_ft = $2, name = $3, soft_order = $4, updated_at = $5 WHERE id = $6",
		r.Code, r.LengthFT, r.Name, r.SoftOrder, r.UpdatedAt, r.ID)
	return err
}

func (r *ContainerType) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM container_types WHERE id = $1", id)
	return err
}
