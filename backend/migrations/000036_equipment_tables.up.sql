-- 000036_equipment_tables.up.sql
-- Issue #54: 物资/设备借用管理

BEGIN;

-- 物资/设备表
CREATE TABLE IF NOT EXISTS equipment_items (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    model         VARCHAR(100),
    category      VARCHAR(50) NOT NULL DEFAULT 'general', -- general/electronic/book/tool/other
    total_quantity INT NOT NULL DEFAULT 1,
    available_quantity INT NOT NULL DEFAULT 1,
    location      VARCHAR(200),
    manager_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    photo_file_id UUID, -- 关联 file 表
    status        SMALLINT NOT NULL DEFAULT 0, -- 0=可用 1=部分借用 2=全部借出 3=维修中 4=已报废
    qr_code       VARCHAR(500), -- 二维码内容/URL
    description   TEXT,
    purchase_date DATE,
    purchase_price NUMERIC(12,2),
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE INDEX idx_equipment_items_name ON equipment_items(name);
CREATE INDEX idx_equipment_items_category ON equipment_items(category);
CREATE INDEX idx_equipment_items_status ON equipment_items(status);
CREATE INDEX idx_equipment_items_manager_id ON equipment_items(manager_id);
CREATE INDEX idx_equipment_items_deleted_at ON equipment_items(deleted_at);

-- 借用记录表
CREATE TABLE IF NOT EXISTS equipment_borrows (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    equipment_id    UUID NOT NULL REFERENCES equipment_items(id) ON DELETE CASCADE,
    borrower_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantity        INT NOT NULL DEFAULT 1,
    borrow_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expected_return_at TIMESTAMP NOT NULL,
    actual_return_at TIMESTAMP,
    status          SMALLINT NOT NULL DEFAULT 0, -- 0=待审批 1=已批准 2=已领取 3=已归还 4=已拒绝 5=逾期
    approver_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at     TIMESTAMP,
    approved_remark VARCHAR(500),
    return_remark   VARCHAR(500),
    return_checker_id UUID REFERENCES users(id) ON DELETE SET NULL,
    remark          VARCHAR(500),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX idx_equipment_borrows_equipment_id ON equipment_borrows(equipment_id);
CREATE INDEX idx_equipment_borrows_borrower_id ON equipment_borrows(borrower_id);
CREATE INDEX idx_equipment_borrows_status ON equipment_borrows(status);
CREATE INDEX idx_equipment_borrows_expected_return ON equipment_borrows(expected_return_at);
CREATE INDEX idx_equipment_borrows_deleted_at ON equipment_borrows(deleted_at);

-- 维修记录表
CREATE TABLE IF NOT EXISTS equipment_maintenance (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    equipment_id  UUID NOT NULL REFERENCES equipment_items(id) ON DELETE CASCADE,
    type          SMALLINT NOT NULL DEFAULT 0, -- 0=维修 1=保养 2=报废
    description   VARCHAR(500) NOT NULL,
    cost          NUMERIC(12,2),
    start_date    DATE NOT NULL,
    end_date      DATE,
    status        SMALLINT NOT NULL DEFAULT 0, -- 0=进行中 1=已完成 2=已取消
    operator_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    remark        VARCHAR(500),
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE INDEX idx_equipment_maintenance_equipment_id ON equipment_maintenance(equipment_id);
CREATE INDEX idx_equipment_maintenance_status ON equipment_maintenance(status);
CREATE INDEX idx_equipment_maintenance_deleted_at ON equipment_maintenance(deleted_at);

-- 库存盘点记录表
CREATE TABLE IF NOT EXISTS equipment_inventories (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    equipment_id  UUID NOT NULL REFERENCES equipment_items(id) ON DELETE CASCADE,
    expected_quantity INT NOT NULL,
    actual_quantity INT NOT NULL,
    difference    INT NOT NULL DEFAULT 0,
    checker_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    check_date    DATE NOT NULL,
    remark        VARCHAR(500),
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE INDEX idx_equipment_inventories_equipment_id ON equipment_inventories(equipment_id);
CREATE INDEX idx_equipment_inventories_check_date ON equipment_inventories(check_date);
CREATE INDEX idx_equipment_inventories_deleted_at ON equipment_inventories(deleted_at);

COMMIT;
