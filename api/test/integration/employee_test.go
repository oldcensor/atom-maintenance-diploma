package integration

import (
	"fmt"
	"net/http"
	"testing"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func newEmployee(i int) dto.CreateEmployeeRequest {
	return dto.CreateEmployeeRequest{
		Email:    fmt.Sprintf("worker%d@atom.ru", i),
		Password: "password123",
		FullName: fmt.Sprintf("Работник %d", i),
		Role:     "technician",
	}
}

func TestEmployeeCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateEmployeeRequest{
		Email:    "ivan@atom.ru",
		Password: "secret12",
		FullName: "Иван Иванов",
		Role:     "engineer",
	}
	var out dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.Email != in.Email {
		t.Fatalf("email: got %q, want %q", out.Email, in.Email)
	}
	if out.Role != in.Role {
		t.Fatalf("role: got %q, want %q", out.Role, in.Role)
	}
}

func TestEmployeeCreate_DuplicateEmail(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	in := dto.CreateEmployeeRequest{Email: "dup@atom.ru", Password: "secret12", FullName: "Дубликат", Role: "technician"}
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", in, &dto.EmployeeResponse{}, http.StatusCreated)
	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodPost, env.Base+"/employees", in, nil, http.StatusConflict)
}

func TestEmployeeList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	for i := range 3 {
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", newEmployee(i), &dto.EmployeeResponse{}, http.StatusCreated)
	}

	var list []dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/employees", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestEmployeeGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", newEmployee(1), &created, http.StatusCreated)

	var got dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/employees/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
}

func TestEmployeeGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodGet, env.Base+"/employees/99999", nil, nil, http.StatusNotFound)
}

func TestEmployeeUpdate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", newEmployee(1), &created, http.StatusCreated)

	upd := dto.UpdateEmployeeRequest{FullName: "Пётр Петров", Role: "engineer"}
	var updated dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/employees/%d", env.Base, created.ID),
		upd, &updated, http.StatusOK)

	if updated.FullName != upd.FullName {
		t.Fatalf("full_name: got %q, want %q", updated.FullName, upd.FullName)
	}
	if updated.Role != upd.Role {
		t.Fatalf("role: got %q, want %q", updated.Role, upd.Role)
	}
}

func TestEmployeeUpdate_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	upd := dto.UpdateEmployeeRequest{FullName: "Никто", Role: "technician"}
	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodPut, env.Base+"/employees/99999",
		upd, nil, http.StatusNotFound)
}

func TestEmployeeDelete(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	var created dto.EmployeeResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees", newEmployee(1), &created, http.StatusCreated)

	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/employees/%d", env.Base, created.ID), nil, nil, http.StatusNoContent)

	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/employees/%d", env.Base, created.ID), nil, nil, http.StatusNotFound)
}

func TestEmployeeDelete_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.EmployeeResponse](t, env.Client, http.MethodDelete,
		env.Base+"/employees/99999", nil, nil, http.StatusNotFound)
}
