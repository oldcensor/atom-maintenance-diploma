package integration

import (
	"fmt"
	"net/http"
	"testing"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func TestDepartmentCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateDepartmentRequest{Name: "Механический цех", Description: "Обслуживание механических систем"}
	var out dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/departments", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.Name != in.Name {
		t.Fatalf("name: got %q, want %q", out.Name, in.Name)
	}
	if out.Description != in.Description {
		t.Fatalf("description: got %q, want %q", out.Description, in.Description)
	}
}

func TestDepartmentList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	for i := range 3 {
		in := dto.CreateDepartmentRequest{Name: fmt.Sprintf("Отдел %d", i)}
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/departments", in, &dto.DepartmentResponse{}, http.StatusCreated)
	}

	var list []dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/departments", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestDepartmentGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateDepartmentRequest{Name: "Электрический отдел"}
	var created dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/departments", in, &created, http.StatusCreated)

	var got dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/departments/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
	if got.Name != in.Name {
		t.Fatalf("name: got %q, want %q", got.Name, in.Name)
	}
}

func TestDepartmentGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.DepartmentResponse](t, env.Client, http.MethodGet, env.Base+"/departments/99999", nil, nil, http.StatusNotFound)
}

func TestDepartmentUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/departments",
		dto.CreateDepartmentRequest{Name: "Старое название"}, &created, http.StatusCreated)

	upd := dto.UpdateDepartmentRequest{Name: "Новое название", Description: "Обновлено"}
	var updated dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/departments/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.Name != upd.Name {
		t.Fatalf("name: got %q, want %q", updated.Name, upd.Name)
	}
	if updated.Description != upd.Description {
		t.Fatalf("description: got %q, want %q", updated.Description, upd.Description)
	}
}

func TestDepartmentUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	upd := dto.UpdateDepartmentRequest{Name: "Не существует"}
	common.DoJSON[dto.DepartmentResponse](t, env.Client, http.MethodPut, env.Base+"/departments/99999",
		upd, nil, http.StatusNotFound)
}

func TestDepartmentDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.DepartmentResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/departments",
		dto.CreateDepartmentRequest{Name: "Удалить меня"}, &created, http.StatusCreated)

	common.DoJSON[dto.DepartmentResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/departments/%d", env.Base, created.ID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.DepartmentResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/departments/%d", env.Base, created.ID), nil, nil, http.StatusNotFound)
}

func TestDepartmentDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.DepartmentResponse](t, env.Client, http.MethodDelete,
		env.Base+"/departments/99999", nil, nil, http.StatusNotFound)
}
