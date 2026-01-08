package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"time"
)

type Report struct {
	ID          int64
	UserID      int64
	UserName    string
	Subject     string
	Contents    string
	Types       string
	PeriodStart time.Time
	PeriodEnd   time.Time
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Report) List(ctx context.Context, p pagination.Pager) ([]Report, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM reports").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT r.id, r.user_id, COALESCE(u.name, ''), r.subject, r.types, r.period_start, r.period_end, r.is_active
		FROM reports r
		LEFT JOIN users u ON u.id = r.user_id
		ORDER BY r.id DESC
		LIMIT $1 OFFSET $2`,
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Report
	for rows.Next() {
		var item Report
		err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Subject, &item.Types, &item.PeriodStart, &item.PeriodEnd, &item.IsActive)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *Report) ListByUser(ctx context.Context, p pagination.Pager, userID int64) ([]Report, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM reports WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT r.id, r.user_id, COALESCE(u.name, ''), r.subject, r.types, r.period_start, r.period_end, r.is_active
		FROM reports r
		LEFT JOIN users u ON u.id = r.user_id
		WHERE r.user_id = $3
		ORDER BY r.id DESC
		LIMIT $1 OFFSET $2`,
		p.PageSize, p.Offset(), userID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Report
	for rows.Next() {
		var item Report
		err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Subject, &item.Types, &item.PeriodStart, &item.PeriodEnd, &item.IsActive)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *Report) GetByID(ctx context.Context, id int64) (*Report, error) {
	var item Report
	err := DB.QueryRow(ctx,
		`SELECT r.id, r.user_id, COALESCE(u.name, ''), r.subject, r.contents, r.types, r.period_start, r.period_end, r.is_active
		FROM reports r
		LEFT JOIN users u ON u.id = r.user_id
		WHERE r.id = $1`, id).
		Scan(
			&item.ID,
			&item.UserID,
			&item.UserName,
			&item.Subject,
			&item.Contents,
			&item.Types,
			&item.PeriodStart,
			&item.PeriodEnd,
			&item.IsActive,
		)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Report) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO reports
		 (user_id, subject, contents, types, period_start, period_end, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		r.UserID,
		r.Subject,
		r.Contents,
		r.Types,
		r.PeriodStart,
		r.PeriodEnd,
		r.IsActive,
		r.CreatedAt,
		r.UpdatedAt,
	)
	return err
}

func (r *Report) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`UPDATE reports SET
		 user_id = $1,
		 subject = $2,
		 contents = $3,
		 types = $4,
		 period_start = $5,
		 period_end = $6,
		 is_active = $7,
		 updated_at = $8
		 WHERE id = $9`,
		r.UserID,
		r.Subject,
		r.Contents,
		r.Types,
		r.PeriodStart,
		r.PeriodEnd,
		r.IsActive,
		r.UpdatedAt,
		r.ID,
	)
	return err
}

func (r *Report) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx,
		"UPDATE reports SET is_active = false, updated_at = $1 WHERE id = $2",
		time.Now(), id)
	return err
}
