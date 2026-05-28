package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	postgresdb "atom-maintenance/internal/adapters/db/postgres"
	"atom-maintenance/internal/config"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg"
	applogger "atom-maintenance/platform/logger"
)

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

	repo := postgresdb.NewEmployeeRepo(db, log, cfg.Timeouts.DBQuery)

	_, err = repo.GetByEmail(context.Background(), "admin@atom.local")
	if err == nil {
		fmt.Println("Admin already exists, skipping seed.")
		return
	}

	hash, err := pkg.HashPassword("Admin123!")
	if err != nil {
		log.Error("hash password", "err", err)
		os.Exit(1)
	}

	admin := &domain.Employee{
		Email:        "admin@atom.local",
		PasswordHash: hash,
		FullName:     "System Administrator",
		Role:         domain.RoleAdmin,
	}

	created, err := repo.Create(context.Background(), admin)
	if err != nil {
		log.Error("create admin", "err", err)
		os.Exit(1)
	}

	fmt.Printf("Admin created: id=%d email=%s\n", created.ID, created.Email)
}
