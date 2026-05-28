package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/test/common"
)

func createInspector(t *testing.T, env *common.Env) int64 {
	t.Helper()
	var out dto.EmployeeResponse
	email := fmt.Sprintf("inspector%d@atom.ru", time.Now().UnixNano())
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/employees",
		dto.CreateEmployeeRequest{Email: email, Password: "secret12", FullName: "Инспектор", Role: "engineer"},
		&out, http.StatusCreated)
	return out.ID
}

func TestInspectionReportCreate(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Насос ИО")
	eqID := createEquipment(t, env, "ГЦН-ИО", "SN-IR1", etID)
	woID := createWorkOrder(t, env, eqID, "ТО для инспекции")
	inspID := createInspector(t, env)

	in := dto.CreateInspectionReportRequest{
		WorkOrderID:     woID,
		InspectorID:     inspID,
		Findings:        "Износ уплотнителей",
		Recommendations: "Замена уплотнителей в течение 30 дней",
	}
	var out dto.InspectionReportResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/inspection-reports", in, &out, http.StatusCreated)

	if out.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if out.WorkOrderID != woID {
		t.Fatalf("work_order_id: got %d, want %d", out.WorkOrderID, woID)
	}
	if out.Findings != in.Findings {
		t.Fatalf("findings: got %q, want %q", out.Findings, in.Findings)
	}
}

func TestInspectionReportCreate_InvalidWorkOrder(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	inspID := createInspector(t, env)
	in := dto.CreateInspectionReportRequest{
		WorkOrderID: 99999,
		InspectorID: inspID,
		Findings:    "Нет наряда",
	}
	common.DoJSON[dto.InspectionReportResponse](t, env.Client, http.MethodPost, env.Base+"/inspection-reports",
		in, nil, http.StatusNotFound)
}

func TestInspectionReportCreate_CompletedWorkOrder(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ИО Завершён")
	eqID := createEquipment(t, env, "Оборуд ИО", "SN-IR-C", etID)
	woID := createWorkOrder(t, env, eqID, "Завершённый наряд")
	inspID := createInspector(t, env)

	// open → in_progress → completed
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/work-orders/%d", env.Base, woID),
		dto.UpdateWorkOrderRequest{EquipmentID: eqID, Title: "Завершённый наряд", Status: "in_progress"},
		&dto.WorkOrderResponse{}, http.StatusOK)
	common.DoJSON(t, env.Client, http.MethodPut, fmt.Sprintf("%s/work-orders/%d", env.Base, woID),
		dto.UpdateWorkOrderRequest{EquipmentID: eqID, Title: "Завершённый наряд", Status: "completed"},
		&dto.WorkOrderResponse{}, http.StatusOK)

	in := dto.CreateInspectionReportRequest{WorkOrderID: woID, InspectorID: inspID, Findings: "Попытка"}
	common.DoJSON[dto.InspectionReportResponse](t, env.Client, http.MethodPost, env.Base+"/inspection-reports",
		in, nil, http.StatusBadRequest)
}

func TestInspectionReportList(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ИО Лист")
	eqID := createEquipment(t, env, "Оборуд ИО Лист", "SN-IR-L", etID)
	inspID := createInspector(t, env)

	// Один протокол на наряд (1:1), поэтому для каждого отчёта — отдельный наряд
	for i := range 3 {
		woID := createWorkOrder(t, env, eqID, fmt.Sprintf("Наряд для отчётов %d", i))
		in := dto.CreateInspectionReportRequest{
			WorkOrderID: woID,
			InspectorID: inspID,
			Findings:    fmt.Sprintf("Нарушение %d", i),
		}
		common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/inspection-reports", in, &dto.InspectionReportResponse{}, http.StatusCreated)
	}

	var list []dto.InspectionReportResponse
	common.DoJSON(t, env.Client, http.MethodGet, env.Base+"/inspection-reports", nil, &list, http.StatusOK)

	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestInspectionReportGetByID(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ИО Гет")
	eqID := createEquipment(t, env, "Оборуд ИО Гет", "SN-IR-G", etID)
	woID := createWorkOrder(t, env, eqID, "Наряд ИО Гет")
	inspID := createInspector(t, env)

	var created dto.InspectionReportResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/inspection-reports",
		dto.CreateInspectionReportRequest{WorkOrderID: woID, InspectorID: inspID, Findings: "Трещины"},
		&created, http.StatusCreated)

	var got dto.InspectionReportResponse
	common.DoJSON(t, env.Client, http.MethodGet, fmt.Sprintf("%s/inspection-reports/%d", env.Base, created.ID), nil, &got, http.StatusOK)

	if got.ID != created.ID {
		t.Fatalf("ID: got %d, want %d", got.ID, created.ID)
	}
	if got.Findings != created.Findings {
		t.Fatalf("findings: got %q, want %q", got.Findings, created.Findings)
	}
}

func TestInspectionReportGetByID_NotFound(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)

	common.DoJSON[dto.InspectionReportResponse](t, env.Client, http.MethodGet,
		env.Base+"/inspection-reports/99999", nil, nil, http.StatusNotFound)
}

func TestInspectionReportDelete_Forbidden(t *testing.T) {
	env := common.MustStart(t)
	common.TruncateAll(t, env.DB)
	common.EnsureTestEmployee(t, env.DB)

	etID := createEquipmentType(t, env, "Тип ИО Запрет")
	eqID := createEquipment(t, env, "Оборуд ИО Запрет", "SN-IR-F", etID)
	woID := createWorkOrder(t, env, eqID, "Наряд ИО Запрет")
	inspID := createInspector(t, env)

	var created dto.InspectionReportResponse
	common.DoJSON(t, env.Client, http.MethodPost, env.Base+"/inspection-reports",
		dto.CreateInspectionReportRequest{WorkOrderID: woID, InspectorID: inspID, Findings: "Неизменяемый"},
		&created, http.StatusCreated)

	// Удаление протокола запрещено — документ неизменяем (ФТ-8)
	common.DoJSON[dto.InspectionReportResponse](t, env.Client, http.MethodDelete,
		fmt.Sprintf("%s/inspection-reports/%d", env.Base, created.ID), nil, nil, http.StatusForbidden)

	// Протокол по-прежнему доступен
	common.DoJSON[dto.InspectionReportResponse](t, env.Client, http.MethodGet,
		fmt.Sprintf("%s/inspection-reports/%d", env.Base, created.ID), nil, nil, http.StatusOK)
}
