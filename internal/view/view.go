package view

import (
	"html/template"
	"net/http"
	"path/filepath"
	"skycontainers/internal/auth"
	"skycontainers/internal/http/middleware"
	"skycontainers/internal/policy"
	"strconv"
	"strings"
	"time"
)

var templates = make(map[string]*template.Template)
var funcMap template.FuncMap

func InitTemplates() {
	layoutPath := filepath.Join("web", "templates", "layout.html")
	pages, err := filepath.Glob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		panic(err)
	}

	funcMap = template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, nil
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"add": func(a, b int) int {
			return a + b
		},
		"formatDate": func(value interface{}) string {
			switch v := value.(type) {
			case time.Time:
				if v.IsZero() {
					return ""
				}
				return v.Format("2006-01-02")
			case *time.Time:
				if v == nil || v.IsZero() {
					return ""
				}
				return v.Format("2006-01-02")
			default:
				return ""
			}
		},
		"formatDateTime": func(value interface{}) string {
			switch v := value.(type) {
			case time.Time:
				if v.IsZero() {
					return ""
				}
				return v.Format("2006-01-02T15:04")
			case *time.Time:
				if v == nil || v.IsZero() {
					return ""
				}
				return v.Format("2006-01-02T15:04")
			default:
				return ""
			}
		},
		"stripParen": func(value string) string {
			if value == "" {
				return ""
			}
			if idx := strings.Index(value, "("); idx >= 0 {
				return strings.TrimSpace(value[:idx])
			}
			return strings.TrimSpace(value)
		},
		"truncateText": func(value string, limit int) string {
			value = strings.TrimSpace(value)
			if value == "" || limit <= 0 {
				return value
			}
			runes := []rune(value)
			if len(runes) <= limit {
				return value
			}
			return strings.TrimSpace(string(runes[:limit])) + "..."
		},
		"reportTypeLabel": func(value string) string {
			switch strings.TrimSpace(value) {
			case "1":
				return "반차"
			case "2":
				return "연차"
			case "3":
				return "경조"
			case "4":
				return "병가"
			case "5":
				return "무급"
			case "6":
				return "기타"
			default:
				return strings.TrimSpace(value)
			}
		},
		"canAccess": func(user *auth.User, action string, resource string, ownerID ...int64) bool {
			if user == nil {
				return false
			}
			var targetID int64
			if len(ownerID) > 0 {
				targetID = ownerID[0]
			}
			return policy.Allow(user, policy.Action(strings.ToLower(strings.TrimSpace(action))),
				policy.Resource(strings.ToLower(strings.TrimSpace(resource))), targetID)
		},
		"formatID": func(value interface{}) string {
			switch v := value.(type) {
			case int64:
				if v == 0 {
					return ""
				}
				return strconv.FormatInt(v, 10)
			case *int64:
				if v == nil || *v == 0 {
					return ""
				}
				return strconv.FormatInt(*v, 10)
			default:
				return ""
			}
		},
		"int64PtrEq": func(value *int64, target int64) bool {
			if value == nil {
				return false
			}
			return *value == target
		},
	}

	for _, page := range pages {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}

		tmpl := template.New(name).Funcs(funcMap)
		// Each template is Layout + Specific Page
		tmpl, err = tmpl.ParseFiles(layoutPath, page)
		if err != nil {
			panic(err)
		}
		templates[name] = tmpl
	}
}

type PageData struct {
	Title         string
	User          *auth.User
	Authenticated bool
	Data          interface{}
	Error         string
	Success       string
	CSRFToken     string
}

func Render(w http.ResponseWriter, r *http.Request, name string, data PageData) {
	if user, ok := r.Context().Value(middleware.UserKey).(*auth.User); ok {
		data.User = user
		data.Authenticated = true
	}

	if token := middleware.CSRFTokenFromContext(r); token != "" {
		data.CSRFToken = token
	}

	if data.Success == "" {
		if msg := strings.TrimSpace(r.URL.Query().Get("success")); msg != "" {
			data.Success = msg
		}
	}

	if data.Error == "" {
		if msg := strings.TrimSpace(r.URL.Query().Get("error")); msg != "" {
			data.Error = msg
		}
	}

	// Hot-reload templates for development
	// In a real production environment, you might want to use the cached 'templates' map
	// or perform this check based on a configuration flag.
	// Determine layout based on prefix
	layoutName := "layout.html"
	if strings.HasPrefix(name, "mobile_") {
		layoutName = "layout_mobile.html"
	}

	layoutPath := filepath.Join("web", "templates", layoutName)
	pagePath := filepath.Join("web", "templates", name)

	tmpl := template.New(name).Funcs(funcMap)
	var err error
	tmpl, err = tmpl.ParseFiles(layoutPath, pagePath)
	if err != nil {
		// Fallback to cached templates if file read fails (though unlikely if glob worked)
		var ok bool
		tmpl, ok = templates[name]
		if !ok {
			http.Error(w, "Template not found: "+name+" ("+err.Error()+")", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		err := tmpl.ExecuteTemplate(w, "content", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Always execute the layout, which will include the specific page's blocks
	err = tmpl.ExecuteTemplate(w, layoutName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
