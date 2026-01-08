package repo

import (
	"context"
	"strings"
	"time"
)

type SupplierPortalItem struct {
	HBLNo             string
	Marks             string
	IsActive          bool
	CreatedAt         time.Time
	BLPositionName    string
	UserName          string
	ContainerNo       string
	ContainerStatus   string
	SupplierName      string
	ContainerTypeName string
	InboundDate       *time.Time
	ProcessingDate    *time.Time
	OutboundDate      *time.Time
}

func (r *SupplierPortalItem) ListByHBLNo(ctx context.Context, hblNo string) ([]SupplierPortalItem, error) {
	number := strings.TrimSpace(hblNo)
	if number == "" {
		return nil, nil
	}

	rows, err := DB.Query(ctx,
		`SELECT b.hbl_no, b.marks, b.is_active, b.created_at,
		        COALESCE(p.name, ''), COALESCE(u.name, ''),
		        c.container_no, c.container_status,
		        c.inbound_date, c.processing_date, c.outbound_date,
		        COALESCE(ct.name, ''), COALESCE(s.name, '')
		   FROM bl_markings b
		   JOIN containers c ON c.id = b.container_id
		   LEFT JOIN container_types ct ON ct.id = c.containers_type_id
		   LEFT JOIN suppliers s ON s.id = c.supplier_id
		   LEFT JOIN bl_positions p ON p.id = b.bl_position_id
		   LEFT JOIN users u ON u.id = b.user_id
		  WHERE LOWER(b.hbl_no) = LOWER($1)
		  ORDER BY b.id DESC`,
		number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SupplierPortalItem
	for rows.Next() {
		var item SupplierPortalItem
		err := rows.Scan(
			&item.HBLNo,
			&item.Marks,
			&item.IsActive,
			&item.CreatedAt,
			&item.BLPositionName,
			&item.UserName,
			&item.ContainerNo,
			&item.ContainerStatus,
			&item.InboundDate,
			&item.ProcessingDate,
			&item.OutboundDate,
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
