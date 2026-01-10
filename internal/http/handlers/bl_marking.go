package handlers

import (
	"context"
	"net/http"
	"net/url"
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

func ListBLMarkings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceBLMarkings, 0, "BL 마킹 관리"); !ok {
		return
	}

	containerNo := strings.TrimSpace(r.URL.Query().Get("container_no"))
	hblNo := strings.TrimSpace(r.URL.Query().Get("hbl_no"))
	unassignedOnly := strings.EqualFold(r.URL.Query().Get("unassigned_only"), "1") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "true") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "on")
	data, err := blMarkingPageData(r.Context(), page, containerNo, hblNo, unassignedOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.Render(w, r, "bl_markings_list.html", view.PageData{
		Title: "BL 마킹 관리",
		Data:  data,
	})
}

func ExportBLMarkings(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceBLMarkings, 0, "BL 마킹 관리"); !ok {
		return
	}

	containerNo := strings.TrimSpace(r.URL.Query().Get("container_no"))
	hblNo := strings.TrimSpace(r.URL.Query().Get("hbl_no"))
	unassignedOnly := strings.EqualFold(r.URL.Query().Get("unassigned_only"), "1") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "true") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "on")

	repoItem := repo.BLMarking{}
	list, err := repoItem.ListForExport(r.Context(), containerNo, hblNo, unassignedOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"컨테이너 번호", "업체", "BL 포지션", "HBL 번호", "Marks", "사용자", "생성일"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	for i, item := range list {
		row := i + 2
		containerValue := item.ContainerNo
		if containerValue == "" {
			containerValue = "-"
		}
		supplierValue := item.SupplierName
		if supplierValue == "" {
			supplierValue = "-"
		}
		positionValue := item.BLPositionName
		if positionValue == "" {
			positionValue = "미지정"
		}
		userValue := item.UserName
		if userValue == "" {
			userValue = "-"
		}
		createdValue := ""
		if !item.CreatedAt.IsZero() {
			createdValue = item.CreatedAt.Format("2006-01-02")
		}
		_ = file.SetCellValue(sheet, "A"+strconv.Itoa(row), containerValue)
		_ = file.SetCellValue(sheet, "B"+strconv.Itoa(row), supplierValue)
		_ = file.SetCellValue(sheet, "C"+strconv.Itoa(row), positionValue)
		_ = file.SetCellValue(sheet, "D"+strconv.Itoa(row), item.HBLNo)
		_ = file.SetCellValue(sheet, "E"+strconv.Itoa(row), item.Marks)
		_ = file.SetCellValue(sheet, "F"+strconv.Itoa(row), userValue)
		_ = file.SetCellValue(sheet, "G"+strconv.Itoa(row), createdValue)
	}

	filename := "bl_markings_" + time.Now().Format("20060102_150405") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	if err := file.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ShowCreateBLMarking(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceBLMarkings, 0, "BL 마킹 등록"); !ok {
		return
	}
	view.Render(w, r, "bl_markings_form.html", view.PageData{
		Title: "BL 마킹 등록",
	})
}

func PostCreateBLMarking(w http.ResponseWriter, r *http.Request) {
	containerID, _ := strconv.ParseInt(r.FormValue("container_id"), 10, 64)
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceBLMarkings, 0, "BL 마킹 등록"); !ok {
		return
	}
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}
	blPositionID, err := parseOptionalInt64(r.FormValue("bl_position_id"))
	if err != nil {
		view.Render(w, r, "bl_markings_form.html", view.PageData{
			Title: "BL 마킹 등록",
			Error: err.Error(),
			Data: repo.BLMarking{
				ContainerID:  containerID,
				UserID:       userID,
				BLPositionID: nil,
				HBLNo:        r.FormValue("hbl_no"),
				Marks:        r.FormValue("marks"),
				IsActive:     r.FormValue("is_active") == "true",
			},
		})
		return
	}
	isActive := r.FormValue("is_active") == "true"

	item := repo.BLMarking{
		ContainerID:  containerID,
		UserID:       userID,
		BLPositionID: blPositionID,
		HBLNo:        r.FormValue("hbl_no"),
		Marks:        r.FormValue("marks"),
		IsActive:     isActive,
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "bl_markings_form.html", view.PageData{
			Title: "BL 마킹 등록",
			Error: "등록 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/bl_markings", "등록이 완료되었습니다.")
}

func ShowEditBLMarking(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.BLMarking{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLMarkings, item.UserID, "BL 마킹 수정"); !ok {
		return
	}

	view.Render(w, r, "bl_markings_form.html", view.PageData{
		Title: "BL 마킹 수정",
		Data:  item,
	})
}

func PostUpdateBLMarking(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.BLMarking{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLMarkings, existing.UserID, "BL 마킹 수정"); !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		view.Render(w, r, "bl_markings_form.html", view.PageData{
			Title: "BL 마킹 수정",
			Error: "요청을 처리할 수 없습니다.",
			Data:  existing,
		})
		return
	}

	containerID := existing.ContainerID
	if values, ok := r.PostForm["container_id"]; ok && len(values) > 0 {
		if strings.TrimSpace(values[0]) != "" {
			parsed, parseErr := strconv.ParseInt(strings.TrimSpace(values[0]), 10, 64)
			if parseErr != nil {
				view.Render(w, r, "bl_markings_form.html", view.PageData{
					Title: "BL 마킹 수정",
					Error: "컨테이너 ID는 숫자여야 합니다.",
					Data:  existing,
				})
				return
			}
			containerID = parsed
		}
	}

	userID := existing.UserID

	blPositionID := existing.BLPositionID
	if values, ok := r.PostForm["bl_position_id"]; ok && len(values) > 0 {
		if strings.TrimSpace(values[0]) != "" {
			parsed, parseErr := parseOptionalInt64(values[0])
			if parseErr != nil {
				view.Render(w, r, "bl_markings_form.html", view.PageData{
					Title: "BL 마킹 수정",
					Error: parseErr.Error(),
					Data:  existing,
				})
				return
			}
			blPositionID = parsed
		}
	}
	isActive := r.FormValue("is_active") == "true"

	item := repo.BLMarking{
		ID:           id,
		ContainerID:  containerID,
		UserID:       userID,
		BLPositionID: blPositionID,
		HBLNo:        r.FormValue("hbl_no"),
		Marks:        r.FormValue("marks"),
		IsActive:     isActive,
	}

	if err := item.Update(r.Context()); err != nil {
		view.Render(w, r, "bl_markings_form.html", view.PageData{
			Title: "BL 마킹 수정",
			Error: "수정 중 오류가 발생했습니다: " + err.Error(),
			Data:  item,
		})
		return
	}

	redirectWithSuccess(w, r, "/admin/bl_markings", "수정이 완료되었습니다.")
}

func PostUpdateBLMarkingStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.BLMarking{}
	existing, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionUpdate, policy.ResourceBLMarkings, existing.UserID, "BL 마킹 수정"); !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "요청을 처리할 수 없습니다.", http.StatusBadRequest)
		return
	}

	existing.IsActive = r.FormValue("is_active") == "true"
	if err := existing.Update(r.Context()); err != nil {
		http.Error(w, "상태 변경 중 오류가 발생했습니다: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redirectWithSuccess(w, r, "/admin/bl_markings", "상태가 변경되었습니다.")
}

func DeleteBLMarking(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	repoItem := repo.BLMarking{}
	item, err := repoItem.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "찾을 수 없는 항목입니다.", http.StatusNotFound)
		return
	}
	if _, ok := requirePermission(w, r, policy.ActionDelete, policy.ResourceBLMarkings, item.UserID, "BL 마킹 삭제"); !ok {
		return
	}
	if err := repoItem.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func blMarkingPageData(ctx context.Context, page int, containerNo string, hblNo string, unassignedOnly bool) (map[string]interface{}, error) {
	pager := pagination.NewPager(0, page, 10)
	repoItem := repo.BLMarking{}
	list, total, err := repoItem.List(ctx, pager, containerNo, hblNo, unassignedOnly)
	if err != nil {
		return nil, err
	}
	pager = pagination.NewPager(total, page, 10)

	return map[string]interface{}{
		"Items":          list,
		"Pager":          pager,
		"ContainerNo":    containerNo,
		"HBLNo":          hblNo,
		"UnassignedOnly": unassignedOnly,
		"ExportURL":      buildBLMarkingExportURL(containerNo, hblNo, unassignedOnly),
		"CargoCardURL":   buildBLMarkingCargoCardURL(containerNo, hblNo, unassignedOnly),
	}, nil
}

func buildBLMarkingExportURL(containerNo string, hblNo string, unassignedOnly bool) string {
	values := url.Values{}
	if strings.TrimSpace(containerNo) != "" {
		values.Set("container_no", strings.TrimSpace(containerNo))
	}
	if strings.TrimSpace(hblNo) != "" {
		values.Set("hbl_no", strings.TrimSpace(hblNo))
	}
	if unassignedOnly {
		values.Set("unassigned_only", "1")
	}
	if len(values) == 0 {
		return "/admin/bl_markings/export"
	}
	return "/admin/bl_markings/export?" + values.Encode()
}

func buildBLMarkingCargoCardURL(containerNo string, hblNo string, unassignedOnly bool) string {
	values := url.Values{}
	if strings.TrimSpace(containerNo) != "" {
		values.Set("container_no", strings.TrimSpace(containerNo))
	}
	if strings.TrimSpace(hblNo) != "" {
		values.Set("hbl_no", strings.TrimSpace(hblNo))
	}
	if unassignedOnly {
		values.Set("unassigned_only", "1")
	}
	if len(values) == 0 {
		return "/admin/bl_markings/cargo_card"
	}
	return "/admin/bl_markings/cargo_card?" + values.Encode()
}
