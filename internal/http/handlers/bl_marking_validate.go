package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"skycontainers/internal/policy"
	"skycontainers/internal/repo"
)

type containerValidationResponse struct {
	OK          bool   `json:"ok"`
	Message     string `json:"message"`
	ContainerID int64  `json:"container_id"`
}

func ValidateBLMarkingContainer(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePermission(w, r, policy.ActionRead, policy.ResourceBLMarkings, 0, "BL 마킹 관리"); !ok {
		return
	}
	containerNo := strings.TrimSpace(r.URL.Query().Get("container_no"))
	if containerNo == "" {
		writeContainerValidation(w, false, "컨테이너 번호를 입력해 주세요.")
		return
	}

	repoContainer := repo.Container{}
	container, err := repoContainer.FindAvailableByNo(r.Context(), containerNo)
	if err != nil {
		if errors.Is(err, repo.ErrContainerUnavailable) {
			writeContainerValidation(w, false, "등록되지 않았거나 이미 출고된 컨테이너입니다.")
			return
		}
		writeContainerValidation(w, false, "컨테이너 확인 중 오류가 발생했습니다.")
		return
	}

	writeContainerValidation(w, true, "사용 가능한 컨테이너입니다.", container.ID)
}

func writeContainerValidation(w http.ResponseWriter, ok bool, message string, containerID ...int64) {
	var id int64
	if len(containerID) > 0 {
		id = containerID[0]
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(containerValidationResponse{
		OK:          ok,
		Message:     message,
		ContainerID: id,
	})
}
