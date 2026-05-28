package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"

	"github.com/go-playground/validator/v10"
)

type EquipmentHandlers struct {
	app      *app.EquipmentApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewEquipmentHandlers(app *app.EquipmentApp, log *slog.Logger) *EquipmentHandlers {
	return &EquipmentHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *EquipmentHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment", "op", "Create")
	start := time.Now()

	var in dto.CreateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.SerialNumber = strings.TrimSpace(in.SerialNumber)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	status := domain.StatusActive
	if in.Status != "" {
		status = domain.EquipmentStatus(in.Status)
	}

	e, err := h.app.Create(ctx, &domain.Equipment{
		Name:            in.Name,
		Description:     in.Description,
		SerialNumber:    in.SerialNumber,
		EquipmentTypeID: in.EquipmentTypeID,
		DepartmentID:    in.DepartmentID,
		ResponsibleID:   in.ResponsibleID,
		Status:          status,
	})
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())

		respond.Error(w, err)
		return
	}

	log.Info("equipment created",
		"id", e.ID,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusCreated, toEquipmentResponse(*e))
}

func (h *EquipmentHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.EquipmentResponse, 0, len(items))
	for _, e := range items {
		out = append(out, toEquipmentResponse(e))
	}

	log.Info("equipment listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *EquipmentHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	e, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("equipment fetched", "id", id, "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, toEquipmentResponse(*e))
}

func (h *EquipmentHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.SerialNumber = strings.TrimSpace(in.SerialNumber)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	e, err := h.app.Update(ctx, &domain.Equipment{
		ID:              id,
		Name:            in.Name,
		Description:     in.Description,
		SerialNumber:    in.SerialNumber,
		EquipmentTypeID: in.EquipmentTypeID,
		DepartmentID:    in.DepartmentID,
		ResponsibleID:   in.ResponsibleID,
		Status:          domain.EquipmentStatus(in.Status),
	})
	if err != nil {
		log.Error("update failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("equipment updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toEquipmentResponse(*e))
}

func (h *EquipmentHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment", "op", "Delete")
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

	log.Info("equipment deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	w.WriteHeader(http.StatusNoContent)
}

func toEquipmentResponse(e domain.Equipment) dto.EquipmentResponse {
	return dto.EquipmentResponse{
		ID:              e.ID,
		Name:            e.Name,
		Description:     e.Description,
		SerialNumber:    e.SerialNumber,
		EquipmentTypeID: e.EquipmentTypeID,
		DepartmentID:    e.DepartmentID,
		ResponsibleID:   e.ResponsibleID,
		Status:          string(e.Status),
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
