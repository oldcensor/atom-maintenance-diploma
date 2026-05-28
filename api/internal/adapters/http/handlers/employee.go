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

type EmployeeHandlers struct {
	app      *app.EmployeeApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewEmployeeHandlers(app *app.EmployeeApp, log *slog.Logger) *EmployeeHandlers {
	return &EmployeeHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *EmployeeHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.employee", "op", "Create")
	start := time.Now()

	var in dto.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.FullName = strings.TrimSpace(in.FullName)
	in.Email = strings.TrimSpace(in.Email)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	e, err := h.app.Create(ctx, &domain.Employee{
		Email:        in.Email,
		FullName:     in.FullName,
		Role:         domain.EmployeeRole(in.Role),
		DepartmentID: in.DepartmentID,
	}, in.Password)
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("employee created",
		"id", e.ID,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusCreated, toEmployeeResponse(*e))
}

func (h *EmployeeHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.employee", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.EmployeeResponse, 0, len(items))
	for _, e := range items {
		out = append(out, toEmployeeResponse(e))
	}

	log.Info("employees listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *EmployeeHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.employee", "op", "GetByID")
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

	log.Info("employee fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toEmployeeResponse(*e))
}

func (h *EmployeeHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.employee", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	in.FullName = strings.TrimSpace(in.FullName)
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	e, err := h.app.Update(ctx, &domain.Employee{
		ID:           id,
		FullName:     in.FullName,
		Role:         domain.EmployeeRole(in.Role),
		DepartmentID: in.DepartmentID,
	})
	if err != nil {
		log.Error("update failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("employee updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, toEmployeeResponse(*e))
}

func (h *EmployeeHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.employee", "op", "Delete")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	if err := h.app.SoftDelete(ctx, id); err != nil {
		log.Error("delete failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("employee deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())

	w.WriteHeader(http.StatusNoContent)
}

func toEmployeeResponse(e domain.Employee) dto.EmployeeResponse {
	return dto.EmployeeResponse{
		ID:           e.ID,
		Email:        e.Email,
		FullName:     e.FullName,
		Role:         string(e.Role),
		DepartmentID: e.DepartmentID,
		CreatedAt:    e.CreatedAt,
	}
}
