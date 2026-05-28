package integration

import (
	"fmt"
	"net/http"
	"testing"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func createWorkOrder(t *testing.T, env *common.Env, eqID int64, title string) int64 {
	t.Helper()
	var out dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/work-orders",
		dto.CreateWorkOrderRequest{EquipmentID: eqID, Title: title, Status: "open"},
		&out, http.StatusCreated)
	return out.ID
}

func TestWorkOrderCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Насос ВО")
	eqID := createEquipment(t, env, "ГЦН-ВО", "SN-WO1", etID)

	in := dto.CreateWorkOrderRequest{
		EquipmentID: eqID,
		Title:       "Замена подшипника",
		Description: "Плановая замена",
		Status:      "open",
	}
	var out dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/work-orders", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.Title != in.Title {
		t.Fatalf("title: got %q, want %q", out.Title, in.Title)
	}
	if out.Status != "open" {
		t.Fatalf("status: got %q, want %q", out.Status, "open")
	}
}

func TestWorkOrderCreate_InvalidEquipment(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateWorkOrderRequest{EquipmentID: 99999, Title: "Нет оборудования", Status: "open"}
	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodPost, env.Base+"/work-orders", in, nil, http.StatusBadRequest)
}

func TestWorkOrderCreate_DecommissionedEquipment(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ВО Списан")
	var eq dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: "Списанное", SerialNumber: "SN-WO-DEC", EquipmentTypeID: etID, Status: "decommissioned"},
		&eq, http.StatusCreated)

	// Наряд на списанном оборудовании создавать нельзя
	in := dto.CreateWorkOrderRequest{EquipmentID: eq.ID, Title: "На списанном", Status: "open"}
	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodPost, env.Base+"/work-orders", in, nil, http.StatusBadRequest)
}

func TestWorkOrderList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Компрессор ВО")
	eqID := createEquipment(t, env, "КМ-ВО", "SN-KM-WO", etID)

	for i := range 3 {
		createWorkOrder(t, env, eqID, fmt.Sprintf("Работа %d", i))
	}

	var list []dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/work-orders", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestWorkOrderGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Генератор ВО")
	eqID := createEquipment(t, env, "ДГ-ВО", "SN-DG-WO", etID)
	woID := createWorkOrder(t, env, eqID, "Техосмотр")

	var got dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/work-orders/%d", env.Base, woID), nil, &got, http.StatusOK)

	if got.ID != woID {
		t.Fatalf("ID: got %d, want %d", got.ID, woID)
	}
}

func TestWorkOrderGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodGet, env.Base+"/work-orders/99999", nil, nil, http.StatusNotFound)
}

func TestWorkOrderUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Кран ВО")
	eqID := createEquipment(t, env, "КР-ВО", "SN-KR-WO", etID)
	woID := createWorkOrder(t, env, eqID, "Первоначальное задание")

	upd := dto.UpdateWorkOrderRequest{
		EquipmentID: eqID,
		Title:       "Обновлённое задание",
		Status:      "in_progress",
	}
	var updated dto.WorkOrderResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/work-orders/%d", env.Base, woID),
		upd, &updated, http.StatusOK)

	if updated.Title != upd.Title {
		t.Fatalf("title: got %q, want %q", updated.Title, upd.Title)
	}
	if updated.Status != upd.Status {
		t.Fatalf("status: got %q, want %q", updated.Status, upd.Status)
	}
}

func TestWorkOrderUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ВО")
	eqID := createEquipment(t, env, "Оборуд ВО", "SN-WO-X", etID)

	upd := dto.UpdateWorkOrderRequest{EquipmentID: eqID, Title: "Нет", Status: "open"}
	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodPut, env.Base+"/work-orders/99999",
		upd, nil, http.StatusNotFound)
}

func TestWorkOrderDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Удалить ВО тип")
	eqID := createEquipment(t, env, "Удалить ВО", "SN-DEL-WO", etID)
	woID := createWorkOrder(t, env, eqID, "Удалить меня")

	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/work-orders/%d", env.Base, woID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/work-orders/%d", env.Base, woID), nil, nil, http.StatusNotFound)
}

func TestWorkOrderDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.WorkOrderResponse](t, env.Client, http.MethodDelete,
		env.Base+"/work-orders/99999", nil, nil, http.StatusNotFound)
}
