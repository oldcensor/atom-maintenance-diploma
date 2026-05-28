package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"atom-maintenance/internal/adapters/cache/redis"
	postgresdb "atom-maintenance/internal/adapters/db/postgres"
	"atom-maintenance/internal/adapters/http/handlers"
	"atom-maintenance/internal/adapters/http/router"
	"atom-maintenance/internal/adapters/simulator"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/config"
	jwtpkg "atom-maintenance/pkg/jwt"
	applogger "atom-maintenance/platform/logger"
	"atom-maintenance/platform/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	log := applogger.New(cfg.Logger)
	slog.SetDefault(log)

	db, err := postgresdb.New(cfg.Database)
	if err != nil {
		log.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	log.Info("postgres connected")

	cache, rdsClient, err := redis.New(cfg.Redis, log)
	if err != nil {
		log.Error("connect redis", "err", err)
		os.Exit(1)
	}
	defer rdsClient.Close()
	log.Info("redis connected")

	jwt := jwtpkg.New(cfg.JWT)

	departmentRepo := postgresdb.NewDepartmentRepo(db, log, cfg.Timeouts.DBQuery)
	equipmentTypeRepo := postgresdb.NewEquipmentTypeRepo(db, log, cfg.Timeouts.DBQuery)
	equipmentRepo := postgresdb.NewEquipmentRepo(db, log, cfg.Timeouts.DBQuery)
	employeeRepo := postgresdb.NewEmployeeRepo(db, log, cfg.Timeouts.DBQuery)
	maintenanceScheduleRepo := postgresdb.NewMaintenanceScheduleRepo(db, log, cfg.Timeouts.DBQuery)
	workOrderRepo := postgresdb.NewWorkOrderRepo(db, log, cfg.Timeouts.DBQuery)
	inspectionReportRepo := postgresdb.NewInspectionReportRepo(db, log, cfg.Timeouts.DBQuery)
	statusLogRepo := postgresdb.NewWorkOrderStatusLogRepo(db, log, cfg.Timeouts.DBQuery)
	woCommentRepo := postgresdb.NewWorkOrderCommentRepo(db, log, cfg.Timeouts.DBQuery)
	checklistRepo := postgresdb.NewWorkOrderChecklistRepo(db, log, cfg.Timeouts.DBQuery)
	txManager := postgresdb.NewTxManager(db)

	departmentApp := app.NewDepartmentApp(departmentRepo, log)
	equipmentTypeApp := app.NewEquipmentTypeApp(equipmentTypeRepo, log)
	equipmentApp := app.NewEquipmentApp(equipmentRepo, log)
	employeeApp := app.NewEmployeeApp(employeeRepo, cache, cfg.JWT.AccessTTL, log)
	authApp := app.NewAuthApp(employeeRepo, cache, jwt, log)
	maintenanceScheduleApp := app.NewMaintenanceScheduleApp(maintenanceScheduleRepo, equipmentRepo, log)
	workOrderApp := app.NewWorkOrderApp(workOrderRepo, equipmentRepo, statusLogRepo, checklistRepo, log)
	inspectionReportApp := app.NewInspectionReportApp(inspectionReportRepo, workOrderRepo, log)
	woCommentApp := app.NewWorkOrderCommentApp(woCommentRepo, log)
	checklistApp := app.NewWorkOrderChecklistApp(checklistRepo, log)

	departmentH := handlers.NewDepartmentHandlers(departmentApp, log)
	equipmentTypeH := handlers.NewEquipmentTypeHandlers(equipmentTypeApp, log)
	equipmentH := handlers.NewEquipmentHandlers(equipmentApp, log)
	employeeH := handlers.NewEmployeeHandlers(employeeApp, log)
	authH := handlers.NewAuthHandlers(authApp, log)
	maintenanceScheduleH := handlers.NewMaintenanceScheduleHandlers(maintenanceScheduleApp, log)
	workOrderH := handlers.NewWorkOrderHandlers(workOrderApp, log)
	inspectionReportH := handlers.NewInspectionReportHandlers(inspectionReportApp, log)
	statusLogH := handlers.NewStatusLogHandlers(statusLogRepo, log)
	woCommentH := handlers.NewWorkOrderCommentHandlers(woCommentApp, log)
	checklistH := handlers.NewChecklistHandlers(checklistApp, log)

	h := router.New(router.Deps{
		Department:          departmentH,
		EquipmentType:       equipmentTypeH,
		Equipment:           equipmentH,
		MaintenanceSchedule: maintenanceScheduleH,
		WorkOrder:           workOrderH,
		InspectionReport:    inspectionReportH,
		StatusLog:           statusLogH,
		WorkOrderComment:    woCommentH,
		Checklist:           checklistH,
		Auth:                authH,
		Employee:            employeeH,
		JWT:                 jwt,
		Cache:               cache,
		Log:                 log,
		CORSOrigins:         cfg.Server.CORSAllowedOrigins,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Simulator.Enabled {
		simClient := simulator.NewClient(cfg.Simulator.URL, cfg.Simulator.Timeout)
		sched := scheduler.New(
			cfg.Scheduler.Interval,
			simClient,
			maintenanceScheduleRepo,
			workOrderRepo,
			txManager,
			log,
		)
		go sched.Start(ctx)
		defer sched.Stop()
		log.Info("scheduler started", "simulator_url", cfg.Simulator.URL)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        h,
		ReadTimeout:    cfg.Timeouts.HTTPRequest,
		WriteTimeout:   cfg.Timeouts.HTTPRequest,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}
	log.Info("server stopped")
}
