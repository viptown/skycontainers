package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"time"
)

type CarNumber struct {
	ID        int64
	LogDate   string
	CarNo     string
	CreatedAt time.Time
}

func (r *CarNumber) List(ctx context.Context, p pagination.Pager) ([]CarNumber, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM carnumbers").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT id, log_date, car_no, created_at
		FROM carnumbers
		ORDER BY id DESC
		LIMIT $1 OFFSET $2`,
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []CarNumber
	for rows.Next() {
		var item CarNumber
		err := rows.Scan(&item.ID, &item.LogDate, &item.CarNo, &item.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *CarNumber) GetByID(ctx context.Context, id int64) (*CarNumber, error) {
	var item CarNumber
	err := DB.QueryRow(ctx,
		`SELECT id, log_date, car_no, created_at
		FROM carnumbers WHERE id = $1`, id).
		Scan(&item.ID, &item.LogDate, &item.CarNo, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CarNumber) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO carnumbers
		 (log_date, car_no, created_at)
		 VALUES ($1, $2, $3)`,
		r.LogDate,
		r.CarNo,
		r.CreatedAt,
	)
	return err
}

func (r *CarNumber) Update(ctx context.Context) error {
	_, err := DB.Exec(ctx,
		`UPDATE carnumbers SET
		 log_date = $1,
		 car_no = $2
		 WHERE id = $3`,
		r.LogDate,
		r.CarNo,
		r.ID,
	)
	return err
}

func (r *CarNumber) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM carnumbers WHERE id = $1", id)
	return err
}
