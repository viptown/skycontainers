package middleware

import (
	"net/http"
	"strings"
)

const maxFormBytes = 10 << 20 // 10MB

// SanitizeForm trims inputs and blocks suspicious control characters.
// It also limits form size to reduce abuse.
func SanitizeForm(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
			if err := r.ParseForm(); err != nil {
				http.Error(w, "요청 데이터가 너무 큽니다.", http.StatusRequestEntityTooLarge)
				return
			}

			for key, values := range r.PostForm {
				for i, value := range values {
					clean := strings.TrimSpace(value)
					if strings.ContainsRune(clean, '\x00') {
						http.Error(w, "잘못된 입력이 포함되어 있습니다.", http.StatusBadRequest)
						return
					}
					r.PostForm[key][i] = clean
				}
			}

			r.Form = r.PostForm
		}

		next.ServeHTTP(w, r)
	})
}
