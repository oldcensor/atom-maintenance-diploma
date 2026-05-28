CREATE INDEX idx_work_order_status_created    ON work_order(status, created_at);
CREATE INDEX idx_work_order_equipment_assigned ON work_order(equipment_id, assigned_to);
CREATE INDEX idx_schedule_equipment_due        ON maintenance_schedule(equipment_id, next_due_at);
