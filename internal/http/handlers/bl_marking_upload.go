package handlers

import (
	"net/http"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const maxUploadRows = 2000

func PostUploadBLMarkings(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionCreate, policy.ResourceBLMarkings, 0, "BL 마킹 관리"); !ok {
		return
	}
	userID, ok := currentUserID(r.Context())
	if !ok {
		http.Error(w, "로그인 사용자 정보를 찾을 수 없습니다.", http.StatusUnauthorized)
		return
	}

	containerNo := strings.TrimSpace(r.FormValue("container_no"))
	if containerNo == "" {
		containerNo = strings.TrimSpace(r.FormValue("container_no_value"))
	}
	containerIDValue, err := parseOptionalInt64(r.FormValue("container_id"))
	if err != nil {
		renderBLMarkingsListError(w, r, err.Error())
		return
	}
	blPositionID, err := parseOptionalInt64(r.FormValue("bl_position_id"))
	if err != nil {
		renderBLMarkingsListError(w, r, err.Error())
		return
	}

	repoContainer := repo.Container{}
	var containerID int64
	if containerIDValue != nil && *containerIDValue > 0 {
		container, err := repoContainer.FindAvailableByID(r.Context(), *containerIDValue)
		if err != nil {
			renderBLMarkingsListError(w, r, err.Error())
			return
		}
		containerID = container.ID
	} else {
		container, err := repoContainer.FindAvailableByNo(r.Context(), containerNo)
		if err != nil {
			renderBLMarkingsListError(w, r, err.Error())
			return
		}
		containerID = container.ID
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		renderBLMarkingsListError(w, r, "업로드 파일을 확인해 주세요.")
		return
	}
	defer file.Close()

	if header == nil || !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		renderBLMarkingsListError(w, r, "xlsx 파일만 업로드할 수 있습니다.")
		return
	}

	workbook, err := excelize.OpenReader(file)
	if err != nil {
		renderBLMarkingsListError(w, r, "엑셀 파일을 읽을 수 없습니다.")
		return
	}
	defer func() { _ = workbook.Close() }()

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		renderBLMarkingsListError(w, r, "엑셀 시트를 찾을 수 없습니다.")
		return
	}

	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		renderBLMarkingsListError(w, r, "엑셀 데이터를 읽을 수 없습니다.")
		return
	}
	if len(rows) == 0 {
		renderBLMarkingsListError(w, r, "업로드할 데이터가 없습니다.")
		return
	}

	startIdx := 0
	if looksLikeBLMarkingHeader(rows[0]) {
		startIdx = 1
	}

	inserted := 0
	skipped := 0
	for rowIdx := startIdx; rowIdx < len(rows); rowIdx++ {
		if inserted >= maxUploadRows {
			break
		}
		row := rows[rowIdx]
		if len(row) == 0 {
			continue
		}
		hblNo := strings.TrimSpace(rowValue(row, 0))
		marks := strings.TrimSpace(rowValue(row, 1))
		if hblNo == "" && marks == "" {
			continue
		}
		if hblNo == "" || marks == "" {
			skipped++
			continue
		}

		item := repo.BLMarking{
			ContainerID:  containerID,
			UserID:       userID,
			BLPositionID: blPositionID,
			HBLNo:        hblNo,
			Marks:        marks,
			IsActive:     true,
		}
		if err := item.Create(r.Context()); err != nil {
			renderBLMarkingsListError(w, r, "엑셀 업로드 중 오류가 발생했습니다: "+err.Error())
			return
		}
		inserted++
	}

	if inserted == 0 {
		renderBLMarkingsListError(w, r, "등록할 데이터가 없습니다.")
		return
	}

	message := "업로드 완료: " + strconv.Itoa(inserted) + "건 등록"
	if skipped > 0 {
		message += ", " + strconv.Itoa(skipped) + "건 제외"
	}
	if inserted >= maxUploadRows {
		message += " (최대 " + strconv.Itoa(maxUploadRows) + "건까지 처리)"
	}
	redirectWithSuccess(w, r, "/admin/bl_markings", message)
}

func looksLikeBLMarkingHeader(row []string) bool {
	if len(row) < 2 {
		return false
	}
	first := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(row[0]), "_", ""))
	second := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(row[1]), "_", ""))
	return strings.Contains(first, "hbl") && strings.Contains(second, "mark")
}

func rowValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func renderBLMarkingsListError(w http.ResponseWriter, r *http.Request, message string) {
	data, err := blMarkingPageData(r.Context(), 1, "", "", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.Render(w, r, "bl_markings_list.html", view.PageData{
		Title: "BL 마킹 관리",
		Error: message,
		Data:  data,
	})
}
