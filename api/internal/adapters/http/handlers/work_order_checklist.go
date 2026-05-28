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

type ChecklistHandlers struct {
	app      *app.WorkOrderChecklistApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewChecklistHandlers(app *app.WorkOrderChecklistApp, log *slog.Logger) *ChecklistHandlers {
	return &ChecklistHandlers{app: app, log: log, validate: validator.New()}
}

func (h *ChecklistHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_checklist", "op", "Create")
	start := time.Now()

	woID, err := parseID(r, "woID", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.CreateChecklistItemRequest
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

	item, err := h.app.Create(ctx, &domain.ChecklistItem{
		WorkOrderID: woID,
		Text:        in.Text,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		log.Error("create failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("checklist item created", "id", item.ID, "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusCreated, toChecklistResponse(*item))
}

func (h *ChecklistHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_checklist", "op", "List")
	start := time.Now()

	woID, err := parseID(r, "woID", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	items, err := h.app.ListByWorkOrderID(ctx, woID)
	if err != nil {
		log.Error("list failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.ChecklistItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toChecklistResponse(item))
	}

	log.Info("checklist listed", "work_order_id", woID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, out)
}

func (h *ChecklistHandlers) Toggle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_checklist", "op", "Toggle")
	start := time.Now()

	principal, ok := authctx.PrincipalFrom(ctx)
	if !ok {
		respond.Error(w, domain.ErrUnauthorized)
		return
	}

	itemID, err := parseID(r, "itemID", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.ToggleChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	item, err := h.app.Toggle(ctx, itemID, in.Checked, &principal.EmployeeID)
	if err != nil {
		log.Error("toggle failed", "id", itemID, "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("checklist item toggled", "id", itemID, "checked", in.Checked, "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, toChecklistResponse(*item))
}

func (h *ChecklistHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_checklist", "op", "Delete")
	start := time.Now()

	itemID, err := parseID(r, "itemID", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	if err := h.app.Delete(ctx, itemID); err != nil {
		log.Error("delete failed", "id", itemID, "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("checklist item deleted", "id", itemID, "duration_ms", time.Since(start).Milliseconds())
	w.WriteHeader(http.StatusNoContent)
}

func toChecklistResponse(item domain.ChecklistItem) dto.ChecklistItemResponse {
	return dto.ChecklistItemResponse{
		ID:          item.ID,
		WorkOrderID: item.WorkOrderID,
		Text:        item.Text,
		Checked:     item.Checked,
		CheckedBy:   item.CheckedBy,
		CheckedAt:   item.CheckedAt,
		SortOrder:   item.SortOrder,
		CreatedAt:   item.CreatedAt,
	}
}
