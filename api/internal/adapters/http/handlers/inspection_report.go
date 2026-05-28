package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"

	"github.com/go-playground/validator/v10"
)

type InspectionReportHandlers struct {
	app      *app.InspectionReportApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewInspectionReportHandlers(app *app.InspectionReportApp, log *slog.Logger) *InspectionReportHandlers {
	return &InspectionReportHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *InspectionReportHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.inspection_report", "op", "Create")
	start := time.Now()

	var in dto.CreateInspectionReportRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	rep, err := h.app.Create(ctx, &domain.InspectionReport{
		WorkOrderID:     in.WorkOrderID,
		InspectorID:     in.InspectorID,
		Findings:        in.Findings,
		Recommendations: in.Recommendations,
	})
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("inspection report created", "id", rep.ID, "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusCreated, toInspectionReportResponse(*rep))
}

func (h *InspectionReportHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.inspection_report", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.InspectionReportResponse, 0, len(items))
	for _, rep := range items {
		out = append(out, toInspectionReportResponse(rep))
	}

	log.Info("inspection reports listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, out)
}

func (h *InspectionReportHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.inspection_report", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	rep, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("inspection report fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toInspectionReportResponse(*rep))
}

func (h *InspectionReportHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.inspection_report", "op", "Delete")
	// Протокол выполнения — юридически значимый документ: изменение и удаление
	// запрещены после создания (ФТ-8).
	log.Warn("delete forbidden: inspection report is immutable")
	respond.Error(w, domain.ErrForbidden)
}

func toInspectionReportResponse(rep domain.InspectionReport) dto.InspectionReportResponse {
	return dto.InspectionReportResponse{
		ID:              rep.ID,
		WorkOrderID:     rep.WorkOrderID,
		InspectorID:     rep.InspectorID,
		Findings:        rep.Findings,
		Recommendations: rep.Recommendations,
		CreatedAt:       rep.CreatedAt,
	}
}
