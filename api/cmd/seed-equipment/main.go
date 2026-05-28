package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	postgresdb "atom-maintenance/internal/adapters/db/postgres"
	"atom-maintenance/internal/config"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg"
	applogger "atom-maintenance/platform/logger"
)

var departments = []struct {
	Name        string
	Description string
}{
	{"Реакторный цех", "Основной отдел реакторного отделения"},
	{"Котельная", "Отдел тепловыделения и охлаждения"},
	{"Турбинный цех", "Отдел турбогенератора и энергосистемы"},
	{"Электроцех", "Отдел электрооборудования и систем управления"},
	{"Система охлаждения", "Отдел первого и второго контуров охлаждения"},
}

var simulatorEquipment = []struct {
	Name          string
	Serial        string
	MeterType     string
	Description   string
	DepartmentIdx int // index в массив departments
}{
	{"ГЦН-1", "GCN-001", "operating_hours", "Главный циркуляционный насос 1", 4},
	{"ГЦН-2", "GCN-002", "operating_hours", "Главный циркуляционный насос 2", 4},
	{"Кран-регулятор КР-7", "KR-007", "cycles", "Кран-регулятор (счётчик циклов)", 0},
	{"Аварийный дизель-генератор", "DG-101", "operating_hours", "Аварийный дизель-генератор", 2},
	{"Теплообменник ТО-3", "HE-003", "days", "Теплообменник ТО-3 (календарные дни)", 1},
}

var employees = []struct {
	Email         string
	FullName      string
	Role          domain.EmployeeRole
	DepartmentIdx int
}{
	{"technician@example.com", "Иван Сидоров", domain.RoleTechnician, 0},
	{"technician2@example.com", "Павел Морозов", domain.RoleTechnician, 4},
	{"engineer@example.com", "Петр Васильев", domain.RoleEngineer, 1},
	{"engineer2@example.com", "Максим Федоров", domain.RoleEngineer, 2},
	{"manager@example.com", "Сергей Петров", domain.RoleManager, 3},
	{"admin@example.com", "Александр Иванов", domain.RoleAdmin, 0},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	log := applogger.New(cfg.Logger)

	db, err := postgresdb.New(cfg.Database)
	if err != nil {
		log.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	ctx := context.Background()
	to := cfg.Timeouts.DBQuery

	typeRepo := postgresdb.NewEquipmentTypeRepo(db, log, to)
	eqRepo := postgresdb.NewEquipmentRepo(db, log, to)
	deptRepo := postgresdb.NewDepartmentRepo(db, log, to)
	empRepo := postgresdb.NewEmployeeRepo(db, log, to)

	types, err := typeRepo.List(ctx)
	if err != nil {
		log.Error("list equipment types", "err", err)
		os.Exit(1)
	}

	var typeID int64
	for _, t := range types {
		if t.Name == "Промышленное оборудование" {
			typeID = t.ID
			break
		}
	}
	if typeID == 0 {
		created, err := typeRepo.Create(ctx, &domain.EquipmentType{
			Name:        "Промышленное оборудование",
			Description: "Базовый тип для демо-оборудования симулятора",
		})
		if err != nil {
			log.Error("create equipment type", "err", err)
			os.Exit(1)
		}
		typeID = created.ID
		fmt.Printf("Equipment type created: id=%d name=%s\n", typeID, created.Name)
	} else {
		fmt.Printf("Equipment type already exists: id=%d\n", typeID)
	}

	fmt.Println("\n--- Seeding Departments ---")
	deptMap := make(map[int]int64)
	for i, dept := range departments {
		existingDepts, err := deptRepo.List(ctx)
		if err != nil {
			log.Error("list departments", "err", err)
			os.Exit(1)
		}

		var found bool
		for _, d := range existingDepts {
			if d.Name == dept.Name {
				deptMap[i] = d.ID
				fmt.Printf("  skip  %s (already exists)\n", dept.Name)
				found = true
				break
			}
		}

		if !found {
			created, err := deptRepo.Create(ctx, &domain.Department{
				Name:        dept.Name,
				Description: dept.Description,
			})
			if err != nil {
				log.Error("create department", "name", dept.Name, "err", err)
				os.Exit(1)
			}
			deptMap[i] = created.ID
			fmt.Printf("  created id=%-3d  %s\n", created.ID, created.Name)
		}
		time.Sleep(5 * time.Millisecond)
	}

	fmt.Println("\n--- Seeding Equipment ---")
	existing, err := eqRepo.List(ctx)
	if err != nil {
		log.Error("list equipment", "err", err)
		os.Exit(1)
	}
	existingNames := make(map[string]bool, len(existing))
	for _, e := range existing {
		existingNames[e.Name] = true
	}

	eqCreated := 0
	for _, item := range simulatorEquipment {
		if existingNames[item.Name] {
			fmt.Printf("  skip  %s (already exists)\n", item.Name)
			continue
		}
		deptID := deptMap[item.DepartmentIdx]
		eq, err := eqRepo.Create(ctx, &domain.Equipment{
			Name:            item.Name,
			Description:     item.Description,
			SerialNumber:    item.Serial,
			EquipmentTypeID: typeID,
			DepartmentID:    &deptID,
			Status:          domain.StatusActive,
		})
		if err != nil {
			log.Error("create equipment", "name", item.Name, "err", err)
			os.Exit(1)
		}
		fmt.Printf("  created id=%-3d  %s\n", eq.ID, eq.Name)
		eqCreated++
		time.Sleep(5 * time.Millisecond)
	}

	if eqCreated == 0 {
		fmt.Println("All equipment already exists, nothing to seed.")
	} else {
		fmt.Printf("Created %d equipment items.\n", eqCreated)
		fmt.Println("IDs should match simulator equipment_id (1–5).")
		fmt.Println("Verify at: http://localhost:8090/api/v1/telemetry")
	}

	fmt.Println("\n--- Seeding Employees ---")
	hash, err := pkg.HashPassword("123456789")
	if err != nil {
		log.Error("hash password", "err", err)
		os.Exit(1)
	}

	for _, emp := range employees {
		_, err := empRepo.GetByEmail(ctx, emp.Email)
		if err == nil {
			fmt.Printf("  skip  %s (already exists)\n", emp.Email)
			continue
		}

		deptID := deptMap[emp.DepartmentIdx]
		created, err := empRepo.Create(ctx, &domain.Employee{
			Email:        emp.Email,
			PasswordHash: hash,
			FullName:     emp.FullName,
			Role:         emp.Role,
			DepartmentID: &deptID,
		})
		if err != nil {
			log.Error("create employee", "email", emp.Email, "err", err)
			os.Exit(1)
		}
		fmt.Printf("  created id=%-3d  %s (%s)\n", created.ID, created.Email, created.Role)
		time.Sleep(5 * time.Millisecond)
	}

	fmt.Println("\nDepartments and employees seeding complete.")
}
