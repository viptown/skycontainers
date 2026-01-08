package repo

import (
	"context"
	"skycontainers/internal/pagination"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID           int64
	SupplierID   *int64
	SupplierName string
	UID          string
	PasswordHash string
	Name         string
	Email        string
	Duty         string
	Phone        string
	Role         string
	Status       string
	LastLoginAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r *User) List(ctx context.Context, p pagination.Pager) ([]User, int, error) {
	var total int
	err := DB.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT u.id, u.uid, u.name, u.role, u.status, u.supplier_id, COALESCE(s.name, ''), u.last_login_at
                FROM users u
                LEFT JOIN suppliers s ON s.id = u.supplier_id
                ORDER BY u.id DESC
                LIMIT $1 OFFSET $2`,
		p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []User
	for rows.Next() {
		var item User
		var supplierID pgtype.Int8
		err := rows.Scan(&item.ID, &item.UID, &item.Name, &item.Role, &item.Status, &supplierID, &item.SupplierName, &item.LastLoginAt)
		if err != nil {
			return nil, 0, err
		}
		if supplierID.Valid {
			value := supplierID.Int64
			item.SupplierID = &value
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *User) GetByID(ctx context.Context, id int64) (*User, error) {
	var item User
	var supplierID pgtype.Int8
	err := DB.QueryRow(ctx,
		`SELECT id, supplier_id, uid, password_hash, name, email, duty, phone, role, status,
                        last_login_at, created_at, updated_at
                FROM users WHERE id = $1`, id).
		Scan(
			&item.ID,
			&supplierID,
			&item.UID,
			&item.PasswordHash,
			&item.Name,
			&item.Email,
			&item.Duty,
			&item.Phone,
			&item.Role,
			&item.Status,
			&item.LastLoginAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	if supplierID.Valid {
		value := supplierID.Int64
		item.SupplierID = &value
	}
	return &item, nil
}

func (r *User) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO users
                 (supplier_id, uid, password_hash, name, email, duty, phone, role, status,
                  last_login_at, created_at, updated_at)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		nullableSupplierID(r.SupplierID),
		r.UID,
		r.PasswordHash,
		r.Name,
		r.Email,
		r.Duty,
		r.Phone,
		r.Role,
		r.Status,
		r.LastLoginAt,
		r.CreatedAt,
		r.UpdatedAt,
	)
	return err
}

func (r *User) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`UPDATE users SET
                 supplier_id = $1,
                 uid = $2,
		 password_hash = $3,
		 name = $4,
		 email = $5,
		 duty = $6,
		 phone = $7,
		 role = $8,
		 status = $9,
		 last_login_at = $10,
                 updated_at = $11
                 WHERE id = $12`,
		nullableSupplierID(r.SupplierID),
		r.UID,
		r.PasswordHash,
		r.Name,
		r.Email,
		r.Duty,
		r.Phone,
		r.Role,
		r.Status,
		r.LastLoginAt,
		r.UpdatedAt,
		r.ID,
	)
	return err
}

func (r *User) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := DB.Exec(ctx,
		"UPDATE users SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id)
	return err
}

func (r *User) UpdateLastLogin(ctx context.Context, id int64, loggedAt time.Time) error {
	_, err := DB.Exec(ctx,
		"UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3",
		loggedAt, loggedAt, id)
	return err
}

func (r *User) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func nullableSupplierID(value *int64) interface{} {
	if value == nil {
		return nil
	}
	return *value
}
