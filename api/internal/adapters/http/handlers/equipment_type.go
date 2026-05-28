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

type EquipmentTypeHandlers struct {
	app      *app.EquipmentTypeApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewEquipmentTypeHandlers(app *app.EquipmentTypeApp, log *slog.Logger) *EquipmentTypeHandlers {
	return &EquipmentTypeHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *EquipmentTypeHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment_type", "op", "Create")
	start := time.Now()

	var in dto.CreateEquipmentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	et, err := h.app.Create(ctx, &domain.EquipmentType{
		Name:        in.Name,
		Description: strings.TrimSpace(in.Description),
	})
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("equipment_type created",
		"id", et.ID,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusCreated, toEquipmentTypeResponse(*et))
}

func (h *EquipmentTypeHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment_type", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.EquipmentTypeResponse, 0, len(items))
	for _, et := range items {
		out = append(out, toEquipmentTypeResponse(et))
	}

	log.Info("equipment_types listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *EquipmentTypeHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment_type", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	et, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("equipment_type fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toEquipmentTypeResponse(*et))
}

func (h *EquipmentTypeHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment_type", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateEquipmentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	et, err := h.app.Update(ctx, &domain.EquipmentType{
		ID:          id,
		Name:        in.Name,
		Description: strings.TrimSpace(in.Description),
	})
	if err != nil {
		log.Error("update failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("equipment_type updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toEquipmentTypeResponse(*et))
}

func (h *EquipmentTypeHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.equipment_type", "op", "Delete")
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

	log.Info("equipment_type deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	w.WriteHeader(http.StatusNoContent)
}

func toEquipmentTypeResponse(et domain.EquipmentType) dto.EquipmentTypeResponse {
	return dto.EquipmentTypeResponse{
		ID:          et.ID,
		Name:        et.Name,
		Description: et.Description,
		CreatedAt:   et.CreatedAt,
	}
}
