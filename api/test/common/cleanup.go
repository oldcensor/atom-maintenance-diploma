package common

import (
	"testing"

	"gorm.io/gorm"
)

func TruncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `TRUNCATE TABLE
		inspection_report,
		work_order,
		maintenance_schedule,
		equipment,
		employee,
		equipment_type,
		department
	RESTART IDENTITY CASCADE`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	db.Exec("SELECT setval('employee_id_seq', 2)")
	db.Exec("SELECT setval('equipment_type_id_seq', 2)")
	db.Exec("SELECT setval('department_id_seq', 2)")
	db.Exec("SELECT setval('equipment_id_seq', 2)")
	db.Exec("SELECT setval('maintenance_schedule_id_seq', 2)")
	db.Exec("SELECT setval('work_order_id_seq', 2)")
	db.Exec("SELECT setval('inspection_report_id_seq', 2)")
}
