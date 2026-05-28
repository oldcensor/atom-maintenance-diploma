package router

import (
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/internal/adapters/http/handlers"
	mw "atom-maintenance/internal/adapters/http/middleware"
	"atom-maintenance/internal/domain"
	"atom-maintenance/internal/ports"
	jwtpkg "atom-maintenance/pkg/jwt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Deps struct {
	Department          *handlers.DepartmentHandlers
	EquipmentType       *handlers.EquipmentTypeHandlers
	Equipment           *handlers.EquipmentHandlers
	MaintenanceSchedule *handlers.MaintenanceScheduleHandlers
	WorkOrder           *handlers.WorkOrderHandlers
	InspectionReport    *handlers.InspectionReportHandlers
	StatusLog           *handlers.StatusLogHandlers
	WorkOrderComment    *handlers.WorkOrderCommentHandlers
	Checklist           *handlers.ChecklistHandlers
	Auth                *handlers.AuthHandlers
	Employee            *handlers.EmployeeHandlers
	JWT           *jwtpkg.Provider
	Cache         ports.Cache
	Log           *slog.Logger
	CORSOrigins   []string
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(mw.RequestID)
	r.Use(mw.Logger(d.Log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           int(12 * time.Hour / time.Second),
	}))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", d.Auth.Login)
			r.Post("/refresh", d.Auth.Refresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(d.JWT, d.Cache))
			r.Post("/auth/logout", d.Auth.Logout)

			r.Route("/departments", func(r chi.Router) {
				r.Get("/", d.Department.List)
				r.Post("/", d.Department.Create)
				r.Get("/{id}", d.Department.GetByID)
				r.Put("/{id}", d.Department.Update)
				r.Delete("/{id}", d.Department.Delete)
			})

			r.Route("/equipment-types", func(r chi.Router) {
				r.Get("/", d.EquipmentType.List)
				r.Post("/", d.EquipmentType.Create)
				r.Get("/{id}", d.EquipmentType.GetByID)
				r.Put("/{id}", d.EquipmentType.Update)
				r.Delete("/{id}", d.EquipmentType.Delete)
			})

			r.Route("/employees", func(r chi.Router) {
				r.Get("/", d.Employee.List)
				r.With(mw.RequireRole(domain.RoleAdmin)).Post("/", d.Employee.Create)
				r.Get("/{id}", d.Employee.GetByID)
				r.With(mw.RequireRole(domain.RoleManager)).Put("/{id}", d.Employee.Update)
				r.With(mw.RequireRole(domain.RoleAdmin)).Delete("/{id}", d.Employee.Delete)
			})


			r.Route("/equipment", func(r chi.Router) {
				r.Get("/", d.Equipment.List)
				r.With(mw.RequireRole(domain.RoleManager)).Post("/", d.Equipment.Create)
				r.Get("/{id}", d.Equipment.GetByID)
				r.With(mw.RequireRole(domain.RoleManager)).Put("/{id}", d.Equipment.Update)
				r.With(mw.RequireRole(domain.RoleManager)).Delete("/{id}", d.Equipment.Delete)
			})

			r.Route("/maintenance-schedules", func(r chi.Router) {
				r.Get("/", d.MaintenanceSchedule.List)
				r.With(mw.RequireRole(domain.RoleEngineer)).Post("/", d.MaintenanceSchedule.Create)
				r.Get("/{id}", d.MaintenanceSchedule.GetByID)
				r.With(mw.RequireRole(domain.RoleEngineer)).Put("/{id}", d.MaintenanceSchedule.Update)
				r.With(mw.RequireRole(domain.RoleEngineer)).Delete("/{id}", d.MaintenanceSchedule.Delete)
			})

			r.Route("/work-orders", func(r chi.Router) {
				r.Get("/", d.WorkOrder.List)
				r.With(mw.RequireRole(domain.RoleEngineer)).Post("/", d.WorkOrder.Create)
				r.Get("/{id}", d.WorkOrder.GetByID)
				r.With(mw.RequireRole(domain.RoleTechnician)).Put("/{id}", d.WorkOrder.Update)
				r.With(mw.RequireRole(domain.RoleEngineer)).Delete("/{id}", d.WorkOrder.Delete)

				r.Get("/{id}/status-log", d.StatusLog.List)

				r.Route("/{woID}/comments", func(r chi.Router) {
					r.Get("/", d.WorkOrderComment.List)
					r.Post("/", d.WorkOrderComment.Create)
					r.With(mw.RequireRole(domain.RoleManager)).Delete("/{id}", d.WorkOrderComment.Delete)
				})

				r.Route("/{woID}/checklist", func(r chi.Router) {
					r.Get("/", d.Checklist.List)
					r.With(mw.RequireRole(domain.RoleEngineer)).Post("/", d.Checklist.Create)
					r.Patch("/{itemID}", d.Checklist.Toggle)
					r.With(mw.RequireRole(domain.RoleEngineer)).Delete("/{itemID}", d.Checklist.Delete)
				})
			})

			r.Route("/inspection-reports", func(r chi.Router) {
				r.Get("/", d.InspectionReport.List)
				r.Post("/", d.InspectionReport.Create)
				r.Get("/{id}", d.InspectionReport.GetByID)
				r.Delete("/{id}", d.InspectionReport.Delete)
			})
		})
	})

	return r
}
