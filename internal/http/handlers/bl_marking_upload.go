package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"
	"strings"
	"time"

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

	if header == nil {
		renderBLMarkingsListError(w, r, "업로드 파일을 확인해 주세요.")
		return
	}

	rows, err := readBLMarkingRows(file, header.Filename)
	if err != nil {
		renderBLMarkingsListError(w, r, err.Error())
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
		hblNo := strings.TrimSpace(cleanCell(rowValue(row, 0)))
		marks := strings.TrimSpace(cleanCell(rowValue(row, 1)))
		if isBLMarkingHeaderRow(hblNo, marks) {
			continue
		}
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
		if unipassXML, ok := fetchUnipassXML(r.Context(), hblNo); ok {
			item.FrmUnipass = &unipassXML
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
	return isBLMarkingHeaderRow(row[0], row[1])
}

func rowValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func cleanCell(value string) string {
	return strings.TrimPrefix(value, "\ufeff")
}

func isBLMarkingHeaderRow(hblNo string, marks string) bool {
	first := normalizeHeaderToken(hblNo)
	second := normalizeHeaderToken(marks)
	if first == "" || second == "" {
		return false
	}
	if strings.IndexFunc(first, isDigit) >= 0 || strings.IndexFunc(second, isDigit) >= 0 {
		return false
	}
	if !containsAny(first, "hbl", "housebl", "houseblno", "blno") {
		return false
	}
	if !containsAny(second, "mark", "marks") {
		return false
	}
	return true
}

func normalizeHeaderToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, "/", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func readBLMarkingRows(file io.Reader, filename string) ([][]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		workbook, err := excelize.OpenReader(file)
		if err != nil {
			return nil, fmt.Errorf("엑셀 파일을 읽을 수 없습니다.")
		}
		defer func() { _ = workbook.Close() }()

		sheets := workbook.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("엑셀 시트를 찾을 수 없습니다.")
		}

		rows, err := workbook.GetRows(sheets[0])
		if err != nil {
			return nil, fmt.Errorf("엑셀 데이터를 읽을 수 없습니다.")
		}
		return rows, nil
	case ".csv":
		reader := csv.NewReader(file)
		reader.TrimLeadingSpace = true
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("CSV 데이터를 읽을 수 없습니다.")
		}
		return rows, nil
	default:
		return nil, fmt.Errorf("xlsx 또는 csv 파일만 업로드할 수 있습니다.")
	}
}

func fetchUnipassXML(ctx context.Context, hblNo string) (string, bool) {
	hblNo = strings.TrimSpace(hblNo)
	if hblNo == "" {
		return "", false
	}
	apiKey := strings.TrimSpace(os.Getenv("crkyCn"))
	if apiKey == "" {
		return "", false
	}
	baseURL := "https://unipass.customs.go.kr:38010/ext/rest/cargCsclPrgsInfoQry/retrieveCargCsclPrgsInfo"
	query := url.Values{}
	query.Set("crkyCn", apiKey)
	query.Set("hblNo", hblNo)
	query.Set("blYy", time.Now().Format("2006"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", false
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}
	xmlBody := strings.TrimSpace(string(body))
	if xmlBody == "" {
		return "", false
	}
	if !unipassResultOK(xmlBody) {
		return "", false
	}
	return xmlBody, true
}

func unipassResultOK(xmlBody string) bool {
	code := extractXMLTagValue(xmlBody, "resultCode")
	if code == "" {
		return !isUnipassErrorBody(xmlBody)
	}
	code = strings.TrimSpace(code)
	if code != "00" {
		return false
	}
	return !isUnipassErrorBody(xmlBody)
}

func extractXMLTagValue(xmlBody string, tag string) string {
	startTag := "<" + tag + ">"
	endTag := "</" + tag + ">"
	start := strings.Index(xmlBody, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)
	end := strings.Index(xmlBody[start:], endTag)
	if end == -1 {
		return ""
	}
	return xmlBody[start : start+end]
}

func isUnipassErrorBody(xmlBody string) bool {
	notice := strings.TrimSpace(extractXMLTagValue(xmlBody, "ntceInfo"))
	if notice != "" {
		return true
	}
	tcnt := strings.TrimSpace(extractXMLTagValue(xmlBody, "tCnt"))
	if tcnt == "-1" || tcnt == "0" {
		return true
	}
	return false
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
