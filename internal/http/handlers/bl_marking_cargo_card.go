package handlers

import (
	"encoding/xml"
	"net/http"
	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"
	"strings"
	"time"
)

type cargoCardItem struct {
	BLNo          string
	ArrivalDate   string
	VesselName    string
	CargoNo       string
	Volume        string
	Weight        string
	Quantity      string
	ProductName   string
	ContainerNo   string
	Consignee     string
	Forwarder     string
	Marks         string
	HasUnipass    bool
	UnipassStatus string
}

func ShowBLCargoCards(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceBLMarkings, 0, "BL 마킹 관리"); !ok {
		return
	}

	containerNo := strings.TrimSpace(r.URL.Query().Get("container_no"))
	hblNo := strings.TrimSpace(r.URL.Query().Get("hbl_no"))
	unassignedOnly := strings.EqualFold(r.URL.Query().Get("unassigned_only"), "1") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "true") ||
		strings.EqualFold(r.URL.Query().Get("unassigned_only"), "on")

	repoItem := repo.BLMarking{}
	list, err := repoItem.ListForCargoCard(r.Context(), containerNo, hblNo, unassignedOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]cargoCardItem, 0, len(list))
	for _, item := range list {
		card := buildCargoCard(item)
		items = append(items, card)
	}

	view.Render(w, r, "bl_markings_cargo_card.html", view.PageData{
		Title: "화물카드",
		Data: map[string]interface{}{
			"Items":     items,
			"PrintedAt": time.Now().Format("2006-01-02"),
		},
	})
}

func buildCargoCard(item repo.BLMarking) cargoCardItem {
	values := map[string]string{}
	if item.FrmUnipass != nil {
		values = parseUnipassValues(*item.FrmUnipass)
	}

	arrival := formatUnipassDate(firstValue(values, "etprdt", "etprdt", "prcsdttm", "arrvdt", "inptdt", "etprde"))
	vessel := firstValue(values, "shipnm", "vslname", "vslname", "shipname")
	cargoNo := firstValue(values, "cargmtno", "cargmtno", "cargono")
	volume := combineValueUnit(
		firstValue(values, "msrm", "m3", "vol", "volume"),
		firstValue(values, "msrmut", "m3ut", "volut", "volumeut"),
	)
	weight := combineValueUnit(
		firstValue(values, "ttwg", "wght", "weight", "totwt", "wgt"),
		firstValue(values, "wghtut", "weightut", "totwtut", "wgtut"),
	)
	quantity := combineValueUnit(
		firstValue(values, "pckgcnt"),
		firstValue(values, "pckut"),
	)
	product := firstValue(values, "prnm", "prdtnm", "goodsnm", "gdsnm", "goodsname")
	containerNo := firstValue(values, "cntrno", "cntrno", "cntrno1", "container")
	consignee := strings.TrimSpace(item.SupplierName)
	forwarder := firstValue(values, "shcoflco", "frwrnm", "forwarder", "frwrdnm")

	hasUnipass := item.FrmUnipass != nil && strings.TrimSpace(*item.FrmUnipass) != ""
	status := "N"
	if hasUnipass {
		status = "Y"
	}

	return cargoCardItem{
		BLNo:          item.HBLNo,
		ArrivalDate:   arrival,
		VesselName:    vessel,
		CargoNo:       cargoNo,
		Volume:        volume,
		Weight:        weight,
		Quantity:      quantity,
		ProductName:   product,
		ContainerNo:   chooseValue(containerNo, item.ContainerNo),
		Consignee:     consignee,
		Forwarder:     forwarder,
		Marks:         item.Marks,
		HasUnipass:    hasUnipass,
		UnipassStatus: status,
	}
}

func parseUnipassValues(xmlBody string) map[string]string {
	values := make(map[string]string)
	decoder := xml.NewDecoder(strings.NewReader(xmlBody))
	var stack []string
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			stack = append(stack, strings.ToLower(t.Name.Local))
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			value := strings.TrimSpace(string(t))
			if value == "" {
				continue
			}
			key := stack[len(stack)-1]
			if _, ok := values[key]; !ok {
				values[key] = value
			}
		}
	}
	return values
}

func firstValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[strings.ToLower(key)]); value != "" {
			return value
		}
	}
	return ""
}

func chooseValue(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return strings.TrimSpace(fallback)
}

func formatUnipassDate(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 8 && isAllDigits(value[:8]) {
		return value[:4] + "-" + value[4:6] + "-" + value[6:8]
	}
	if len(value) >= 10 && value[4] == '-' && value[7] == '-' {
		return value[:10]
	}
	return value
}

func isAllDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func combineValueUnit(value string, unit string) string {
	value = strings.TrimSpace(value)
	unit = strings.TrimSpace(unit)
	if value == "" {
		return ""
	}
	if unit == "" {
		return value
	}
	return value + " " + unit
}
