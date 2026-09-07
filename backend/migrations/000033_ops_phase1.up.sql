-- ============================================================
-- 000033_ops_phase1.up.sql
-- 对齐 #22/#23/#24：处分状态/等级语义、合同类型金额签约方、到期扫描
-- ============================================================

COMMENT ON COLUMN discipline_records.level IS '1=警告 2=严重警告 3=记过 4=留会察看 5=开除会籍';
COMMENT ON COLUMN discipline_records.status IS '0=待审批 1=生效 2=已撤销 3=申诉中';

ALTER TABLE discipline_records
    ADD COLUMN IF NOT EXISTS flow_instance_id UUID,
    ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS revoke_reason TEXT,
    ADD COLUMN IF NOT EXISTS revoked_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP;

COMMENT ON COLUMN contracts.status IS '0=草稿 1=生效中 2=已到期 3=已终止';

ALTER TABLE contracts
    ADD COLUMN IF NOT EXISTS contract_type SMALLINT NOT NULL DEFAULT 4,
    ADD COLUMN IF NOT EXISTS party_name VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS amount DECIMAL(12,2),
    ADD COLUMN IF NOT EXISTS start_at DATE;

COMMENT ON COLUMN contracts.contract_type IS '1=赞助 2=活动 3=采购 4=其他';

CREATE INDEX IF NOT EXISTS idx_contracts_expired_at ON contracts(expired_at);
CREATE INDEX IF NOT EXISTS idx_discipline_records_level ON discipline_records(level);
