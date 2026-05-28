package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/authctx"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"

	"github.com/go-playground/validator/v10"
)

type WorkOrderHandlers struct {
	app      *app.WorkOrderApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewWorkOrderHandlers(app *app.WorkOrderApp, log *slog.Logger) *WorkOrderHandlers {
	return &WorkOrderHandlers{
		app:      app,
		log:      log,
		validate: validator.New(),
	}
}

func (h *WorkOrderHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.work_order", "op", "Create")
	start := time.Now()

	principal, ok := authctx.PrincipalFrom(ctx)
	if !ok {
		respond.Error(w, domain.ErrUnauthorized)
		return
	}

	var in dto.CreateWorkOrderRequest
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

	status := domain.WorkOrderStatusOpen
	if in.Status != "" {
		status = domain.WorkOrderStatus(in.Status)
	}
	workType := domain.WorkOrderTypeCorrective
	if in.WorkType != "" {
		workType = domain.WorkOrderType(in.WorkType)
	}

	createdByID := principal.EmployeeID
	wo, err := h.app.Create(ctx, &domain.WorkOrder{
		ScheduleID:  in.ScheduleID,
		EquipmentID: in.EquipmentID,
		Title:       in.Title,
		Description: in.Description,
		AssignedTo:  in.AssignedTo,
		CreatedBy:   &createdByID,
		Status:      status,
		WorkType:    workType,
	})
	if err != nil {
		log.Error("create failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("work order created",
		"id", wo.ID,
		"duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusCreated, toWorkOrderResponse(*wo))
}

func (h *WorkOrderHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.work_order", "op", "List")
	start := time.Now()

	items, err := h.app.List(ctx)
	if err != nil {
		log.Error("list failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.WorkOrderResponse, 0, len(items))
	for _, wo := range items {
		out = append(out, toWorkOrderResponse(wo))
	}

	log.Info("work orders listed",
		"count", len(out),
		"duration_ms", time.Since(start).Milliseconds())

	respond.JSON(w, http.StatusOK, out)
}

func (h *WorkOrderHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.work_order", "op", "GetByID")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	wo, err := h.app.GetByID(ctx, id)
	if err != nil {
		log.Warn("not found",
			"id", id,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("work order fetched",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, toWorkOrderResponse(*wo))
}

func (h *WorkOrderHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.work_order", "op", "Update")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.UpdateWorkOrderRequest
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

	updateWorkType := domain.WorkOrderTypeCorrective
	if in.WorkType != "" {
		updateWorkType = domain.WorkOrderType(in.WorkType)
	}

	wo, err := h.app.Update(ctx, &domain.WorkOrder{
		ID:          id,
		ScheduleID:  in.ScheduleID,
		EquipmentID: in.EquipmentID,
		Title:       in.Title,
		Description: in.Description,
		AssignedTo:  in.AssignedTo,
		Status:      domain.WorkOrderStatus(in.Status),
		WorkType:    updateWorkType,
		CompletedAt: in.CompletedAt,
	})
	if err != nil {
		log.Error("update failed",
			"id", id,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("work order updated",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, toWorkOrderResponse(*wo))
}

func (h *WorkOrderHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.work_order", "op", "Delete")
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

	log.Info("work order deleted",
		"id", id,
		"duration_ms", time.Since(start).Milliseconds())
	w.WriteHeader(http.StatusNoContent)
}

func toWorkOrderResponse(w domain.WorkOrder) dto.WorkOrderResponse {
	return dto.WorkOrderResponse{
		ID:          w.ID,
		ScheduleID:  w.ScheduleID,
		EquipmentID: w.EquipmentID,
		Title:       w.Title,
		Description: w.Description,
		AssignedTo:  w.AssignedTo,
		CreatedBy:   w.CreatedBy,
		Status:      string(w.Status),
		WorkType:    string(w.WorkType),
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
		CompletedAt: w.CompletedAt,
	}
}
