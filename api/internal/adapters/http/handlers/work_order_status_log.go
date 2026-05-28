package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"
)

type StatusLogHandlers struct {
	repo domain.WorkOrderStatusLogRepository
	log  *slog.Logger
}

func NewStatusLogHandlers(repo domain.WorkOrderStatusLogRepository, log *slog.Logger) *StatusLogHandlers {
	return &StatusLogHandlers{repo: repo, log: log}
}

func (h *StatusLogHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.status_log", "op", "List")
	start := time.Now()

	woID, err := parseID(r, "id", log)
	if err != nil {
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	items, err := h.repo.ListByWorkOrderID(ctx, woID)
	if err != nil {
		log.Error("list failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	out := make([]dto.StatusLogResponse, 0, len(items))
	for _, l := range items {
		out = append(out, dto.StatusLogResponse{
			ID:          l.ID,
			WorkOrderID: l.WorkOrderID,
			FromStatus:  string(l.FromStatus),
			ToStatus:    string(l.ToStatus),
			ChangedBy:   l.ChangedBy,
			Comment:     l.Comment,
			CreatedAt:   l.CreatedAt,
		})
	}

	log.Info("status logs listed", "work_order_id", woID, "count", len(out), "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, out)
}
