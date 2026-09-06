-- Issue #48：数据字典。新表，不改已有列名。

CREATE TABLE IF NOT EXISTS dict_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dict_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_id UUID NOT NULL REFERENCES dict_types(id) ON DELETE CASCADE,
    item_value VARCHAR(50) NOT NULL,
    item_label VARCHAR(100) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type_id, item_value)
);

CREATE INDEX IF NOT EXISTS idx_dict_items_type_id ON dict_items(type_id);
CREATE INDEX IF NOT EXISTS idx_dict_types_status ON dict_types(status);
