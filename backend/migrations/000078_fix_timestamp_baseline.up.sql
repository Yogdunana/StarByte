-- 把被写偏的「记录发生时间」回拨成 UTC 墙钟。
--
-- 为什么会有写偏的行：
--   timestamp（不带时区）这一列上，应用和数据库用的是两套基准。
--   - 应用侧：pgx 写入时会把 time.Time 转成 UTC 墙钟（discardTimeZone），
--     读取也按 UTC 解释 —— 所以这一列里「正确的语义」是 UTC 墙钟。
--   - 数据库侧：DEFAULT CURRENT_TIMESTAMP 写入时按**服务端 timezone** 取墙钟。
--     服务端 timezone 是 initdb 那一刻定死的，后来往 compose 里加 TZ=Asia/Shanghai
--     并不会改它。一旦它落在 Asia/Shanghai，DB 默认值就会写进北京时间墙钟，
--     比应用写的值快 8 小时，两种基准混在同一列里。
--
-- 怎么修：把服务端 timezone 显式固定成 UTC（见 docker-compose.yml 的 -c timezone=UTC），
-- 再把**已经写偏的历史行**回拨回去。偏移量从 current_setting 现场算，
-- 不写死 8 小时，免得以后换时区改错。
--
-- 判定条件必须写成「UTC 墙钟」比较：
--   写 `col > now()` 是错的 —— Postgres 会把 col 按服务端 timezone 解释成 timestamptz，
--   而服务�� timezone 恰好就是当初写入用的那个，结果两种基准看起来都正常，一行都检不出来。
--   正确写法是跟 `now() AT TIME ZONE 'UTC'` 比（两边都是裸 timestamp）。
--
-- 只动 created_at / updated_at / deleted_at 这类「已经发生过」的时间戳：
--   它们的真实值永远不可能晚于现在。业务时刻（start_time / expires_at / scheduled_at 等）
--   天然就在未来，一旦被这条 UPDATE 扫到就会无辜减 8 小时，所以一律不碰。
--
-- 幂等：回拨后这些值不再晚于现在，重跑不会有任何行命中。
-- 无损：只 UPDATE 命中的单元格，不删不改表结构，数据一字不丢。

DO $$
DECLARE
    tz_offset   interval;
    rec         record;
    changed     bigint;
    total_rows  bigint := 0;
    total_cells bigint := 0;
BEGIN
    -- 服务端 timezone 相对 UTC 的偏移。UTC 下为 0，Asia/Shanghai 下为 +08:00。
    SELECT (now()::timestamp - (now() AT TIME ZONE 'UTC')) INTO tz_offset;

    IF tz_offset = interval '0' THEN
        RAISE NOTICE 'timezone 已是 UTC，created_at/updated_at 无需回拨';
        RETURN;
    END IF;

    RAISE NOTICE '检测到服务端 timezone 偏移 %，开始回拨写偏的时间戳', tz_offset;

    FOR rec IN
        SELECT c.table_schema, c.table_name, c.column_name
        FROM information_schema.columns c
        JOIN information_schema.tables t
          ON t.table_schema = c.table_schema AND t.table_name = c.table_name
        WHERE c.data_type = 'timestamp without time zone'
          AND c.column_name IN ('created_at', 'updated_at', 'deleted_at')
          AND t.table_type = 'BASE TABLE'
          AND c.table_schema NOT IN ('pg_catalog', 'information_schema')
        ORDER BY c.table_schema, c.table_name, c.column_name
    LOOP
        EXECUTE format(
            'UPDATE %I.%I SET %I = %I - %L::interval '
            'WHERE %I > (now() AT TIME ZONE ''UTC'') + interval ''30 minutes''',
            rec.table_schema, rec.table_name, rec.column_name,
            rec.column_name, tz_offset,
            rec.column_name
        );
        GET DIAGNOSTICS changed = ROW_COUNT;
        IF changed > 0 THEN
            total_cells := total_cells + changed;
            total_rows := total_rows + 1;
            RAISE NOTICE '已回拨 %.%.% 共 % 行',
                rec.table_schema, rec.table_name, rec.column_name, changed;
        END IF;
    END LOOP;

    RAISE NOTICE '回拨完成：% 个列、共 % 个单元格', total_rows, total_cells;
END $$;
