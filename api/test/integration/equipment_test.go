package integration

import (
	"fmt"
	"net/http"
	"testing"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func createEquipmentType(t *testing.T, env *common.Env, name string) int64 {
	t.Helper()
	var out dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types",
		dto.CreateEquipmentTypeRequest{Name: name}, &out, http.StatusCreated)
	return out.ID
}

func TestEquipmentCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Насос")

	in := dto.CreateEquipmentRequest{
		Name:            "ГЦН-1",
		SerialNumber:    "SN-001",
		EquipmentTypeID: etID,
		Status:          "active",
	}
	var out dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.Name != in.Name {
		t.Fatalf("name: got %q, want %q", out.Name, in.Name)
	}
	if out.EquipmentTypeID != etID {
		t.Fatalf("equipment_type_id: got %d, want %d", out.EquipmentTypeID, etID)
	}
	if out.Status != "active" {
		t.Fatalf("status: got %q, want %q", out.Status, "active")
	}
}

func TestEquipmentCreate_InvalidEquipmentType(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateEquipmentRequest{
		Name:            "Фантомный насос",
		SerialNumber:    "SN-X",
		EquipmentTypeID: 99999,
	}
	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodPost, env.Base+"/equipment", in, nil, http.StatusBadRequest)
}

func TestEquipmentList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Турбина")
	for i := range 3 {
		in := dto.CreateEquipmentRequest{
			Name:            fmt.Sprintf("Оборудование %d", i),
			SerialNumber:    fmt.Sprintf("SN-%03d", i),
			EquipmentTypeID: etID,
		}
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment", in, &dto.EquipmentResponse{}, http.StatusCreated)
	}

	var list []dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/equipment", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestEquipmentGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Компрессор")
	var created dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: "КМ-1", SerialNumber: "SN-KM1", EquipmentTypeID: etID},
		&created, http.StatusCreated)

	var got dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/equipment/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestEquipmentGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodGet, env.Base+"/equipment/99999", nil, nil, http.StatusNotFound)
}

func TestEquipmentUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Генератор")
	var created dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: "ДГ-1", SerialNumber: "SN-DG1", EquipmentTypeID: etID},
		&created, http.StatusCreated)

	upd := dto.UpdateEquipmentRequest{
		Name:            "ДГ-1 обновлён",
		SerialNumber:    "SN-DG1-v2",
		EquipmentTypeID: etID,
		Status:          "under_maintenance",
	}
	var updated dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/equipment/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.Name != upd.Name {
		t.Fatalf("name: got %q, want %q", updated.Name, upd.Name)
	}
	if updated.Status != upd.Status {
		t.Fatalf("status: got %q, want %q", updated.Status, upd.Status)
	}
}

func TestEquipmentUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Тип")
	upd := dto.UpdateEquipmentRequest{Name: "Призрак", SerialNumber: "SN-0", EquipmentTypeID: etID, Status: "active"}
	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodPut, env.Base+"/equipment/99999",
		upd, nil, http.StatusNotFound)
}

func TestEquipmentDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	etID := createEquipmentType(t, env, "Кран")
	var created dto.EquipmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment",
		dto.CreateEquipmentRequest{Name: "КР-7", SerialNumber: "SN-KR7", EquipmentTypeID: etID},
		&created, http.StatusCreated)

	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/equipment/%d", env.Base, created.ID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/equipment/%d", env.Base, created.ID), nil, nil, http.StatusNotFound)
}

func TestEquipmentDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EquipmentResponse](t, env.Client, http.MethodDelete,
		env.Base+"/equipment/99999", nil, nil, http.StatusNotFound)
}
