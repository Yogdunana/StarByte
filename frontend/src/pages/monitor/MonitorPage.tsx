import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, Card, Col, Row, Space, Statistic, Switch, Table, Tag, Typography } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import ChartCard from '@/components/Chart/ChartCard';
import {
  getMonitorAPIStats,
  getMonitorApp,
  getMonitorDatabase,
  getMonitorRedis,
  getMonitorServer,
  getMonitorSlowQueries,
  type MonitorLiveSnapshot,
  type MonitorSlowQueries,
  type MonitorSlowQuery,
} from '@/api/monitor';
import { isCanceledError } from '@/api/error';
import { getToken } from '@/utils/storage';
import { gaugeOption, sparkOption } from './charts';
import { formatBytes, formatDuration, formatPercent, formatSeconds, pushSample } from './format';
import {
  SNAPSHOT_KEYS,
  applyLiveSnapshot,
  applySettledIfCurrent,
  createPollSession,
  type Snapshot,
  type SnapshotKey,
} from './load';
import { buildMonitorWSUrl, parseMonitorWSFrame } from './ws';
import './monitor.css';

const POLL_MS = 8000;
const SLOW_POLL_MS = 15000;

const MonitorPage: React.FC = () => {
  const { t } = useTranslation();
  const [data, setData] = useState<Snapshot>({});
  const [failed, setFailed] = useState<SnapshotKey[]>([]);
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);
  const [auto, setAuto] = useState(true);
  const [updatedAt, setUpdatedAt] = useState<string>();
  const [cpuHist, setCpuHist] = useState<number[]>([]);
  const [labels, setLabels] = useState<string[]>([]);
  const [live, setLive] = useState(false);
  const [slow, setSlow] = useState<MonitorSlowQueries>();
  const [slowFailed, setSlowFailed] = useState(false);
  const dataRef = useRef<Snapshot>({});
  const pollRef = useRef(createPollSession());
  const liveRef = useRef(false);
  dataRef.current = data;
  liveRef.current = live;

  const sectionTitle = useCallback((key: SnapshotKey) => {
    const map: Record<SnapshotKey, string> = {
      server: t('monitor.server'),
      app: t('monitor.app'),
      database: t('monitor.database'),
      redis: t('monitor.redis'),
      api: t('monitor.api'),
    };
    return map[key];
  }, [t]);

  const load = useCallback(async () => {
    const ticket = pollRef.current.begin();
    setLoading(true);
    const results = await Promise.allSettled([
      getMonitorServer(ticket.signal),
      getMonitorApp(ticket.signal),
      getMonitorDatabase(ticket.signal),
      getMonitorRedis(ticket.signal),
      getMonitorAPIStats(ticket.signal),
    ]);
    const merged = applySettledIfCurrent(
      pollRef.current,
      ticket.gen,
      dataRef.current,
      results,
      SNAPSHOT_KEYS,
    );
    if (!merged) return;
    setData(merged.next);
    setFailed(merged.failed);
    if (merged.succeeded.includes('server') && merged.next.server) {
      const stamp = new Date().toLocaleTimeString();
      setUpdatedAt(stamp);
      setCpuHist((hist) => pushSample(hist, formatPercent(merged.next.server?.cpu_percent)));
      setLabels((prevLabels) => pushSample(prevLabels, stamp));
    } else if (merged.succeeded.length > 0) {
      setUpdatedAt(new Date().toLocaleTimeString());
    }
    setError(merged.failed.length ? t('monitor.loadFail') : undefined);
    setLoading(false);
  }, [t]);

  const loadSlow = useCallback(async (signal?: AbortSignal) => {
    try {
      const next = await getMonitorSlowQueries(signal);
      setSlow(next);
      setSlowFailed(false);
    } catch (err) {
      if (isCanceledError(err)) return;
      setSlowFailed(true);
    }
  }, []);

  const applyLive = useCallback((incoming: MonitorLiveSnapshot) => {
    const merged = applyLiveSnapshot(dataRef.current, incoming);
    setData(merged.next);
    setFailed(merged.failed);
    if (merged.succeeded.includes('server') && merged.next.server) {
      const stamp = new Date().toLocaleTimeString();
      setUpdatedAt(stamp);
      setCpuHist((hist) => pushSample(hist, formatPercent(merged.next.server?.cpu_percent)));
      setLabels((prevLabels) => pushSample(prevLabels, stamp));
    } else if (merged.succeeded.length > 0) {
      setUpdatedAt(new Date().toLocaleTimeString());
    }
    setError(merged.failed.length ? t('monitor.loadFail') : undefined);
    setLoading(false);
  }, [t]);

  useEffect(() => {
    const session = pollRef.current;
    void load();
    void loadSlow();
    return () => { session.invalidate(); };
  }, [load, loadSlow]);

  useEffect(() => {
    if (!auto) return undefined;
    const id = window.setInterval(() => {
      if (!liveRef.current) void load();
    }, POLL_MS);
    return () => window.clearInterval(id);
  }, [auto, load]);

  useEffect(() => {
    if (!auto) return undefined;
    const id = window.setInterval(() => { void loadSlow(); }, SLOW_POLL_MS);
    return () => window.clearInterval(id);
  }, [auto, loadSlow]);

  useEffect(() => {
    if (!auto) {
      setLive(false);
      return undefined;
    }
    const token = getToken();
    if (!token) return undefined;
    let disposed = false;
    let socket: WebSocket | null = null;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let attempt = 0;
    const connect = () => {
      const currentToken = getToken();
      if (disposed || !currentToken) return;
      const current = new WebSocket(buildMonitorWSUrl(currentToken));
      socket = current;
      current.onopen = () => {
        if (disposed) { current.close(); return; }
        attempt = 0;
        setLive(true);
      };
      current.onmessage = (event) => {
        if (disposed) return;
        const frame = parseMonitorWSFrame(String(event.data));
        if (frame?.type === 'snapshot' && frame.data && typeof frame.data === 'object') {
          applyLive(frame.data as MonitorLiveSnapshot);
        }
      };
      current.onclose = () => {
        if (disposed) return;
        setLive(false);
        reconnect = setTimeout(connect, Math.min(30000, 2000 * 2 ** Math.min(attempt++, 4)));
      };
      current.onerror = () => { if (!disposed) setLive(false); };
    };
    connect();
    const heartbeat = window.setInterval(() => {
      if (socket?.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'ping' }));
    }, 30000);
    return () => {
      disposed = true;
      window.clearInterval(heartbeat);
      clearTimeout(reconnect);
      if (socket) {
        socket.onopen = socket.onclose = socket.onmessage = socket.onerror = null;
        socket.close(1000, 'page disconnected');
      }
      setLive(false);
    };
  }, [applyLive, auto]);

  const server = data.server;
  const gauges = useMemo(() => ([
    { key: 'cpu', title: t('monitor.cpu'), value: server?.cpu_percent ?? 0 },
    { key: 'mem', title: t('monitor.memory'), value: server?.mem_percent ?? 0 },
    { key: 'disk', title: t('monitor.disk'), value: server?.disk_percent ?? 0 },
  ]), [server, t]);

  const failedNow = (key: SnapshotKey) => failed.includes(key);
  const cardEmpty = (key: SnapshotKey) => failedNow(key) && data[key] == null;

  return (
    <div className="monitor-page">
      <div className="monitor-hero">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>{t('monitor.title')}</Typography.Title>
          <p className="monitor-meta">{t('monitor.hint')}</p>
        </div>
        <Space wrap>
          <Tag color={live ? 'success' : 'default'}>{live ? t('monitor.liveOn') : t('monitor.liveOff')}</Tag>
          <span>{t('monitor.autoRefresh')}</span>
          <Switch checked={auto} onChange={setAuto} />
          <Button size="large" icon={<ReloadOutlined />} onClick={() => { void load(); void loadSlow(); }}>
            {t('monitor.refresh')}
          </Button>
        </Space>
      </div>

      {error ? (
        <Alert
          type="warning"
          showIcon
          message={error}
          description={failed.length ? t('monitor.loadPartial', {
            parts: failed.map(sectionTitle).join(t('monitor.partSep')),
          }) : undefined}
          style={{ marginBottom: 16 }}
        />
      ) : null}
      {updatedAt ? (
        <Typography.Paragraph type="secondary">{t('monitor.updatedAt', { time: updatedAt })}</Typography.Paragraph>
      ) : null}

      <Row gutter={[16, 16]} className="monitor-section">
        {cardEmpty('server') ? (
          <Col span={24}><Alert type="warning" showIcon message={t('monitor.cardUnavailable')} /></Col>
        ) : gauges.map((g) => (
          <Col xs={24} md={8} key={g.key}>
            <ChartCard
              title={g.title}
              loading={loading && !server}
              option={gaugeOption(g.title, g.value)}
              height={220}
            />
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]} className="monitor-section">
        <Col xs={24} lg={12}>
          <ChartCard
            title={t('monitor.cpuTrend')}
            loading={loading && cpuHist.length === 0}
            empty={cpuHist.length === 0}
            option={sparkOption(labels, cpuHist, t('monitor.cpu'))}
            height={240}
          />
        </Col>
        <Col xs={24} lg={12}>
          <Card title={t('monitor.load')} className="monitor-section">
            {cardEmpty('server') ? (
              <Alert type="warning" message={t('monitor.cardUnavailable')} />
            ) : (
              <>
                <Row gutter={16}>
                  <Col span={8}><Statistic title="1m" value={server?.load1 ?? 0} precision={2} /></Col>
                  <Col span={8}><Statistic title="5m" value={server?.load5 ?? 0} precision={2} /></Col>
                  <Col span={8}><Statistic title="15m" value={server?.load15 ?? 0} precision={2} /></Col>
                </Row>
                <p className="monitor-note">
                  {t('monitor.hostMeta', { os: server?.goos || '-', arch: server?.goarch || '-' })}
                </p>
                <Statistic
                  title={t('monitor.memDetail')}
                  value={`${formatBytes(server?.mem_used)} / ${formatBytes(server?.mem_total)}`}
                />
                <Statistic
                  title={t('monitor.diskDetail')}
                  value={`${formatBytes(server?.disk_used)} / ${formatBytes(server?.disk_total)}`}
                />
              </>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="monitor-section">
        <Col xs={24} md={12}>
          <Card title={t('monitor.app')} loading={loading && !data.app}>
            {cardEmpty('app') ? (
              <Alert type="warning" message={t('monitor.cardUnavailable')} />
            ) : (
              <>
                <Row gutter={[16, 16]}>
                  <Col span={12}><Statistic title={t('monitor.uptime')} value={formatDuration(data.app?.uptime_seconds)} /></Col>
                  <Col span={12}><Statistic title={t('monitor.goroutines')} value={data.app?.goroutines ?? 0} /></Col>
                  <Col span={12}><Statistic title={t('monitor.heap')} value={formatBytes(data.app?.heap_alloc)} /></Col>
                  <Col span={12}><Statistic title={t('monitor.numGC')} value={data.app?.num_gc ?? 0} /></Col>
                </Row>
                <p className="monitor-note">{data.app?.go_version} · {data.app?.version}</p>
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('monitor.database')} loading={loading && !data.database}>
            {cardEmpty('database') ? (
              <Alert type="warning" message={t('monitor.cardUnavailable')} />
            ) : !data.database?.available ? (
              <Alert type="warning" message={t('monitor.dbDown')} />
            ) : (
              <Row gutter={[16, 16]}>
                <Col span={8}><Statistic title={t('monitor.dbInUse')} value={data.database.in_use} /></Col>
                <Col span={8}><Statistic title={t('monitor.dbIdle')} value={data.database.idle} /></Col>
                <Col span={8}><Statistic title={t('monitor.dbOpen')} value={data.database.open_connections} /></Col>
                <Col span={8}><Statistic title={t('monitor.dbWait')} value={data.database.wait_count} /></Col>
                <Col span={8}><Statistic title={t('monitor.dbMax')} value={data.database.max_open_connections} /></Col>
                <Col span={8}><Statistic title={t('monitor.dbWaitMs')} value={data.database.wait_duration_ms} /></Col>
              </Row>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="monitor-section">
        <Col xs={24} md={12}>
          <Card title={t('monitor.redis')} loading={loading && !data.redis}>
            {cardEmpty('redis') ? (
              <Alert type="warning" message={t('monitor.cardUnavailable')} />
            ) : !data.redis?.available ? (
              <Alert type="warning" message={t('monitor.redisDown')} />
            ) : (
              <>
                <Row gutter={[16, 16]}>
                  <Col span={8}><Statistic title={t('monitor.redisClients')} value={data.redis.connected_clients} /></Col>
                  <Col span={8}><Statistic title={t('monitor.redisMemory')} value={formatBytes(data.redis.used_memory)} /></Col>
                  <Col span={8}><Statistic title={t('monitor.redisHitRate')} value={data.redis.hit_rate} suffix="%" /></Col>
                  <Col span={8}><Statistic title={t('monitor.redisKeys')} value={data.redis.keys} /></Col>
                  <Col span={8}><Statistic title={t('monitor.redisHits')} value={data.redis.keyspace_hits} /></Col>
                  <Col span={8}><Statistic title={t('monitor.redisMisses')} value={data.redis.keyspace_misses} /></Col>
                </Row>
                <p className="monitor-note">{data.redis?.version}</p>
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('monitor.api')} loading={loading && !data.api}>
            {cardEmpty('api') ? (
              <Alert type="warning" message={t('monitor.cardUnavailable')} />
            ) : (
              <>
                <Row gutter={[16, 16]}>
                  <Col span={8}><Statistic title={t('monitor.apiTotal')} value={data.api?.request_total ?? 0} /></Col>
                  <Col span={8}><Statistic title={t('monitor.apiErrors')} value={data.api?.error_total ?? 0} /></Col>
                  <Col span={8}><Statistic title={t('monitor.apiErrorRate')} value={data.api?.error_rate ?? 0} suffix="%" /></Col>
                  <Col span={8}><Statistic title={t('monitor.apiP50')} value={formatSeconds(data.api?.p50_seconds)} /></Col>
                  <Col span={8}><Statistic title={t('monitor.apiP95')} value={formatSeconds(data.api?.p95_seconds)} /></Col>
                  <Col span={8}><Statistic title={t('monitor.apiP99')} value={formatSeconds(data.api?.p99_seconds)} /></Col>
                </Row>
                <p className="monitor-note">{data.api?.percentiles_note || t('monitor.apiNote')}</p>
              </>
            )}
          </Card>
        </Col>
      </Row>

      <Card title={t('monitor.slowQueries')} className="monitor-section">
        {slowFailed && !slow ? (
          <Alert type="warning" message={t('monitor.cardUnavailable')} />
        ) : (
          <>
            <p className="monitor-note">{slow?.note || t('monitor.slowEmpty')}</p>
            <Table<MonitorSlowQuery>
              size="small"
              rowKey={(row, i) => `${row.source}-${row.query}-${i}`}
              pagination={false}
              dataSource={slow?.queries || []}
              locale={{ emptyText: t('monitor.slowEmpty') }}
              columns={[
                { title: t('monitor.slowQuery'), dataIndex: 'query', ellipsis: true },
                { title: t('monitor.slowSource'), dataIndex: 'source', width: 160 },
                { title: t('monitor.slowCalls'), dataIndex: 'calls', width: 80 },
                { title: t('monitor.slowMean'), dataIndex: 'mean_time_ms', width: 110, render: (v: number) => `${Number(v || 0).toFixed(1)} ms` },
                { title: t('monitor.slowMax'), dataIndex: 'max_time_ms', width: 110, render: (v: number) => `${Number(v || 0).toFixed(1)} ms` },
                { title: t('monitor.slowRows'), dataIndex: 'rows', width: 80 },
              ]}
            />
            {(slow?.redis_commands || []).length > 0 ? (
              <>
                <p className="monitor-note">{t('monitor.redisSlow')}</p>
                <Table<MonitorSlowQuery>
                  size="small"
                  rowKey={(row, i) => `redis-${row.query}-${i}`}
                  pagination={false}
                  dataSource={slow?.redis_commands || []}
                  columns={[
                    { title: t('monitor.slowQuery'), dataIndex: 'query', ellipsis: true },
                    { title: t('monitor.slowMean'), dataIndex: 'mean_time_ms', width: 110, render: (v: number) => `${Number(v || 0).toFixed(1)} ms` },
                  ]}
                />
              </>
            ) : null}
          </>
        )}
      </Card>
    </div>
  );
};

export default MonitorPage;
