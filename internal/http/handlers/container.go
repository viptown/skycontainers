package handlers

import (
	"context"
	"net/http"
	"skycontainers/internal/pagination"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
)

type containerFilters struct {
	ContainerNo        string
	SupplierID         int64
	InboundStartRaw    string
	InboundEndRaw      string
	ProcessingStartRaw string
	ProcessingEndRaw   string
	OutboundStartRaw   string
	OutboundEndRaw     string
	InboundStart       *time.Time
	InboundEnd         *time.Time
	ProcessingStart    *time.Time
	ProcessingEnd      *time.Time
	OutboundStart      *time.Time
	OutboundEnd        *time.Time
	ErrMsg             string
}

func parseContainerFilters(r *http.Request) containerFilters {
	filters := containerFilters{
		ContainerNo:        strings.TrimSpace(r.URL.Query().Get("container_no")),
		InboundStartRaw:    strings.TrimSpace(r.URL.Query().Get("inbound_start")),
		InboundEndRaw:      strings.TrimSpace(r.URL.Query().Get("inbound_end")),
		ProcessingStartRaw: strings.TrimSpace(r.URL.Query().Get("processing_start")),
		ProcessingEndRaw:   strings.TrimSpace(r.URL.Query().Get("processing_end")),
		OutboundStartRaw:   strings.TrimSpace(r.URL.Query().Get("outbound_start")),
		OutboundEndRaw:     strings.TrimSpace(r.URL.Query().Get("outbound_end")),
	}
	filters.SupplierID, _ = strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("supplier_id")), 10, 64)

	var err error
	filters.InboundStart, err = parseOptionalDate(filters.InboundStartRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}
	filters.InboundEnd, err = parseOptionalDate(filters.InboundEndRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}
	filters.ProcessingStart, err = parseOptionalDate(filters.ProcessingStartRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}
	filters.ProcessingEnd, err = parseOptionalDate(filters.ProcessingEndRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}
	filters.OutboundStart, err = parseOptionalDate(filters.OutboundStartRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}
	filters.OutboundEnd, err = parseOptionalDate(filters.OutboundEndRaw)
	if err != nil && filters.ErrMsg == "" {
		filters.ErrMsg = err.Error()
	}

	return filters
}

func ListContainers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
        if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceContainers, 0, "컨테이너 관리"); !ok {
                return
        }

	filters := parseContainerFilters(r)

	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.Container{}
	list, total, err := repoItem.List(r.Context(), pager, filters.ContainerNo, filters.SupplierID, filters.InboundStart, filters.InboundEnd, filters.ProcessingStart, filters.ProcessingEnd, filters.OutboundStart, filters.OutboundEnd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoSup := repo.Supplier{}
	suppliers, err := repoSup.ListAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager = pagination.NewPager(total, page, 10)
        view.Render(w, r, "containers_list.html", view.PageData{
                Title: "컨테이너 관리",
                Error: filters.ErrMsg,
                Data: map[string]interface{}{
			"Items":           list,
			"Pager":           pager,
			"ContainerNo":     filters.ContainerNo,
			"SupplierID":      filters.SupplierID,
			"Suppliers":       suppliers,
			"InboundStart":    filters.InboundStartRaw,
			"InboundEnd":      filters.InboundEndRaw,
			"ProcessingStart": filters.ProcessingStartRaw,
			"ProcessingEnd":   filters.ProcessingEndRaw,
			"OutboundStart":   filters.OutboundStartRaw,
			"OutboundEnd":     filters.OutboundEndRaw,
		},
	})
}

func ShowCreateContainer(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceContainers, 0, "입출고 등록"); !ok {
		return
	}
	item := repo.Container{
		ContainerStatus: "empty",
	}
	data, err := containerFormData(r.Context(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.Render(w, r, "containers_form.html", view.PageData{
		Title: "입출고 등록",
		Data:  data,
	})
}

func PostCreateContainer(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceContainers, 0, "입출고 등록"); !ok {
		return
	}
	containerTypeID, _ := strconv.ParseInt(r.FormValue("containers_type_id"), 10, 64)
	supplierID, _ := strconv.ParseInt(r.FormValue("supplier_id"), 10, 64)
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}

	item := repo.Container{
		ContainerTypeID:       containerTypeID,
		ContainerNo:           r.FormValue("container_no"),
		ContainerStatus:       r.FormValue("container_status"),
		SupplierID:            supplierID,
		BookingNo:             r.FormValue("booking_no"),
		CarNo:                 r.FormValue("car_no"),
		Memo:                  r.FormValue("memo"),
		UserID:                userID,
		ProcessingCancelledBy: userID,
	}

	inboundDate, err := parseOptionalDate(r.FormValue("inbound_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 등록",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	processingDate, err := parseOptionalDate(r.FormValue("processing_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 등록",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	outboundDate, err := parseOptionalDate(r.FormValue("outbound_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 등록",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	cancelledAt, err := parseOptionalDateTime(r.FormValue("processing_cancelled_at"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 등록",
			Error: err.Error(),
			Data:  data,
		})
		return
	}

	item.InboundDate = inboundDate
	item.ProcessingDate = processingDate
	item.OutboundDate = outboundDate
	item.ProcessingCancelledAt = cancelledAt

	if err := item.Create(r.Context()); err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  data,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/containers", "등록이 완료되었습니다.")
}

func ShowEditContainer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Container{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainers, item.UserID, "입출고 수정"); !ok {
		return
	}

	data, err := containerFormData(r.Context(), *item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.Render(w, r, "containers_form.html", view.PageData{
		Title: "입출고 수정",
		Data:  data,
	})
}

func PostUpdateContainer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Container{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceContainers, existing.UserID, "입출고 수정"); !ok {
		return
	}
	containerTypeID, _ := strconv.ParseInt(r.FormValue("containers_type_id"), 10, 64)
	supplierID, _ := strconv.ParseInt(r.FormValue("supplier_id"), 10, 64)

	item := repo.Container{
		ID:                    id,
		ContainerTypeID:       containerTypeID,
		ContainerNo:           r.FormValue("container_no"),
		ContainerStatus:       r.FormValue("container_status"),
		SupplierID:            supplierID,
		BookingNo:             r.FormValue("booking_no"),
		CarNo:                 r.FormValue("car_no"),
		Memo:                  r.FormValue("memo"),
		UserID:                existing.UserID,
		ProcessingCancelledBy: existing.ProcessingCancelledBy,
	}

	inboundDate, err := parseOptionalDate(r.FormValue("inbound_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 수정",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	processingDate, err := parseOptionalDate(r.FormValue("processing_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 수정",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	outboundDate, err := parseOptionalDate(r.FormValue("outbound_date"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 수정",
			Error: err.Error(),
			Data:  data,
		})
		return
	}
	cancelledAt, err := parseOptionalDateTime(r.FormValue("processing_cancelled_at"))
	if err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 수정",
			Error: err.Error(),
			Data:  data,
		})
		return
	}

	item.InboundDate = inboundDate
	item.ProcessingDate = processingDate
	item.OutboundDate = outboundDate
	item.ProcessingCancelledAt = cancelledAt

	if err := item.Update(r.Context()); err != nil {
		data, dataErr := containerFormData(r.Context(), item)
		if dataErr != nil {
			http.Error(w, dataErr.Error(), http.StatusInternalServerError)
			return
		}
		view.Render(w, r, "containers_form.html", view.PageData{
			Title: "입출고 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  data,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/containers", "수정이 완료되었습니다.")
}

func DeleteContainer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.Container{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceContainers, item.UserID, "입출고 삭제"); !ok {
		return
	}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ExportContainers(w http.ResponseWriter, r *http.Request) {
    if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceContainers, 0, "컨테이너 관리"); !ok {
        return
    }
	filters := parseContainerFilters(r)
	if filters.ErrMsg != "" {
		http.Error(w, filters.ErrMsg, http.StatusBadRequest)
		return
	}

	repoItem := repo.Container{}
	list, err := repoItem.ListForExport(r.Context(), filters.ContainerNo, filters.SupplierID, filters.InboundStart, filters.InboundEnd, filters.ProcessingStart, filters.ProcessingEnd, filters.OutboundStart, filters.OutboundEnd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"컨테이너 번호", "컨테이너 타입", "상태", "업체", "입고일", "작업일", "출고일"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	for i, item := range list {
		row := i + 2
		_ = file.SetCellValue(sheet, "A"+strconv.Itoa(row), item.ContainerNo)
		if item.ContainerTypeCode != "" {
			_ = file.SetCellValue(sheet, "B"+strconv.Itoa(row), item.ContainerTypeCode)
		} else {
			_ = file.SetCellValue(sheet, "B"+strconv.Itoa(row), "-")
		}
		_ = file.SetCellValue(sheet, "C"+strconv.Itoa(row), containerStatusLabel(item.ContainerStatus))
		if item.SupplierName != "" {
			_ = file.SetCellValue(sheet, "D"+strconv.Itoa(row), item.SupplierName)
		} else {
			_ = file.SetCellValue(sheet, "D"+strconv.Itoa(row), "-")
		}
		_ = file.SetCellValue(sheet, "E"+strconv.Itoa(row), formatDatePtr(item.InboundDate))
		_ = file.SetCellValue(sheet, "F"+strconv.Itoa(row), formatDatePtr(item.ProcessingDate))
		_ = file.SetCellValue(sheet, "G"+strconv.Itoa(row), formatDatePtr(item.OutboundDate))
	}

	filename := "containers_" + time.Now().Format("20060102_150405") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	if err := file.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func containerStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "empty":
		return "EMPTY"
	case "full":
		return "FULL"
	case "store":
		return "보관"
	default:
		return status
	}
}

func formatDatePtr(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func containerFormData(ctx context.Context, item repo.Container) (map[string]interface{}, error) {
	repoCT := repo.ContainerType{}
	containerTypes, err := repoCT.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	repoSup := repo.Supplier{}
	suppliers, err := repoSup.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"Item":           item,
		"ContainerTypes": containerTypes,
		"Suppliers":      suppliers,
	}, nil
}
