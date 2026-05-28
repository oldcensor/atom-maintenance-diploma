package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type DepartmentHandlers struct {
	app      *app.DepartmentApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewDepartmentHandlers(app *app.DepartmentApp, log *slog.Logger) *DepartmentHandlers {
	return &DepartmentHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *DepartmentHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.department", "op", "Create")
	start := time.Now()

	var in dto.CreateDepartmentRequest
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

	d, err := h.app.Create(ctx, &domain.Department{
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

	log.Info("department created",
		"id", d.ID,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusCreated, toDepartmentResponse(*d))
}

func (h *DepartmentHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.department", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.DepartmentResponse, 0, len(items))
	for _, d := range items {
		out = append(out, toDepartmentResponse(d))
	}

	log.Info("departments listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *DepartmentHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.department", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	d, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("department fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toDepartmentResponse(*d))
}

func (h *DepartmentHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.department", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateDepartmentRequest
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

	d, err := h.app.Update(ctx, &domain.Department{
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

	log.Info("department updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toDepartmentResponse(*d))
}

func (h *DepartmentHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.department", "op", "Delete")
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

	log.Info("department deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	w.WriteHeader(http.StatusNoContent)
}

func toDepartmentResponse(d domain.Department) dto.DepartmentResponse {
	return dto.DepartmentResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
	}
}

func parseID(r *http.Request, param string, log *slog.Logger) (int64, error) {
	val := chi.URLParam(r, param)
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Warn("invalid url param", "param", param, "value", val, "err", err)
	}
	return id, err
}
