package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/adapters/simulator"
	appscheduler "atom-maintenance/platform/scheduler"
	"atom-maintenance/test/common"
)

func createSchedulerSetup(
	t *testing.T,
	env *common.Env,
	description, intervalUnit string,
	intervalValue int,
	scheduledAt time.Time,
) (eqID, schedID int64) {
	t.Helper()

	etID := createEquipmentType(t, env, fmt.Sprintf("Тип %s", description))
	eqID = createEquipment(t, env, description, fmt.Sprintf("SN-%s-%d", description, time.Now().UnixNano()), etID)

	req := dto.CreateMaintenanceScheduleRequest{
		EquipmentID: eqID,
		ScheduledAt: scheduledAt,
		Description: description,
		Status:      "scheduled",
	}
	if intervalUnit != "" {
		req.IntervalUnit = &intervalUnit
		req.IntervalValue = &intervalValue
	}

	var out dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules", req, &out, http.StatusCreated)
	schedID = out.ID
	return
}

func newSched(env *common.Env, simClient *simulator.Client) *appscheduler.Scheduler {
	return appscheduler.New(time.Minute, simClient, env.MaintenanceSchedule, env.WorkOrder, env.TxManager, env.Log)
}

func countWorkOrders(t *testing.T, env *common.Env) int {
	t.Helper()
	var list []dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/work-orders", nil, &list, http.StatusOK)
	return len(list)
}

func TestScheduler_DueByOperatingHours(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, _ := createSchedulerSetup(t, env, "ГЦН-Ч", "operating_hours", 100, time.Now().Add(-time.Hour))

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 5100},
	})
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order, got %d", n)
	}
}

func TestScheduler_DueByDays_FirstRun(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	createSchedulerSetup(t, env, "ТО-Дни-1", "days", 30, time.Now().Add(-48*time.Hour))

	simClient := common.FakeSimulator(t, nil)
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order, got %d", n)
	}
}

func TestScheduler_DueByDays_NextDue(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	_, schedID := createSchedulerSetup(t, env, "ТО-Дни-2", "days", 30, time.Now().Add(-48*time.Hour))

	yesterday := time.Now().AddDate(0, 0, -1)
	if err := env.DB.Exec("UPDATE maintenance_schedule SET next_due_at = ? WHERE id = ?", yesterday, schedID).Error; err != nil {
		t.Fatalf("set next_due_at: %v", err)
	}

	simClient := common.FakeSimulator(t, nil)
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order, got %d", n)
	}
}

func TestScheduler_NotDue_MeterBelowThreshold(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, schedID := createSchedulerSetup(t, env, "КМ-Ч", "operating_hours", 100, time.Now().Add(-time.Hour))

	if err := env.DB.Exec("UPDATE maintenance_schedule SET last_meter_value = 5000 WHERE id = ?", schedID).Error; err != nil {
		t.Fatalf("set last_meter_value: %v", err)
	}

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 5050},
	})
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 0 {
		t.Fatalf("expected 0 work orders, got %d", n)
	}
}

func TestScheduler_NoDuplicate_OpenWorkOrder(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, _ := createSchedulerSetup(t, env, "ДГ-Ч", "operating_hours", 100, time.Now().Add(-time.Hour))

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 9999},
	})
	sched := newSched(env, simClient)

	sched.Tick(context.Background()) // creates WO
	sched.Tick(context.Background()) // should skip

	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order after 2 ticks, got %d", n)
	}
}

func TestScheduler_NoDuplicate_InProgressWorkOrder(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, _ := createSchedulerSetup(t, env, "ГЦН-3", "operating_hours", 100, time.Now().Add(-time.Hour))

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 9999},
	})
	sched := newSched(env, simClient)

	sched.Tick(context.Background()) // creates WO

	// move WO to in_progress
	var list []dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/work-orders", nil, &list, http.StatusOK)
	if len(list) != 1 {
		t.Fatalf("expected 1 work order, got %d", len(list))
	}
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/work-orders/%d", env.Base, list[0].ID),
		dto.UpdateWorkOrderRequest{EquipmentID: eqID, Title: list[0].Title, Status: "in_progress"},
		&dto.WorkOrderResponse{}, http.StatusOK)

	sched.Tick(context.Background()) // should skip

	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order after 2 ticks, got %d", n)
	}
}

func TestScheduler_Atomic_MeterFieldsUpdated(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, schedID := createSchedulerSetup(t, env, "ТО-Атомик", "operating_hours", 100, time.Now().Add(-time.Hour))

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 5100},
	})
	newSched(env, simClient).Tick(context.Background())

	// verify WO created
	if n := countWorkOrders(t, env); n != 1 {
		t.Fatalf("expected 1 work order, got %d", n)
	}

	// verify meter fields updated on schedule
	var sched dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, schedID), nil, &sched, http.StatusOK)

	if sched.LastMeterValue == nil {
		t.Fatal("expected last_meter_value to be set")
	}
	if *sched.LastMeterValue != 5100 {
		t.Fatalf("last_meter_value: got %v, want 5100", *sched.LastMeterValue)
	}
	if sched.NextDueAt == nil {
		t.Fatal("expected next_due_at to be set")
	}
}

func TestScheduler_SkipsInactiveStatus(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	eqID, schedID := createSchedulerSetup(t, env, "ТО-Стоп", "operating_hours", 100, time.Now().Add(-time.Hour))

	if err := env.DB.Exec("UPDATE maintenance_schedule SET status = 'in_progress' WHERE id = ?", schedID).Error; err != nil {
		t.Fatalf("set status: %v", err)
	}

	simClient := common.FakeSimulator(t, []simulator.TelemetryItem{
		{EquipmentID: eqID, MeterType: "operating_hours", CurrentValue: 9999},
	})
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 0 {
		t.Fatalf("expected 0 work orders, got %d", n)
	}
}

func TestScheduler_SkipsNoIntervalUnit(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	// pass empty intervalUnit so no interval fields are set
	createSchedulerSetup(t, env, "ТО-Без-интервала", "", 0, time.Now().Add(-time.Hour))

	simClient := common.FakeSimulator(t, nil)
	newSched(env, simClient).Tick(context.Background())

	if n := countWorkOrders(t, env); n != 0 {
		t.Fatalf("expected 0 work orders, got %d", n)
	}
}
