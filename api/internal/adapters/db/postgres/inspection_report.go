package postgres

import (
	"context"
	"log/slog"
	"time"

	"atom-maintenance/internal/adapters/db/postgres/model"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg"
	"atom-maintenance/platform/logger"

	"gorm.io/gorm"
)

func toInspectionReportDomain(m model.InspectionReport) domain.InspectionReport {
	return domain.InspectionReport{
		ID:              m.ID,
		WorkOrderID:     m.WorkOrderID,
		InspectorID:     m.InspectorID,
		Findings:        m.Findings,
		Recommendations: m.Recommendations,
		CreatedAt:       m.CreatedAt,
	}
}

type InspectionReportRepo struct {
	db      *gorm.DB
	log     *slog.Logger
	queryTO time.Duration
}

func NewInspectionReportRepo(db *gorm.DB, log *slog.Logger, queryTO time.Duration) *InspectionReportRepo {
	return &InspectionReportRepo{db: db, log: log, queryTO: queryTO}
}

func (r *InspectionReportRepo) Create(ctx context.Context, rep *domain.InspectionReport) (*domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.inspection_report", "op", "Create")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	m := model.InspectionReport{
		WorkOrderID:     rep.WorkOrderID,
		InspectorID:     rep.InspectorID,
		Findings:        rep.Findings,
		Recommendations: rep.Recommendations,
	}
	if err := db.Create(&m).Error; err != nil {
		mapped := pkg.MapDB(err)
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, mapped
	}

	out := toInspectionReportDomain(m)
	log.Info("inspection report created",
		"id", m.ID,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *InspectionReportRepo) GetByID(ctx context.Context, id int64) (*domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.inspection_report", "op", "GetByID")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var m model.InspectionReport
	if err := db.First(&m, id).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := toInspectionReportDomain(m)
	log.Info("inspection report fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return &out, nil
}

func (r *InspectionReportRepo) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.inspection_report", "op", "Delete")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	res := db.Delete(&model.InspectionReport{}, id)
	if res.Error != nil {
		log.Error("delete failed",
			"id", id,
			"err", res.Error,
			"duration_ms", time.Since(start).Milliseconds())
		return pkg.MapDB(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	log.Info("inspection report deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (r *InspectionReportRepo) List(ctx context.Context) ([]domain.InspectionReport, error) {
	log := logger.WithReqID(ctx, r.log).With("module", "repo.inspection_report", "op", "List")
	start := time.Now()

	db, _, cancel := withTimeout(ctx, r.db, r.queryTO)
	defer cancel()

	var rows []model.InspectionReport
	if err := db.Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, pkg.MapDB(err)
	}

	out := make([]domain.InspectionReport, 0, len(rows))
	for _, m := range rows {
		out = append(out, toInspectionReportDomain(m))
	}

	log.Info("inspection reports listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	return out, nil
}
