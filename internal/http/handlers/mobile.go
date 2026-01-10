package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"skycontainers/internal/auth"
	"skycontainers/internal/http/middleware"
	"skycontainers/internal/pagination"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strconv"
	"strings"
	"time"
)

// Helper to get mobile layout page data
func mobilePageData(r *http.Request, title string, activeNav string, data interface{}) view.PageData {
	pd := view.PageData{
		Title: title,
		Data: map[string]interface{}{
			"ActiveNav": activeNav,
			"Data":      data,
		},
	}
	// Extract explicit Data field if map
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			if pd.Data.(map[string]interface{}) != nil {
				pd.Data.(map[string]interface{})[k] = v
			}
		}
	}
	return pd
}

// 1. Position Scan Routes
func ShowMobileScan(w http.ResponseWriter, r *http.Request) {
	// Get Positions
	posRepo := repo.BLPosition{}
	positions, err := posRepo.ListAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.Render(w, r, "mobile_scan.html", view.PageData{
		Title: "위치등록",
		Data: map[string]interface{}{
			"ActiveNav": "scan",
			"Positions": positions,
		},
	})
}

func PostMobileCheckHBL(w http.ResponseWriter, r *http.Request) {
	hblNo := strings.TrimSpace(r.FormValue("hbl_no"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if hblNo == "" {
		w.Header().Set("HX-Trigger", `{"hblCheck":{"valid":false}}`)
		return
	}

	markingRepo := repo.BLMarking{}
	// Just check if exists.
	// We can use GetByHBLNo
	item, err := markingRepo.GetByHBLNo(r.Context(), hblNo)

	if err != nil || item == nil {
		w.Header().Set("HX-Trigger", `{"hblCheck":{"valid":false}}`)
		fmt.Fprint(w, `<span class="input-status error">⚠️ 존재하지 않는 HBL입니다.</span>`)
	} else {
		w.Header().Set("HX-Trigger", `{"hblCheck":{"valid":true}}`)
		fmt.Fprint(w, `<span class="input-status ok">✅ 확인되었습니다.</span>`)
	}
}

func PostMobileScanSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	positionID, _ := strconv.ParseInt(r.FormValue("position_id"), 10, 64)
	hblNo := strings.TrimSpace(r.FormValue("hbl_no"))

	if positionID <= 0 || hblNo == "" {
		w.Header().Set("HX-Reswap", "innerHTML")
		fmt.Fprint(w, `<div class="toast toast-error">필수 정보가 누락되었습니다.</div>`)
		return
	}

	markingRepo := repo.BLMarking{}
	// First find the marking
	item, err := markingRepo.GetByHBLNo(r.Context(), hblNo)
	if err != nil || item == nil {
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Header().Set("HX-Trigger", `{"hblCheck":{"valid":false}}`)
		fmt.Fprint(w, `<div class="toast toast-error">존재하지 않는 HBL 입니다.</div>`)
		return
	}

	// Update Position
	if err := markingRepo.UpdatePosition(r.Context(), item.ID, positionID); err != nil {
		w.Header().Set("HX-Reswap", "innerHTML")
		fmt.Fprint(w, `<div class="toast toast-error">저장 실패: `+err.Error()+`</div>`)
		return
	}

	w.Header().Set("HX-Reswap", "innerHTML")
	w.Header().Set("HX-Trigger", `{"hblCheck":{"valid":false},"scanSaved":{}}`)
	fmt.Fprint(w, `<div class="toast toast-success">저장되었습니다.</div>`)
}

// 2. Search Routes
func ShowMobileSearch(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, "mobile_search.html", view.PageData{
		Title: "BL 찾기",
		Data: map[string]interface{}{
			"ActiveNav": "search",
		},
	})
}

func GetMobileSearchResult(w http.ResponseWriter, r *http.Request) {
	hblNo := r.URL.Query().Get("hbl_no")

	markingRepo := repo.BLMarking{}
	item, err := markingRepo.GetByHBLNo(r.Context(), hblNo)

	w.Header().Set("Content-Type", "text/html")

	if err != nil || item == nil {
		fmt.Fprint(w, `<div class="result-card"><p class="input-status error">데이터를 찾을 수 없습니다.</p></div>`)
		return
	}

	// Output: Supplier Name & Position Name (Big)
	supplierName := item.SupplierName
	if supplierName == "" {
		supplierName = "업체 미지정"
	}
	posName := item.BLPositionName
	if posName == "" {
		posName = "위치 미지정"
	}

	fmt.Fprintf(w, `
		<div class="result-card">
			<div class="supplier-name">%s</div>
			<div class="position-name">%s</div>
		</div>
	`, supplierName, posName)
}

// 3. Leave Routes
func ShowMobileLeaves(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserKey).(*auth.User)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pager := pagination.NewPager(0, page, 10)
	reportRepo := repo.Report{}
	// Use ListByUser
	list, total, err := reportRepo.ListByUser(r.Context(), pager, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pager = pagination.NewPager(total, page, 10)

	view.Render(w, r, "mobile_leaves.html", view.PageData{
		Title: "휴가신청",
		Data: map[string]interface{}{
			"ActiveNav": "leaves",
			"Items":     list,
			"Pager":     pager,
		},
	})
}

func ShowMobileLeaveForm(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, "mobile_leaves_form.html", view.PageData{
		Title: "휴가신청 작성",
		Data: map[string]interface{}{
			"ActiveNav": "leaves",
		},
	})
}

func PostMobileLeave(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserKey).(*auth.User)

	periodStart, _ := time.Parse("2006-01-02", r.FormValue("period_start"))
	periodEnd, _ := time.Parse("2006-01-02", r.FormValue("period_end"))

	item := repo.Report{
		UserID:      user.ID,
		Subject:     r.FormValue("subject"),
		Types:       r.FormValue("type"),
		Contents:    r.FormValue("contents"),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		IsActive:    true,
	}

	if err := item.Create(r.Context()); err != nil {
		view.Render(w, r, "mobile_leaves_form.html", view.PageData{
			Error: "신청 중 오류가 발생했습니다: " + err.Error(),
			Data: map[string]interface{}{
				"ActiveNav": "leaves",
			},
		})
		return
	}

	http.Redirect(w, r, "/mobile/leaves", http.StatusSeeOther)
}

// 4. Mobile Auth
func ShowMobileLogin(w http.ResponseWriter, r *http.Request) {
	if auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/mobile/scan", http.StatusSeeOther)
		return
	}

	// Render mobile_login.html directly without layout
	tmpl, err := template.ParseFiles("web/templates/mobile_login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		CSRFToken string
		Error     string
	}{
		CSRFToken: middleware.CSRFTokenFromContext(r),
		Error:     r.URL.Query().Get("error"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostMobileLogin(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	password := r.FormValue("password")

	user, err := auth.Authenticate(r.Context(), uid, password)
	if err != nil {
		errMsg := "아이디 또는 비밀번호가 올바르지 않습니다."
		if errors.Is(err, auth.ErrUserInactive) {
			errMsg = "비활성화된 계정입니다."
		}

		tmpl, _ := template.ParseFiles("web/templates/mobile_login.html")
		data := struct {
			CSRFToken string
			Error     string
		}{
			CSRFToken: middleware.CSRFTokenFromContext(r),
			Error:     errMsg,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
		return
	}

	if err := auth.SetSession(w, r, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoItem := repo.User{}
	if err := repoItem.UpdateLastLogin(r.Context(), user.ID, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/mobile/scan", http.StatusSeeOther)
}

func PostMobileLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w, r)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/mobile/login")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/mobile/login", http.StatusSeeOther)
}
