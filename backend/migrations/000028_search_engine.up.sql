-- 000028_search_engine.up.sql
-- Unified search: CJK tokenizer + FTS/trgm indexes (#74)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION starbyte_cjk_tokens(t text)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    s text := coalesce(t, '');
    i int;
    n int;
    ch text;
    buf text := '';
    pieces text := '';
    last_cjk text := NULL;
BEGIN
    n := char_length(s);
    FOR i IN 1..n LOOP
        ch := substr(s, i, 1);
        IF ch ~ '^[一-鿿]$' THEN
            IF buf <> '' THEN
                pieces := pieces || ' ' || lower(buf);
                buf := '';
            END IF;
            pieces := pieces || ' ' || ch;
            IF last_cjk IS NOT NULL THEN
                pieces := pieces || ' ' || last_cjk || ch;
            END IF;
            last_cjk := ch;
        ELSIF ch ~ '^[A-Za-z0-9_]$' THEN
            last_cjk := NULL;
            buf := buf || ch;
        ELSE
            IF buf <> '' THEN
                pieces := pieces || ' ' || lower(buf);
                buf := '';
            END IF;
            last_cjk := NULL;
        END IF;
    END LOOP;
    IF buf <> '' THEN
        pieces := pieces || ' ' || lower(buf);
    END IF;
    RETURN btrim(pieces);
END;
$$;

CREATE INDEX IF NOT EXISTS idx_tasks_search_fts ON tasks USING GIN (
    to_tsvector('simple', starbyte_cjk_tokens(
        coalesce(title, '') || ' ' || coalesce(description, '') || ' ' || coalesce(tags, '')
    ))
);
CREATE INDEX IF NOT EXISTS idx_tasks_title_trgm ON tasks USING GIN (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_users_search_fts ON users USING GIN (
    to_tsvector('simple', starbyte_cjk_tokens(
        coalesce(username, '') || ' ' || coalesce(real_name, '') || ' ' || coalesce(email, '') || ' ' || coalesce(phone, '')
    ))
) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_real_name_trgm ON users USING GIN (real_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_audit_logs_search_fts ON audit_logs USING GIN (
    to_tsvector('simple', starbyte_cjk_tokens(
        coalesce(operation, '') || ' ' || coalesce(path, '') || ' ' ||
        coalesce(username, '') || ' ' || coalesce(real_name, '') || ' ' || coalesce(module, '')
    ))
);

CREATE INDEX IF NOT EXISTS idx_member_applications_search_fts ON member_applications USING GIN (
    to_tsvector('simple', starbyte_cjk_tokens(
        coalesce(reason, '') || ' ' || coalesce(current_stage, '')
    ))
);
