package repo

import (
	"context"
	"fmt"
	"skycontainers/internal/pagination"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type BLMarking struct {
	ID             int64
	ContainerID    int64
	ContainerNo    string
	UserID         int64
	UserName       string
	BLPositionID   *int64
	BLPositionName string
	HBLNo          string
	Marks          string
	Cnee           string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	SupplierName   string
	SupplierColor  string
	FrmUnipass     *string
	HasUnipass     bool
}

type BLMarkingUnipassTarget struct {
	ID    int64
	HBLNo string
}

func (r *BLMarking) List(ctx context.Context, p pagination.Pager, containerNo string, hblNo string, unassignedOnly bool, unipassStatus string) ([]BLMarking, int, error) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 4)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("c.container_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if strings.TrimSpace(hblNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.hbl_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(hblNo)+"%")
	}
	if unassignedOnly {
		conditions = append(conditions, "(b.bl_position_id IS NULL OR b.bl_position_id = 0)")
	}
	if strings.EqualFold(unipassStatus, "y") {
		conditions = append(conditions, "b.frm_unipass IS NOT NULL")
	}
	if strings.EqualFold(unipassStatus, "n") {
		conditions = append(conditions, "b.frm_unipass IS NULL")
	}
	whereClause := strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT count(*) FROM bl_markings b LEFT JOIN containers c ON c.id = b.container_id WHERE " + whereClause
	err := DB.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
		rows, err := DB.Query(ctx,
			fmt.Sprintf(`SELECT b.id, b.container_id, b.user_id, b.bl_position_id, b.hbl_no, b.marks, b.cnee, b.is_active, b.created_at,
													 c.container_no, p.name, u.name, s.name,
													 CASE WHEN b.frm_unipass IS NULL THEN false ELSE true END
											 FROM bl_markings b
											 LEFT JOIN containers c ON c.id = b.container_id
											 LEFT JOIN suppliers s ON s.id = c.supplier_id
						LEFT JOIN bl_positions p ON p.id = b.bl_position_id
						LEFT JOIN users u ON u.id = b.user_id
			WHERE %s
			ORDER BY b.id DESC
			LIMIT $%d OFFSET $%d`, whereClause, limitIndex, offsetIndex),
		append(args, p.PageSize, p.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []BLMarking
	for rows.Next() {
		var item BLMarking
		var blPositionID pgtype.Int8
		var containerNo pgtype.Text
		var cnee pgtype.Text
		var positionName pgtype.Text
		var userName pgtype.Text
		var supplierName pgtype.Text
		err := rows.Scan(
			&item.ID,
			&item.ContainerID,
			&item.UserID,
			&blPositionID,
			&item.HBLNo,
			&item.Marks,
			&cnee,
			&item.IsActive,
			&item.CreatedAt,
			&containerNo,
			&positionName,
			&userName,
			&supplierName,
			&item.HasUnipass,
		)
		if err != nil {
			return nil, 0, err
		}
		if blPositionID.Valid {
			value := blPositionID.Int64
			item.BLPositionID = &value
		}
		if containerNo.Valid {
			item.ContainerNo = containerNo.String
		}
		if cnee.Valid {
			item.Cnee = cnee.String
		}
		if positionName.Valid {
			item.BLPositionName = positionName.String
		}
		if userName.Valid {
			item.UserName = userName.String
		}
		if supplierName.Valid {
			item.SupplierName = supplierName.String
		}
		list = append(list, item)
	}

	return list, total, nil
}

func (r *BLMarking) ListForExport(ctx context.Context, containerNo string, hblNo string, unassignedOnly bool, unipassStatus string) ([]BLMarking, error) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 3)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("c.container_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if strings.TrimSpace(hblNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.hbl_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(hblNo)+"%")
	}
	if unassignedOnly {
		conditions = append(conditions, "(b.bl_position_id IS NULL OR b.bl_position_id = 0)")
	}
	if strings.EqualFold(unipassStatus, "y") {
		conditions = append(conditions, "b.frm_unipass IS NOT NULL")
	}
	if strings.EqualFold(unipassStatus, "n") {
		conditions = append(conditions, "b.frm_unipass IS NULL")
	}
	whereClause := strings.Join(conditions, " AND ")

	rows, err := DB.Query(ctx,
		fmt.Sprintf(`SELECT b.id, b.container_id, b.user_id, b.bl_position_id, b.hbl_no, b.marks, b.cnee, b.is_active, b.created_at,
											c.container_no, p.name, u.name, s.name
									FROM bl_markings b
									LEFT JOIN containers c ON c.id = b.container_id
									LEFT JOIN suppliers s ON s.id = c.supplier_id
					LEFT JOIN bl_positions p ON p.id = b.bl_position_id
					LEFT JOIN users u ON u.id = b.user_id
					WHERE %s
					ORDER BY b.id DESC`, whereClause),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BLMarking
	for rows.Next() {
		var item BLMarking
		var blPositionID pgtype.Int8
		var containerNo pgtype.Text
		var cnee pgtype.Text
		var positionName pgtype.Text
		var userName pgtype.Text
		var supplierName pgtype.Text
		err := rows.Scan(
			&item.ID,
			&item.ContainerID,
			&item.UserID,
			&blPositionID,
			&item.HBLNo,
			&item.Marks,
			&cnee,
			&item.IsActive,
			&item.CreatedAt,
			&containerNo,
			&positionName,
			&userName,
			&supplierName,
		)
		if err != nil {
			return nil, err
		}
		if blPositionID.Valid {
			value := blPositionID.Int64
			item.BLPositionID = &value
		}
		if containerNo.Valid {
			item.ContainerNo = containerNo.String
		}
		if cnee.Valid {
			item.Cnee = cnee.String
		}
		if positionName.Valid {
			item.BLPositionName = positionName.String
		}
		if userName.Valid {
			item.UserName = userName.String
		}
		if supplierName.Valid {
			item.SupplierName = supplierName.String
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *BLMarking) DeleteByFilters(ctx context.Context, containerNo string, hblNo string, unassignedOnly bool) (int64, error) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 3)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.container_id IN (SELECT id FROM containers WHERE container_no ILIKE $%d)", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if strings.TrimSpace(hblNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.hbl_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(hblNo)+"%")
	}
	if unassignedOnly {
		conditions = append(conditions, "(b.bl_position_id IS NULL OR b.bl_position_id = 0)")
	}
	whereClause := strings.Join(conditions, " AND ")

	result, err := DB.Exec(ctx, "DELETE FROM bl_markings b WHERE "+whereClause, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *BLMarking) ListForCargoCard(ctx context.Context, containerNo string, hblNo string, unassignedOnly bool, unipassStatus string) ([]BLMarking, error) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 3)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("c.container_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if strings.TrimSpace(hblNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.hbl_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(hblNo)+"%")
	}
	if unassignedOnly {
		conditions = append(conditions, "(b.bl_position_id IS NULL OR b.bl_position_id = 0)")
	}
	if strings.EqualFold(unipassStatus, "y") {
		conditions = append(conditions, "b.frm_unipass IS NOT NULL")
	}
	if strings.EqualFold(unipassStatus, "n") {
		conditions = append(conditions, "b.frm_unipass IS NULL")
	}
	whereClause := strings.Join(conditions, " AND ")

		rows, err := DB.Query(ctx,
			fmt.Sprintf(`SELECT b.id, b.container_id, b.user_id, b.bl_position_id, b.hbl_no, b.marks, b.cnee, b.is_active, b.created_at,
																			 c.container_no, p.name, u.name, s.name, s.color, b.frm_unipass
																			 FROM bl_markings b
																			 LEFT JOIN containers c ON c.id = b.container_id
																			 LEFT JOIN suppliers s ON s.id = c.supplier_id
										LEFT JOIN bl_positions p ON p.id = b.bl_position_id
										LEFT JOIN users u ON u.id = b.user_id   
					WHERE %s
					ORDER BY b.id DESC`, whereClause),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BLMarking
	for rows.Next() {
		var item BLMarking
		var blPositionID pgtype.Int8
		var containerNo pgtype.Text
		var cnee pgtype.Text
		var positionName pgtype.Text
		var userName pgtype.Text
		var supplierName pgtype.Text
		var supplierColor pgtype.Text
		var frmUnipass pgtype.Text
		err := rows.Scan(
			&item.ID,
			&item.ContainerID,
			&item.UserID,
			&blPositionID,
			&item.HBLNo,
			&item.Marks,
			&cnee,
			&item.IsActive,
			&item.CreatedAt,
			&containerNo,
			&positionName,
			&userName,
			&supplierName,
			&supplierColor,
			&frmUnipass,
		)
		if err != nil {
			return nil, err
		}
		if blPositionID.Valid {
			value := blPositionID.Int64
			item.BLPositionID = &value
		}
		if containerNo.Valid {
			item.ContainerNo = containerNo.String
		}
		if cnee.Valid {
			item.Cnee = cnee.String
		}
		if positionName.Valid {
			item.BLPositionName = positionName.String
		}
		if userName.Valid {
			item.UserName = userName.String
		}
		if supplierName.Valid {
			item.SupplierName = supplierName.String
		}
		if supplierColor.Valid {
			item.SupplierColor = supplierColor.String
		}
		if frmUnipass.Valid {
			value := frmUnipass.String
			item.FrmUnipass = &value
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *BLMarking) ListForUnipassApply(ctx context.Context, containerNo string, hblNo string, unassignedOnly bool, unipassStatus string) ([]BLMarkingUnipassTarget, error) {
	conditions := []string{"1=1"}
	args := make([]interface{}, 0, 3)
	if strings.TrimSpace(containerNo) != "" {
		conditions = append(conditions, fmt.Sprintf("c.container_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(containerNo)+"%")
	}
	if strings.TrimSpace(hblNo) != "" {
		conditions = append(conditions, fmt.Sprintf("b.hbl_no ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(hblNo)+"%")
	}
	if unassignedOnly {
		conditions = append(conditions, "(b.bl_position_id IS NULL OR b.bl_position_id = 0)")
	}
	if strings.EqualFold(unipassStatus, "y") {
		conditions = append(conditions, "b.frm_unipass IS NOT NULL")
	}
	if strings.EqualFold(unipassStatus, "n") {
		conditions = append(conditions, "b.frm_unipass IS NULL")
	}
	whereClause := strings.Join(conditions, " AND ")

	rows, err := DB.Query(ctx,
		fmt.Sprintf(`SELECT b.id, b.hbl_no
					FROM bl_markings b
					LEFT JOIN containers c ON c.id = b.container_id
					WHERE %s
					ORDER BY b.id DESC`, whereClause),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BLMarkingUnipassTarget
	for rows.Next() {
		var item BLMarkingUnipassTarget
		if err := rows.Scan(&item.ID, &item.HBLNo); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func (r *BLMarking) UpdateUnipassXML(ctx context.Context, id int64, xmlData *string) error {
	_, err := DB.Exec(ctx,
		`UPDATE bl_markings SET
		 frm_unipass = $1,
		 updated_at = $2
		 WHERE id = $3`,
		xmlData,
		time.Now(),
		id,
	)
	return err
}

func (r *BLMarking) GetByID(ctx context.Context, id int64) (*BLMarking, error) {
	var item BLMarking
	var blPositionID pgtype.Int8
	var containerNo pgtype.Text
	var supplierName pgtype.Text
	err := DB.QueryRow(ctx,
	`SELECT b.id, b.container_id, b.user_id, b.bl_position_id, b.hbl_no, b.marks, b.cnee, b.is_active, b.created_at, b.updated_at,
						c.container_no, s.name
				FROM bl_markings b
				LEFT JOIN containers c ON c.id = b.container_id
		LEFT JOIN suppliers s ON s.id = c.supplier_id
		WHERE b.id = $1`, id).
		Scan(
			&item.ID,
		&item.ContainerID,
		&item.UserID,
		&blPositionID,
		&item.HBLNo,
		&item.Marks,
		&item.Cnee,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
			&containerNo,
			&supplierName,
		)
	if err != nil {
		return nil, err
	}
	if blPositionID.Valid {
		value := blPositionID.Int64
		item.BLPositionID = &value
	}
	if containerNo.Valid {
		item.ContainerNo = containerNo.String
	}
	if supplierName.Valid {
		item.SupplierName = supplierName.String
	}
	return &item, nil
}

func (r *BLMarking) GetByHBLNo(ctx context.Context, hblNo string) (*BLMarking, error) {
	var item BLMarking
	var blPositionID pgtype.Int8
	var containerNo pgtype.Text
	var supplierName pgtype.Text
	var positionName pgtype.Text
	err := DB.QueryRow(ctx,
	`SELECT b.id, b.container_id, b.user_id, b.bl_position_id, b.hbl_no, b.marks, b.cnee, b.is_active, b.created_at, b.updated_at,
						c.container_no, s.name, p.name
				FROM bl_markings b
				LEFT JOIN containers c ON c.id = b.container_id
		LEFT JOIN suppliers s ON s.id = c.supplier_id
		LEFT JOIN bl_positions p ON p.id = b.bl_position_id
		WHERE b.hbl_no = $1 AND b.is_active = true`, hblNo).
		Scan(
			&item.ID,
		&item.ContainerID,
		&item.UserID,
		&blPositionID,
		&item.HBLNo,
		&item.Marks,
		&item.Cnee,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
			&containerNo,
			&supplierName,
			&positionName,
		)
	if err != nil {
		return nil, err
	}
	if blPositionID.Valid {
		value := blPositionID.Int64
		item.BLPositionID = &value
	}
	if containerNo.Valid {
		item.ContainerNo = containerNo.String
	}
	if supplierName.Valid {
		item.SupplierName = supplierName.String
	}
	if positionName.Valid {
		item.BLPositionName = positionName.String
	}
	return &item, nil
}

func (r *BLMarking) Create(ctx context.Context) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`INSERT INTO bl_markings
		 (container_id, user_id, bl_position_id, hbl_no, marks, cnee, is_active, frm_unipass, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		r.ContainerID,
		r.UserID,
		r.BLPositionID,
		r.HBLNo,
		r.Marks,
		r.Cnee,
		r.IsActive,
		r.FrmUnipass,
		r.CreatedAt,
		r.UpdatedAt,
	)
	return err
}

func (r *BLMarking) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx,
		`UPDATE bl_markings SET
		 container_id = $1,
		 user_id = $2,
		 bl_position_id = $3,
		 hbl_no = $4,
		 marks = $5,
		 cnee = $6,
		 is_active = $7,
		 updated_at = $8
		 WHERE id = $9`,
		r.ContainerID,
		r.UserID,
		r.BLPositionID,
		r.HBLNo,
		r.Marks,
		r.Cnee,
		r.IsActive,
		r.UpdatedAt,
		r.ID,
	)
	return err
}

func (r *BLMarking) UpdatePosition(ctx context.Context, id int64, positionID int64) error {
	_, err := DB.Exec(ctx,
		`UPDATE bl_markings SET
		 bl_position_id = $1,
		 updated_at = $2
		 WHERE id = $3`,
		positionID,
		time.Now(),
		id,
	)
	return err
}

func (r *BLMarking) Delete(ctx context.Context, id int64) error {
	_, err := DB.Exec(ctx,
		"UPDATE bl_markings SET is_active = false, updated_at = $1 WHERE id = $2",
		time.Now(), id)
	return err
}
