-- 000036_equipment_tables.down.sql
-- Issue #54: 物资/设备借用管理

BEGIN;

DROP TABLE IF EXISTS equipment_inventories;
DROP TABLE IF EXISTS equipment_maintenance;
DROP TABLE IF EXISTS equipment_borrows;
DROP TABLE IF EXISTS equipment_items;

COMMIT;
