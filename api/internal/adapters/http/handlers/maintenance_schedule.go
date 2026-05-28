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

type MaintenanceScheduleHandlers struct {
	app      *app.MaintenanceScheduleApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewMaintenanceScheduleHandlers(app *app.MaintenanceScheduleApp, log *slog.Logger) *MaintenanceScheduleHandlers {
	return &MaintenanceScheduleHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *MaintenanceScheduleHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.schedule", "op", "Create")
	start := time.Now()

	var in dto.CreateMaintenanceScheduleRequest
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

	status := domain.ScheduleStatusScheduled
	if in.Status != "" {
		status = domain.ScheduleStatus(in.Status)
	}

	s, err := h.app.Create(ctx, &domain.MaintenanceSchedule{
		EquipmentID:   in.EquipmentID,
		ScheduledAt:   in.ScheduledAt,
		Description:   in.Description,
		AssignedTo:    in.AssignedTo,
		Status:        status,
		IntervalUnit:  in.IntervalUnit,
		IntervalValue: in.IntervalValue,
	})
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("schedule created",
		"id", s.ID,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusCreated, toScheduleResponse(*s))
}

func (h *MaintenanceScheduleHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.schedule", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.MaintenanceScheduleResponse, 0, len(items))
	for _, s := range items {
		out = append(out, toScheduleResponse(s))
	}

	log.Info("schedules listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *MaintenanceScheduleHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.schedule", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	s, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("schedule fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toScheduleResponse(*s))
}

func (h *MaintenanceScheduleHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.schedule", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateMaintenanceScheduleRequest
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

	s, err := h.app.Update(ctx, &domain.MaintenanceSchedule{
		ID:            id,
		EquipmentID:   in.EquipmentID,
		ScheduledAt:   in.ScheduledAt,
		Description:   in.Description,
		AssignedTo:    in.AssignedTo,
		Status:        domain.ScheduleStatus(in.Status),
		IntervalUnit:  in.IntervalUnit,
		IntervalValue: in.IntervalValue,
	})
	if err != nil {
		log.Error("update failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("schedule updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toScheduleResponse(*s))
}

func (h *MaintenanceScheduleHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.schedule", "op", "Delete")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	if err := h.app.Delete(ctx, id); err != nil {
		log.Error("delete failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("schedule deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	w.WriteHeader(http.StatusNoContent)
}

func toScheduleResponse(s domain.MaintenanceSchedule) dto.MaintenanceScheduleResponse {
	return dto.MaintenanceScheduleResponse{
		ID:             s.ID,
		EquipmentID:    s.EquipmentID,
		ScheduledAt:    s.ScheduledAt,
		Description:    s.Description,
		AssignedTo:     s.AssignedTo,
		Status:         string(s.Status),
		IntervalUnit:   s.IntervalUnit,
		IntervalValue:  s.IntervalValue,
		LastMeterValue: s.LastMeterValue,
		NextDueAt:      s.NextDueAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}
