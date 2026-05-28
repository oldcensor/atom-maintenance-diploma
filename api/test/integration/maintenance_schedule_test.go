package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func createEquipment(t *testing.T, env *common.Env, name, serial string, etID int64) int64 {
	t.Helper()
	var out dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: name, SerialNumber: serial, EquipmentTypeID: etID},
		&out, http.StatusCreated)
	return out.ID
}

func TestMaintenanceScheduleCreate_DecommissionedEquipment(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Тип МС Списан")
	var eq dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: "Списанное МС", SerialNumber: "SN-MS-DEC", EquipmentTypeID: etID, Status: "decommissioned"},
		&eq, http.StatusCreated)

	// Регламент для списанного оборудования создавать нельзя
	in := dto.CreateMaintenanceScheduleRequest{
		EquipmentID: eq.ID,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Description: "ТО списанного",
		Status:      "scheduled",
	}
	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules", in, nil, http.StatusBadRequest)
}

func TestMaintenanceScheduleCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Насос МС")
	eqID := createEquipment(t, env, "ГЦН-МС", "SN-MS1", etID)

	in := dto.CreateMaintenanceScheduleRequest{
		EquipmentID: eqID,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Description: "Плановое ТО",
		Status:      "scheduled",
	}
	var out dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.EquipmentID != eqID {
		t.Fatalf("equipment_id: got %d, want %d", out.EquipmentID, eqID)
	}
	if out.Status != "scheduled" {
		t.Fatalf("status: got %q, want %q", out.Status, "scheduled")
	}
}

func TestMaintenanceScheduleList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Компрессор МС")
	eqID := createEquipment(t, env, "КМ-МС", "SN-KM-MS", etID)

	for i := range 3 {
		in := dto.CreateMaintenanceScheduleRequest{
			EquipmentID: eqID,
			ScheduledAt: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
		}
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules", in, &dto.MaintenanceScheduleResponse{}, http.StatusCreated)
	}

	var list []dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/maintenance-schedules", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestMaintenanceScheduleGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Генератор МС")
	eqID := createEquipment(t, env, "ДГ-МС", "SN-DG-MS", etID)

	var created dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules",
		dto.CreateMaintenanceScheduleRequest{EquipmentID: eqID, ScheduledAt: time.Now().Add(48 * time.Hour)},
		&created, http.StatusCreated)

	var got dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestMaintenanceScheduleGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodGet,
		env.Base+"/maintenance-schedules/99999", nil, nil, http.StatusNotFound)
}

func TestMaintenanceScheduleUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Турбина МС")
	eqID := createEquipment(t, env, "ТГ-МС", "SN-TG-MS", etID)

	var created dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules",
		dto.CreateMaintenanceScheduleRequest{EquipmentID: eqID, ScheduledAt: time.Now().Add(24 * time.Hour)},
		&created, http.StatusCreated)

	upd := dto.UpdateMaintenanceScheduleRequest{
		EquipmentID: eqID,
		ScheduledAt: time.Now().Add(72 * time.Hour),
		Description: "Внеплановое ТО",
		Status:      "in_progress",
	}
	var updated dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.Status != upd.Status {
		t.Fatalf("status: got %q, want %q", updated.Status, upd.Status)
	}
	if updated.Description != upd.Description {
		t.Fatalf("description: got %q, want %q", updated.Description, upd.Description)
	}
}

func TestMaintenanceScheduleUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Тип")
	eqID := createEquipment(t, env, "Оборуд", "SN-X1", etID)

	upd := dto.UpdateMaintenanceScheduleRequest{EquipmentID: eqID, ScheduledAt: time.Now(), Status: "scheduled"}
	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodPut,
		env.Base+"/maintenance-schedules/99999", upd, nil, http.StatusNotFound)
}

func TestMaintenanceScheduleDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Удалить МС тип")
	eqID := createEquipment(t, env, "Удалить МС", "SN-DEL-MS", etID)

	var created dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules",
		dto.CreateMaintenanceScheduleRequest{EquipmentID: eqID, ScheduledAt: time.Now().Add(24 * time.Hour)},
		&created, http.StatusCreated)

	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, created.ID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, created.ID), nil, nil, http.StatusNotFound)
}

func TestMaintenanceScheduleDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.MaintenanceScheduleResponse](t, env.Client, http.MethodDelete,
		env.Base+"/maintenance-schedules/99999", nil, nil, http.StatusNotFound)
}

func TestMaintenanceScheduleCreate_WithInterval(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Насос ИН")
	eqID := createEquipment(t, env, "ГЦН-ИН", "SN-IN1", etID)

	intervalUnit := "operating_hours"
	intervalValue := 500
	in := dto.CreateMaintenanceScheduleRequest{
		EquipmentID:   eqID,
		ScheduledAt:   time.Now().Add(24 * time.Hour),
		Description:   "Плановое ТО с интервалом",
		Status:        "scheduled",
		IntervalUnit:  &intervalUnit,
		IntervalValue: &intervalValue,
	}
	var out dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules", in, &out, http.StatusCreated)

	if out.IntervalUnit == nil || *out.IntervalUnit != intervalUnit {
		t.Fatalf("interval_unit: got %v, want %q", out.IntervalUnit, intervalUnit)
	}
	if out.IntervalValue == nil || *out.IntervalValue != intervalValue {
		t.Fatalf("interval_value: got %v, want %d", out.IntervalValue, intervalValue)
	}
}

func TestMaintenanceScheduleUpdate_WithInterval(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Кран ИН")
	eqID := createEquipment(t, env, "КР-ИН", "SN-IN2", etID)

	var created dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/maintenance-schedules",
		dto.CreateMaintenanceScheduleRequest{EquipmentID: eqID, ScheduledAt: time.Now().Add(24 * time.Hour)},
		&created, http.StatusCreated)

	intervalUnit := "cycles"
	intervalValue := 200
	upd := dto.UpdateMaintenanceScheduleRequest{
		EquipmentID:   eqID,
		ScheduledAt:   created.ScheduledAt,
		Status:        "scheduled",
		IntervalUnit:  &intervalUnit,
		IntervalValue: &intervalValue,
	}
	var updated dto.MaintenanceScheduleResponse
	common.DoJSON(t, env.Client, http.MethodPut,
		fmt.Sprintf("%s/maintenance-schedules/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.IntervalUnit == nil || *updated.IntervalUnit != intervalUnit {
		t.Fatalf("interval_unit: got %v, want %q", updated.IntervalUnit, intervalUnit)
	}
	if updated.IntervalValue == nil || *updated.IntervalValue != intervalValue {
		t.Fatalf("interval_value: got %v, want %d", updated.IntervalValue, intervalValue)
	}
}
