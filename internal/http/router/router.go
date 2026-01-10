package router

import (
	"context"
	"net/http"
	"skycontainers/internal/auth"
	"skycontainers/internal/http/handlers"
	"skycontainers/internal/http/middleware"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SanitizeForm)
	r.Use(middleware.CSRFMiddleware)

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Auth routes
	r.Get("/login", handlers.ShowLogin)
	r.Post("/login", handlers.PostLogin)
	r.Post("/logout", handlers.PostLogout)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Detect mobile devices
		userAgent := r.Header.Get("User-Agent")
		if strings.Contains(strings.ToLower(userAgent), "mobile") ||
			strings.Contains(strings.ToLower(userAgent), "android") ||
			strings.Contains(strings.ToLower(userAgent), "iphone") {
			http.Redirect(w, r, "/mobile/login", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRequired)
		r.Get("/admin", handlers.ShowDashboard)

		r.Route("/admin", func(r chi.Router) {
			r.Get("/dashboard", handlers.ShowDashboard)
			r.Route("/io_management", func(r chi.Router) {
				r.Get("/", handlers.ShowIOManagement)
				r.Get("/{id}/inbound_modal", handlers.ShowInboundModal)
				r.Post("/{id}/inbound", handlers.PostIOInbound)
				r.Post("/{id}/processing", handlers.PostIOProcessing)
				r.Post("/{id}/outbound", handlers.PostIOOutbound)
			})
			r.Route("/container_types", func(r chi.Router) {
				r.Get("/", handlers.ListContainerTypes)
				r.Get("/new", handlers.ShowCreateContainerType)
				r.Post("/", handlers.PostCreateContainerType)
				r.Get("/{id}/edit", handlers.ShowEditContainerType)
				r.Post("/{id}/edit", handlers.PostUpdateContainerType)
				r.Delete("/{id}", handlers.DeleteContainerType)
			})

			r.Route("/suppliers", func(r chi.Router) {
				r.Get("/", handlers.ListSuppliers)
				r.Get("/new", handlers.ShowCreateSupplier)
				r.Post("/", handlers.PostCreateSupplier)
				r.Get("/{id}/edit", handlers.ShowEditSupplier)
				r.Post("/{id}/edit", handlers.PostUpdateSupplier)
				r.Delete("/{id}", handlers.DeleteSupplier)
			})

			r.Route("/containers", func(r chi.Router) {
				r.Get("/", handlers.ListContainers)
				r.Get("/export", handlers.ExportContainers)
				r.Get("/new", handlers.ShowCreateContainer)
				r.Post("/", handlers.PostCreateContainer)
				r.Get("/{id}/edit", handlers.ShowEditContainer)
				r.Post("/{id}/edit", handlers.PostUpdateContainer)
				r.Delete("/{id}", handlers.DeleteContainer)
			})

			r.Route("/users", func(r chi.Router) {
				r.Get("/", handlers.ListUsers)
				r.Get("/new", handlers.ShowCreateUser)
				r.Post("/", handlers.PostCreateUser)
				r.Get("/{id}/edit", handlers.ShowEditUser)
				r.Post("/{id}/edit", handlers.PostUpdateUser)
				r.Post("/{id}/status", handlers.PostUpdateUserStatus)
				r.Delete("/{id}", handlers.DeleteUser)
			})
			r.Route("/policies", func(r chi.Router) {
				r.Get("/", handlers.ShowPolicySettings)
				r.Post("/", handlers.PostUpdatePolicySettings)
			})

			r.Route("/reports", func(r chi.Router) {
				r.Get("/", handlers.ListReports)
				r.Get("/new", handlers.ShowCreateReport)
				r.Post("/", handlers.PostCreateReport)
				r.Get("/{id}/view", handlers.ShowReport)
				r.Get("/{id}/edit", handlers.ShowEditReport)
				r.Post("/{id}/edit", handlers.PostUpdateReport)
				r.Delete("/{id}", handlers.DeleteReport)
			})

			r.Route("/bl_markings", func(r chi.Router) {
				r.Get("/", handlers.ListBLMarkings)
				r.Get("/export", handlers.ExportBLMarkings)
				r.Get("/cargo_card", handlers.ShowBLCargoCards)
				r.Post("/apply_unipass", handlers.PostApplyUnipassFiltered)
				r.Post("/delete_filtered", handlers.PostDeleteBLMarkingsFiltered)
				r.Get("/new", handlers.ShowCreateBLMarking)
				r.Post("/", handlers.PostCreateBLMarking)
				r.Post("/upload", handlers.PostUploadBLMarkings)
				r.Get("/validate-container", handlers.ValidateBLMarkingContainer)
				r.Get("/{id}/edit", handlers.ShowEditBLMarking)
				r.Post("/{id}/edit", handlers.PostUpdateBLMarking)
				r.Post("/{id}/status", handlers.PostUpdateBLMarkingStatus)
				r.Delete("/{id}", handlers.DeleteBLMarking)
			})

			r.Route("/bl_positions", func(r chi.Router) {
				r.Get("/", handlers.ListBLPositions)
				r.Get("/new", handlers.ShowCreateBLPosition)
				r.Post("/", handlers.PostCreateBLPosition)
				r.Get("/{id}/edit", handlers.ShowEditBLPosition)
				r.Post("/{id}/edit", handlers.PostUpdateBLPosition)
				r.Post("/{id}/status", handlers.PostUpdateBLPositionStatus)
				r.Delete("/{id}", handlers.DeleteBLPosition)
			})

			r.Route("/carnumbers", func(r chi.Router) {
				r.Get("/", handlers.ListCarNumbers)
				r.Get("/new", handlers.ShowCreateCarNumber)
				r.Post("/", handlers.PostCreateCarNumber)
				r.Get("/{id}/edit", handlers.ShowEditCarNumber)
				r.Post("/{id}/edit", handlers.PostUpdateCarNumber)
				r.Delete("/{id}", handlers.DeleteCarNumber)
			})
		})

		r.Route("/supplier", func(r chi.Router) {
			r.Get("/portal", handlers.ShowSupplierPortal)
		})
	})

	// Mobile Routes
	r.Route("/mobile", func(r chi.Router) {
		r.Get("/login", handlers.ShowMobileLogin)
		r.Post("/login", handlers.PostMobileLogin)

		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !auth.IsAuthenticated(r) {
						http.Redirect(w, r, "/mobile/login", http.StatusSeeOther)
						return
					}
					session, _ := auth.Store.Get(r, "session-name")
					if session.Values["user_id"] == nil {
						http.Redirect(w, r, "/mobile/login", http.StatusSeeOther)
						return
					}

					user := &auth.User{
						ID:   session.Values["user_id"].(int64),
						Name: session.Values["user_name"].(string),
						Role: session.Values["user_role"].(string),
					}
					ctx := context.WithValue(r.Context(), middleware.UserKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/mobile/scan", http.StatusFound)
			})
			r.Get("/scan", handlers.ShowMobileScan)
			r.Post("/scan", handlers.PostMobileScanSave)
			r.Post("/check_hbl", handlers.PostMobileCheckHBL)
			r.Post("/logout", handlers.PostMobileLogout)

			r.Get("/search", handlers.ShowMobileSearch)
			r.Get("/search_result", handlers.GetMobileSearchResult)

			r.Get("/leaves", handlers.ShowMobileLeaves)
			r.Get("/leaves/new", handlers.ShowMobileLeaveForm)
			r.Post("/leaves/new", handlers.PostMobileLeave)
		})
	})

	return r
}
