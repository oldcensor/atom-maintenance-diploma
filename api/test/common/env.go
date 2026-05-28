package common

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	postgresdb "atom-maintenance/internal/adapters/db/postgres"
	"atom-maintenance/internal/adapters/http/handlers"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/config"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/authctx"
	jwtpkg "atom-maintenance/pkg/jwt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type Env struct {
	DB                  *gorm.DB
	Server              *httptest.Server
	Client              *http.Client
	Base                string
	Log                 *slog.Logger
	MaintenanceSchedule domain.MaintenanceScheduleRepository
	WorkOrder           domain.WorkOrderRepository
	TxManager           domain.TxManager
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("go.mod not found: cannot determine project root")
		}
		dir = parent
	}
	if err := os.Chdir(dir); err != nil {
		panic("chdir to project root: " + err.Error())
	}
}

func MustStart(t *testing.T) *Env {
	t.Helper()

	if err := os.Setenv("CONFIG_PATH", "configs/config.test.yml"); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	db, err := postgresdb.New(cfg.Database)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	EnsureTestEmployee(t, db)

	cache := NewMemCache()
	jwt := jwtpkg.New(cfg.JWT)

	departmentRepo := postgresdb.NewDepartmentRepo(db, log, cfg.Timeouts.DBQuery)
	equipmentTypeRepo := postgresdb.NewEquipmentTypeRepo(db, log, cfg.Timeouts.DBQuery)
	equipmentRepo := postgresdb.NewEquipmentRepo(db, log, cfg.Timeouts.DBQuery)
	employeeRepo := postgresdb.NewEmployeeRepo(db, log, cfg.Timeouts.DBQuery)
	maintenanceScheduleRepo := postgresdb.NewMaintenanceScheduleRepo(db, log, cfg.Timeouts.DBQuery)
	workOrderRepo := postgresdb.NewWorkOrderRepo(db, log, cfg.Timeouts.DBQuery)
	statusLogRepo := postgresdb.NewWorkOrderStatusLogRepo(db, log, cfg.Timeouts.DBQuery)
	woCommentRepo := postgresdb.NewWorkOrderCommentRepo(db, log, cfg.Timeouts.DBQuery)
	checklistRepo := postgresdb.NewWorkOrderChecklistRepo(db, log, cfg.Timeouts.DBQuery)
	inspectionReportRepo := postgresdb.NewInspectionReportRepo(db, log, cfg.Timeouts.DBQuery)
	txManager := postgresdb.NewTxManager(db)

	departmentApp := app.NewDepartmentApp(departmentRepo, log)
	equipmentTypeApp := app.NewEquipmentTypeApp(equipmentTypeRepo, log)
	equipmentApp := app.NewEquipmentApp(equipmentRepo, log)
	employeeApp := app.NewEmployeeApp(employeeRepo, cache, cfg.JWT.AccessTTL, log)
	authApp := app.NewAuthApp(employeeRepo, cache, jwt, log)
	maintenanceScheduleApp := app.NewMaintenanceScheduleApp(maintenanceScheduleRepo, equipmentRepo, log)
	workOrderApp := app.NewWorkOrderApp(workOrderRepo, equipmentRepo, statusLogRepo, checklistRepo, log)
	woCommentApp := app.NewWorkOrderCommentApp(woCommentRepo, log)
	woChecklistApp := app.NewWorkOrderChecklistApp(checklistRepo, log)
	inspectionReportApp := app.NewInspectionReportApp(inspectionReportRepo, workOrderRepo, log)

	departmentH := handlers.NewDepartmentHandlers(departmentApp, log)
	equipmentTypeH := handlers.NewEquipmentTypeHandlers(equipmentTypeApp, log)
	equipmentH := handlers.NewEquipmentHandlers(equipmentApp, log)
	employeeH := handlers.NewEmployeeHandlers(employeeApp, log)
	authH := handlers.NewAuthHandlers(authApp, log)
	maintenanceScheduleH := handlers.NewMaintenanceScheduleHandlers(maintenanceScheduleApp, log)
	workOrderH := handlers.NewWorkOrderHandlers(workOrderApp, log)
	statusLogH := handlers.NewStatusLogHandlers(statusLogRepo, log)
	woCommentH := handlers.NewWorkOrderCommentHandlers(woCommentApp, log)
	checklistH := handlers.NewChecklistHandlers(woChecklistApp, log)
	inspectionReportH := handlers.NewInspectionReportHandlers(inspectionReportApp, log)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(testPrincipalMiddleware)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)
		})

		r.Route("/departments", func(r chi.Router) {
			r.Get("/", departmentH.List)
			r.Post("/", departmentH.Create)
			r.Get("/{id}", departmentH.GetByID)
			r.Put("/{id}", departmentH.Update)
			r.Delete("/{id}", departmentH.Delete)
		})

		r.Route("/equipment-types", func(r chi.Router) {
			r.Get("/", equipmentTypeH.List)
			r.Post("/", equipmentTypeH.Create)
			r.Get("/{id}", equipmentTypeH.GetByID)
			r.Put("/{id}", equipmentTypeH.Update)
			r.Delete("/{id}", equipmentTypeH.Delete)
		})

		r.Route("/employees", func(r chi.Router) {
			r.Get("/", employeeH.List)
			r.Post("/", employeeH.Create)
			r.Get("/{id}", employeeH.GetByID)
			r.Put("/{id}", employeeH.Update)
			r.Delete("/{id}", employeeH.Delete)
		})

		r.Route("/equipment", func(r chi.Router) {
			r.Get("/", equipmentH.List)
			r.Post("/", equipmentH.Create)
			r.Get("/{id}", equipmentH.GetByID)
			r.Put("/{id}", equipmentH.Update)
			r.Delete("/{id}", equipmentH.Delete)
		})

		r.Route("/maintenance-schedules", func(r chi.Router) {
			r.Get("/", maintenanceScheduleH.List)
			r.Post("/", maintenanceScheduleH.Create)
			r.Get("/{id}", maintenanceScheduleH.GetByID)
			r.Put("/{id}", maintenanceScheduleH.Update)
			r.Delete("/{id}", maintenanceScheduleH.Delete)
		})

		r.Route("/work-orders", func(r chi.Router) {
			r.Get("/", workOrderH.List)
			r.Post("/", workOrderH.Create)
			r.Get("/{id}", workOrderH.GetByID)
			r.Put("/{id}", workOrderH.Update)
			r.Delete("/{id}", workOrderH.Delete)
			r.Get("/{id}/status-log", statusLogH.List)
			r.Route("/{woID}/comments", func(r chi.Router) {
				r.Get("/", woCommentH.List)
				r.Post("/", woCommentH.Create)
				r.Delete("/{id}", woCommentH.Delete)
			})
			r.Route("/{woID}/checklist", func(r chi.Router) {
				r.Get("/", checklistH.List)
				r.Post("/", checklistH.Create)
				r.Patch("/{itemID}", checklistH.Toggle)
				r.Delete("/{itemID}", checklistH.Delete)
			})
		})

		r.Route("/inspection-reports", func(r chi.Router) {
			r.Get("/", inspectionReportH.List)
			r.Post("/", inspectionReportH.Create)
			r.Get("/{id}", inspectionReportH.GetByID)
			r.Delete("/{id}", inspectionReportH.Delete)
		})
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})

	return &Env{
		DB:                  db,
		Server:              srv,
		Client:              srv.Client(),
		Base:                srv.URL + "/api/v1",
		Log:                 log,
		MaintenanceSchedule: maintenanceScheduleRepo,
		WorkOrder:           workOrderRepo,
		TxManager:           txManager,
	}
}

func EnsureTestEmployee(t *testing.T, db *gorm.DB) {
	t.Helper()
	result := db.Exec(`INSERT INTO employee (id, email, password_hash, full_name, role)
		VALUES (1, 'test-admin@atom.ru', 'dummy-hash', 'Test Admin', 'admin')
		ON CONFLICT (id) DO NOTHING`)
	if result.Error != nil {
		t.Fatalf("ensure test employee: %v", result.Error)
	}
}

func testPrincipalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := domain.EmployeeRole(r.Header.Get("X-Test-Role"))
		if role == "" {
			role = domain.RoleAdmin
		}
		p := authctx.Principal{
			EmployeeID: 1,
			Role:       role,
			JTI:        "test-jti",
			AccessTTL:  time.Hour,
		}
		next.ServeHTTP(w, r.WithContext(authctx.WithPrincipal(r.Context(), p)))
	})
}
