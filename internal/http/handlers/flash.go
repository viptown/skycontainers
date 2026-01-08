package handlers

import (
	"net/http"
	"net/url"
)

func redirectWithSuccess(w http.ResponseWriter, r *http.Request, path string, message string) {
	if message != "" {
		if parsed, err := url.Parse(path); err == nil {
			q := parsed.Query()
			q.Set("success", message)
			parsed.RawQuery = q.Encode()
			path = parsed.String()
		} else {
			path = path + "?success=" + url.QueryEscape(message)
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, path string, message string) {
	if message != "" {
		if parsed, err := url.Parse(path); err == nil {
			q := parsed.Query()
			q.Set("error", message)
			parsed.RawQuery = q.Encode()
			path = parsed.String()
		} else {
			path = path + "?error=" + url.QueryEscape(message)
		}
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}
