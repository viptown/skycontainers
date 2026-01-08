package repo

import (
	"context"
	"errors"
	"fmt"
	"skycontainers/internal/pagination"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Container struct {
	ID                    int64
	ContainerTypeID       int64
	ContainerTypeCode     string
	ContainerTypeName     string
	ContainerNo           string
	ContainerStatus       string
	SupplierID            int64
	SupplierName          string
	BookingNo             string
	CarNo                 string
	Memo                  string
	UserID                int64
	InboundDate           *time.Time
	ProcessingDate        *time.Time
	OutboundDate          *time.Time
	ProcessingCancelledAt *time.Time
	ProcessingCancelledBy int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

var ErrContainerUnavailable = errors.New("등록되지 않았거나 이미 출고된 컨테이너입니다.")

var ErrInvalidIOStage = errors.New("invalid io stage")

func (r *Container) ListForIOManagement(ctx context.Context, p pagination.Pager, stage string) ([]Container, int, error) {
	whereClause := "c.outbound_date IS NULL"
	switch strings.ToLower(strings.TrimSpace(stage)) {
	case "", "inbound":
		whereClause = whereClause + " AND c.inbound_date IS NULL AND c.processing_date IS NULL"
	case "work", "processing":
		whereClause = whereClause + " AND c.inbound_date IS NOT NULL AND c.processing_date IS NULL"
	case "outbound":
		whereClause = whereClause + " AND c.inbound_date IS NOT NULL AND c.processing_date IS NOT NULL"
	default:
		return nil, 0, ErrInvalidIOStage
	}

	var total int
	countQuery := "SELECT count(*) FROM containers c WHERE " + whereClause
	if err := DB.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(ctx,
		`SELECT c.id, c.container_no, c.container_status, c.supplier_id,
                        c.booking_no, c.memo, c.car_no,
                        c.inbound_date, c.processing_date, c.outbound_date,
                        ct.code, ct.name, s.name, c.user_id
                 FROM containers c
                 LEFT JOIN container_types ct ON ct.id = c.containers_type_id
                 LEFT JOIN suppliers s ON s.id = c.supplier_id
                 WHERE `+whereClause+`
                 ORDER BY c.id DESC
                 LIMIT $1 OFFSET $2`, p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		if err := rows.Scan(
			&item.ID,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.SupplierID,
			&item.BookingNo,
			&item.Memo,
			&item.CarNo,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
			&item.ContainerTypeCode,
			&item.ContainerTypeName,
			&item.SupplierName,
			&item.UserID,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, nil
}

func (r *Container) MarkInboundToday(ctx context.Context, id int64) error {
	tag, err := DB.Exec(ctx, `UPDATE containers
                SET inbound_date = CURRENT_DATE, updated_at = NOW()
                WHERE id = $1 AND outbound_date IS NULL AND inbound_date IS NULL AND processing_date IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrContainerUnavailable
	}
	return nil
}

func (r *Container) MarkProcessingToday(ctx context.Context, id int64) error {
	tag, err := DB.Exec(ctx, `UPDATE containers
                SET processing_date = CURRENT_DATE, updated_at = NOW()
                WHERE id = $1 AND outbound_date IS NULL
                  AND inbound_date IS NOT NULL AND processing_date IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrContainerUnavailable
	}
	return nil
}

func (r *Container) MarkOutboundToday(ctx context.Context, id int64) error {
	tag, err := DB.Exec(ctx, `UPDATE containers
                SET outbound_date = CURRENT_DATE, updated_at = NOW()
                WHERE id = $1 AND outbound_date IS NULL
                  AND inbound_date IS NOT NULL AND processing_date IS NOT NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrContainerUnavailable
	}
	return nil
}

func buildContainerFilter(containerNo string, supplierID int64, inboundStart *time.Time, inboundEnd *time.Time, processingStart *time.Time, processingEnd *time.Time, outboundStart *time.Time, outboundEnd *time.Time) (string, []interface{}) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 10)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("c.container_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if supplierID > 0 {
		conditions = append(conditions, fmt.Sprintf("c.supplier_id = $%d", len(args)+1))
		args = append(args, supplierID)
	}
	if inboundStart != nil {
		conditions = append(conditions, fmt.Sprintf("c.inbound_date >= $%d", len(args)+1))
		args = append(args, *inboundStart)
	}
	if inboundEnd != nil {
		conditions = append(conditions, fmt.Sprintf("c.inbound_date <= $%d", len(args)+1))
		args = append(args, *inboundEnd)
	}
	if processingStart != nil {
		conditions = append(conditions, fmt.Sprintf("c.processing_date >= $%d", len(args)+1))
		args = append(args, *processingStart)
	}
	if processingEnd != nil {
		conditions = append(conditions, fmt.Sprintf("c.processing_date <= $%d", len(args)+1))
		args = append(args, *processingEnd)
	}
	if outboundStart != nil {
		conditions = append(conditions, fmt.Sprintf("c.outbound_date >= $%d", len(args)+1))
		args = append(args, *outboundStart)
	}
	if outboundEnd != nil {
		conditions = append(conditions, fmt.Sprintf("c.outbound_date <= $%d", len(args)+1))
		args = append(args, *outboundEnd)
	}
	whereClause := strings.Join(conditions, " AND ")
	return whereClause, args
}

func (r *Container) List(ctx context.Context, p pagination.Pager, containerNo string, supplierID int64, inboundStart *time.Time, inboundEnd *time.Time, processingStart *time.Time, processingEnd *time.Time, outboundStart *time.Time, outboundEnd *time.Time) ([]Container, int, error) {
	whereClause, args := buildContainerFilter(containerNo, supplierID, inboundStart, inboundEnd, processingStart, processingEnd, outboundStart, outboundEnd)

	var total int
	countQuery := "SELECT count(*) FROM containers c WHERE " + whereClause
	err := DB.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	rows, err := DB.Query(ctx,
		fmt.Sprintf(`SELECT c.id, c.container_no, c.container_status, c.supplier_id, c.inbound_date, c.processing_date, c.outbound_date, c.car_no,      
                        ct.code, ct.name, s.name, c.user_id
                 FROM containers c
                 LEFT JOIN container_types ct ON ct.id = c.containers_type_id   
                 LEFT JOIN suppliers s ON s.id = c.supplier_id
                 WHERE %s
		 ORDER BY c.id DESC
		 LIMIT $%d OFFSET $%d`, whereClause, limitIndex, offsetIndex),
		append(args, p.PageSize, p.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		err := rows.Scan(
			&item.ID,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.SupplierID,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
			&item.CarNo,
			&item.ContainerTypeCode,
			&item.ContainerTypeName,
			&item.SupplierName,
			&item.UserID,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *Container) ListForExport(ctx context.Context, containerNo string, supplierID int64, inboundStart *time.Time, inboundEnd *time.Time, processingStart *time.Time, processingEnd *time.Time, outboundStart *time.Time, outboundEnd *time.Time) ([]Container, error) {
	whereClause, args := buildContainerFilter(containerNo, supplierID, inboundStart, inboundEnd, processingStart, processingEnd, outboundStart, outboundEnd)
	rows, err := DB.Query(ctx,
		fmt.Sprintf(`SELECT c.id, c.container_no, c.container_status, c.supplier_id, c.inbound_date, c.processing_date, c.outbound_date, c.car_no,
			ct.code, ct.name, s.name
		 FROM containers c
		 LEFT JOIN container_types ct ON ct.id = c.containers_type_id
		 LEFT JOIN suppliers s ON s.id = c.supplier_id
		 WHERE %s
		 ORDER BY c.id DESC`, whereClause),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		err := rows.Scan(
			&item.ID,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.SupplierID,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
			&item.CarNo,
			&item.ContainerTypeCode,
			&item.ContainerTypeName,
			&item.SupplierName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *Container) ListAll(ctx context.Context) ([]Container, error) {
	rows, err := DB.Query(ctx,
		`SELECT id, container_no
		 FROM containers
		 ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		err := rows.Scan(&item.ID, &item.ContainerNo)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *Container) ListByHBLNo(ctx context.Context, hblNo string) ([]Container, error) {
	number := strings.TrimSpace(hblNo)
	if number == "" {
		return nil, nil
	}

	rows, err := DB.Query(ctx,
		`SELECT DISTINCT c.id, c.container_no, c.container_status, c.supplier_id,
		        c.inbound_date, c.processing_date, c.outbound_date,
		        ct.code, ct.name, s.name
		   FROM bl_markings b
		   JOIN containers c ON c.id = b.container_id
		   LEFT JOIN container_types ct ON ct.id = c.containers_type_id
		   LEFT JOIN suppliers s ON s.id = c.supplier_id
		  WHERE b.hbl_no ILIKE $1
		  ORDER BY c.id DESC`,
		"%"+number+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		err := rows.Scan(
			&item.ID,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.SupplierID,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
			&item.ContainerTypeCode,
			&item.ContainerTypeName,
			&item.SupplierName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *Container) ListAvailableForBLMarking(ctx context.Context) ([]Container, error) {
	rows, err := DB.Query(ctx,
		`SELECT id, container_no
		 FROM containers
		 WHERE outbound_date IS NULL
		 ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Container
	for rows.Next() {
		var item Container
		err := rows.Scan(&item.ID, &item.ContainerNo)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *Container) FindAvailableByNo(ctx context.Context, containerNo string) (*Container, error) {
	number := strings.TrimSpace(containerNo)
	if number == "" {
		return nil, errors.New("컨테이너 번호를 입력해 주세요.")
	}

	var item Container
	err := DB.QueryRow(ctx,
		`SELECT id, container_no
		 FROM containers
		 WHERE outbound_date IS NULL AND lower(container_no) = lower($1)
		 LIMIT 1`, number).
		Scan(&item.ID, &item.ContainerNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrContainerUnavailable
		}
		return nil, err
	}
	return &item, nil
}

func (r *Container) FindAvailableByID(ctx context.Context, id int64) (*Container, error) {
	if id <= 0 {
		return nil, ErrContainerUnavailable
	}

	var item Container
	err := DB.QueryRow(ctx,
		`SELECT id, container_no
		 FROM containers
		 WHERE outbound_date IS NULL AND id = $1
		 LIMIT 1`, id).
		Scan(&item.ID, &item.ContainerNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrContainerUnavailable
		}
		return nil, err
	}
	return &item, nil
}

func (r *Container) GetByID(ctx context.Context, id int64) (*Container, error) {
	var item Container
	err := DB.QueryRow(ctx,
		`SELECT id, containers_type_id, container_no, container_status, supplier_id, booking_no, car_no,
		        memo, user_id, inbound_date, processing_date, outbound_date, processing_cancelled_at,
		        processing_cancelled_by, created_at, updated_at
		FROM containers WHERE id = $1`, id).
		Scan(
			&item.ID,
			&item.ContainerTypeID,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.SupplierID,
			&item.BookingNo,
			&item.CarNo,
			&item.Memo,
			&item.UserID,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
			&item.ProcessingCancelledAt,
			&item.ProcessingCancelledBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Container) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO containers
		 (containers_type_id, container_no, container_status, supplier_id, booking_no, car_no,
		  memo, user_id, inbound_date, processing_date, outbound_date, processing_cancelled_at,
		  processing_cancelled_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6,
		         $7, $8, $9, $10, $11, $12,
		         $13, $14, $15)`,
		r.ContainerTypeID,
		r.ContainerNo,
		r.ContainerStatus,
		r.SupplierID,
		r.BookingNo,
		r.CarNo,
		r.Memo,
		r.UserID,
		r.InboundDate,
		r.ProcessingDate,
		r.OutboundDate,
		r.ProcessingCancelledAt,
		r.ProcessingCancelledBy,
		r.CreatedAt,
		r.UpdatedAt,
	)
	return err
}

func (r *Container) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`UPDATE containers SET
		 containers_type_id = $1,
		 container_no = $2,
		 container_status = $3,
		 supplier_id = $4,
		 booking_no = $5,
		 car_no = $6,
		 memo = $7,
		 user_id = $8,
		 inbound_date = $9,
		 processing_date = $10,
		 outbound_date = $11,
		 processing_cancelled_at = $12,
		 processing_cancelled_by = $13,
		 updated_at = $14
		 WHERE id = $15`,
		r.ContainerTypeID,
		r.ContainerNo,
		r.ContainerStatus,
		r.SupplierID,
		r.BookingNo,
		r.CarNo,
		r.Memo,
		r.UserID,
		r.InboundDate,
		r.ProcessingDate,
		r.OutboundDate,
		r.ProcessingCancelledAt,
		r.ProcessingCancelledBy,
		r.UpdatedAt,
		r.ID,
	)
	return err
}

func (r *Container) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx, "DELETE FROM containers WHERE id = $1", id)
	return err
}
