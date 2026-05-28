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

type WorkOrderCommentHandlers struct {
	app      *app.WorkOrderCommentApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewWorkOrderCommentHandlers(app *app.WorkOrderCommentApp, log *slog.Logger) *WorkOrderCommentHandlers {
	return &WorkOrderCommentHandlers{app: app, log: log, validate: validator.New()}
}

func (h *WorkOrderCommentHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_comment", "op", "Create")
	start := time.Now()

	principal, ok := authctx.PrincipalFrom(ctx)
	if !ok {
		respond.Error(w, domain.ErrUnauthorized)
		return
	}

	woID, err := parseID(r, "woID", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	var in dto.CreateCommentRequest
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

	c, err := h.app.Create(ctx, &domain.WorkOrderComment{
		WorkOrderID: woID,
		AuthorID:    principal.EmployeeID,
		Text:        in.Text,
	})
	if err != nil {
		log.Error("create failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("comment created", "id", c.ID, "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusCreated, toCommentResponse(*c))
}

func (h *WorkOrderCommentHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_comment", "op", "List")
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

	out := make([]dto.CommentResponse, 0, len(items))
	for _, c := range items {
		out = append(out, toCommentResponse(c))
	}

	log.Info("comments listed", "work_order_id", woID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, out)
}

func (h *WorkOrderCommentHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.wo_comment", "op", "Delete")
	start := time.Now()

	id, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	if err := h.app.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("comment deleted", "id", id, "duration_ms", time.Since(start).Milliseconds())
	w.WriteHeader(http.StatusNoContent)
}

func toCommentResponse(c domain.WorkOrderComment) dto.CommentResponse {
	return dto.CommentResponse{
		ID:          c.ID,
		WorkOrderID: c.WorkOrderID,
		AuthorID:    c.AuthorID,
		Text:        c.Text,
		CreatedAt:   c.CreatedAt,
	}
}
