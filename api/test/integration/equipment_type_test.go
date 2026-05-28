package integration

import (
	"fmt"
	"net/http"
	"testing"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func TestEquipmentTypeCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateEquipmentTypeRequest{Name: "Насос", Description: "Центробежные насосы"}
	var out dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.Name != in.Name {
		t.Fatalf("name: got %q, want %q", out.Name, in.Name)
	}
}

func TestEquipmentTypeList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	for i := range 3 {
		in := dto.CreateEquipmentTypeRequest{Name: fmt.Sprintf("Тип %d", i)}
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types", in, &dto.EquipmentTypeResponse{}, http.StatusCreated)
	}

	var list []dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/equipment-types", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestEquipmentTypeGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateEquipmentTypeRequest{Name: "Турбина"}
	var created dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types", in, &created, http.StatusCreated)

	var got dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/equipment-types/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestEquipmentTypeGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EquipmentTypeResponse](t, env.Client, http.MethodGet, env.Base+"/equipment-types/99999", nil, nil, http.StatusNotFound)
}

func TestEquipmentTypeUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types",
		dto.CreateEquipmentTypeRequest{Name: "Старый тип"}, &created, http.StatusCreated)

	upd := dto.UpdateEquipmentTypeRequest{Name: "Новый тип", Description: "Обновлено"}
	var updated dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/equipment-types/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.Name != upd.Name {
		t.Fatalf("name: got %q, want %q", updated.Name, upd.Name)
	}
}

func TestEquipmentTypeUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	upd := dto.UpdateEquipmentTypeRequest{Name: "Не существует"}
	common.DoJSON[dto.EquipmentTypeResponse](t, env.Client, http.MethodPut, env.Base+"/equipment-types/99999",
		upd, nil, http.StatusNotFound)
}

func TestEquipmentTypeDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.EquipmentTypeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/equipment-types",
		dto.CreateEquipmentTypeRequest{Name: "Удалить"}, &created, http.StatusCreated)

	common.DoJSON[dto.EquipmentTypeResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/equipment-types/%d", env.Base, created.ID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.EquipmentTypeResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/equipment-types/%d", env.Base, created.ID), nil, nil, http.StatusNotFound)
}

func TestEquipmentTypeDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EquipmentTypeResponse](t, env.Client, http.MethodDelete,
		env.Base+"/equipment-types/99999", nil, nil, http.StatusNotFound)
}
