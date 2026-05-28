package app

import (
	"context"
	"log/slog"

	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/authctx"
	"atom-maintenance/platform/logger"
)

var validTransitions = map[domain.WorkOrderStatus][]domain.WorkOrderStatus{
	domain.WorkOrderStatusOpen:       {domain.WorkOrderStatusInProgress, domain.WorkOrderStatusCancelled},
	domain.WorkOrderStatusInProgress: {domain.WorkOrderStatusOpen, domain.WorkOrderStatusCompleted, domain.WorkOrderStatusCancelled},
	domain.WorkOrderStatusCompleted:  {},
	domain.WorkOrderStatusCancelled:  {},
}

func isTransitionAllowed(from, to domain.WorkOrderStatus) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

type WorkOrderApp struct {
	repo          domain.WorkOrderRepository
	eqRepo        domain.EquipmentRepository
	statusLogRepo domain.WorkOrderStatusLogRepository
	checklistRepo domain.ChecklistItemRepository
	log           *slog.Logger
}

func NewWorkOrderApp(
	repo domain.WorkOrderRepository,
	eqRepo domain.EquipmentRepository,
	statusLogRepo domain.WorkOrderStatusLogRepository,
	checklistRepo domain.ChecklistItemRepository,
	log *slog.Logger,
) *WorkOrderApp {
	return &WorkOrderApp{
		repo:          repo,
		eqRepo:        eqRepo,
		statusLogRepo: statusLogRepo,
		checklistRepo: checklistRepo,
		log:           log,
	}
}

func (a *WorkOrderApp) Create(ctx context.Context, w *domain.WorkOrder) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.work_order", "op", "Create")

	if eq, err := a.eqRepo.GetByID(ctx, w.EquipmentID); err == nil {
		if eq.Status == domain.StatusDecommissioned {
			log.Warn("equipment decommissioned", "equipment_id", w.EquipmentID)
			return nil, domain.ErrBadRequest
		}
		if w.AssignedTo == nil && eq.ResponsibleID != nil {
			w.AssignedTo = eq.ResponsibleID
		}
	}

	out, err := a.repo.Create(ctx, w)
	if err != nil {
		log.Error("create failed", "err", err)
		return nil, err
	}

	// Если наряд сразу создаётся в работе — переводим оборудование на ТО
	if out.Status == domain.WorkOrderStatusInProgress {
		a.syncEquipmentStatus(ctx, log, domain.WorkOrderStatusOpen, domain.WorkOrderStatusInProgress, out.EquipmentID)
	}

	return out, nil
}

func (a *WorkOrderApp) GetByID(ctx context.Context, id int64) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.work_order", "op", "GetByID")

	out, err := a.repo.GetByID(ctx, id)
	if err != nil {
		log.Warn("get failed", "id", id, "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderApp) List(ctx context.Context) ([]domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.work_order", "op", "List")

	out, err := a.repo.List(ctx)
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, err
	}
	return out, nil
}

func (a *WorkOrderApp) Update(ctx context.Context, w *domain.WorkOrder) (*domain.WorkOrder, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.work_order", "op", "Update")

	current, err := a.repo.GetByID(ctx, w.ID)
	if err != nil {
		log.Warn("get current failed", "id", w.ID, "err", err)
		return nil, err
	}
	statusChanged := current.Status != w.Status

	if statusChanged && !isTransitionAllowed(current.Status, w.Status) {
		log.Warn("invalid status transition", "from", current.Status, "to", w.Status)
		return nil, domain.ErrBadRequest
	}

	if statusChanged && w.Status == domain.WorkOrderStatusCompleted {
		items, err := a.checklistRepo.ListByWorkOrderID(ctx, w.ID)
		if err != nil {
			log.Error("checklist check failed", "id", w.ID, "err", err)
			return nil, err
		}
		for _, item := range items {
			if !item.Checked {
				return nil, domain.ErrChecklistIncomplete
			}
		}
	}

	out, err := a.repo.Update(ctx, w)
	if err != nil {
		log.Error("update failed", "id", w.ID, "err", err)
		return nil, err
	}

	if statusChanged {
		var changedBy *int64
		if p, ok := authctx.PrincipalFrom(ctx); ok {
			changedBy = &p.EmployeeID
		}
		if _, err := a.statusLogRepo.Create(ctx, &domain.WorkOrderStatusLog{
			WorkOrderID: w.ID,
			FromStatus:  current.Status,
			ToStatus:    w.Status,
			ChangedBy:   changedBy,
		}); err != nil {
			log.Warn("status log create failed", "id", w.ID, "err", err)
		}

		a.syncEquipmentStatus(ctx, log, current.Status, w.Status, w.EquipmentID)
	}

	return out, nil
}

func (a *WorkOrderApp) Delete(ctx context.Context, id int64) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.work_order", "op", "Delete")

	if err := a.repo.Delete(ctx, id); err != nil {
		log.Error("delete failed", "id", id, "err", err)
		return err
	}
	return nil
}

// syncEquipmentStatus обновляет статус оборудования при смене статуса наряда.
// Не блокирует основной поток — ошибки только логируются.
func (a *WorkOrderApp) syncEquipmentStatus(ctx context.Context, log *slog.Logger, prevStatus, newStatus domain.WorkOrderStatus, equipmentID int64) {
	var targetStatus domain.EquipmentStatus
	switch {
	case newStatus == domain.WorkOrderStatusInProgress:
		targetStatus = domain.StatusUnderMaintenance
	case prevStatus == domain.WorkOrderStatusInProgress:
		// Выходим из «в работе» → освобождаем оборудование
		targetStatus = domain.StatusActive
	default:
		return
	}

	eq, err := a.eqRepo.GetByID(ctx, equipmentID)
	if err != nil {
		log.Warn("syncEquipmentStatus: get failed", "equipment_id", equipmentID, "err", err)
		return
	}
	eq.Status = targetStatus
	if _, err := a.eqRepo.Update(ctx, eq); err != nil {
		log.Warn("syncEquipmentStatus: update failed", "equipment_id", equipmentID, "err", err)
	}
}
