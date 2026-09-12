import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Row, Space, Statistic, Switch, Typography } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import ChartCard from '@/components/Chart/ChartCard';
import {
  getMonitorAPIStats,
  getMonitorApp,
  getMonitorDatabase,
  getMonitorRedis,
  getMonitorServer,
  type MonitorAPIStats,
  type MonitorApp,
  type MonitorDatabase,
  type MonitorRedis,
  type MonitorServer,
} from '@/api/monitor';
import { gaugeOption, sparkOption } from './charts';
import { formatBytes, formatDuration, formatPercent, pushSample } from './format';
import './monitor.css';

const POLL_MS = 8000;

interface Snapshot {
  server?: MonitorServer;
  app?: MonitorApp;
  database?: MonitorDatabase;
  redis?: MonitorRedis;
  api?: MonitorAPIStats;
}

const MonitorPage: React.FC = () => {
  const { t } = useTranslation();
  const [data, setData] = useState<Snapshot>({});
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);
  const [auto, setAuto] = useState(true);
  const [updatedAt, setUpdatedAt] = useState<string>();
  const [cpuHist, setCpuHist] = useState<number[]>([]);
  const [labels, setLabels] = useState<string[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [server, app, database, redis, api] = await Promise.all([
        getMonitorServer(),
        getMonitorApp(),
        getMonitorDatabase(),
        getMonitorRedis(),
        getMonitorAPIStats(),
      ]);
      setData({ server, app, database, redis, api });
      setError(undefined);
      const stamp = new Date().toLocaleTimeString();
      setUpdatedAt(stamp);
      setCpuHist((prev) => pushSample(prev, formatPercent(server.cpu_percent)));
      setLabels((prev) => pushSample(prev, stamp));
    } catch (e) {
      setError(e instanceof Error ? e.message : t('monitor.loadFail'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!auto) return undefined;
    const id = window.setInterval(() => { void load(); }, POLL_MS);
    return () => window.clearInterval(id);
  }, [auto, load]);

  const server = data.server;
  const gauges = useMemo(() => ([
    { key: 'cpu', title: t('monitor.cpu'), value: server?.cpu_percent ?? 0 },
    { key: 'mem', title: t('monitor.memory'), value: server?.mem_percent ?? 0 },
    { key: 'disk', title: t('monitor.disk'), value: server?.disk_percent ?? 0 },
  ]), [server, t]);

  return (
    <div className="monitor-page">
      <div className="monitor-hero">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>{t('monitor.title')}</Typography.Title>
          <p className="monitor-meta">{t('monitor.hint')}</p>
        </div>
        <Space wrap>
          <span>{t('monitor.autoRefresh')}</span>
          <Switch checked={auto} onChange={setAuto} />
          <Button size="large" icon={<ReloadOutlined />} onClick={() => void load()}>
            {t('monitor.refresh')}
          </Button>
        </Space>
      </div>

      {error ? <Alert type="error" showIcon message={error} style={{ marginBottom: 16 }} /> : null}
      {updatedAt ? (
        <Typography.Paragraph type="secondary">{t('monitor.updatedAt', { time: updatedAt })}</Typography.Paragraph>
      ) : null}

      <Row gutter={[16, 16]} className="monitor-section">
        {gauges.map((g) => (
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
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="monitor-section">
        <Col xs={24} md={12}>
          <Card title={t('monitor.app')} loading={loading && !data.app}>
            <Row gutter={[16, 16]}>
              <Col span={12}><Statistic title={t('monitor.uptime')} value={formatDuration(data.app?.uptime_seconds)} /></Col>
              <Col span={12}><Statistic title={t('monitor.goroutines')} value={data.app?.goroutines ?? 0} /></Col>
              <Col span={12}><Statistic title={t('monitor.heap')} value={formatBytes(data.app?.heap_alloc)} /></Col>
              <Col span={12}><Statistic title={t('monitor.numGC')} value={data.app?.num_gc ?? 0} /></Col>
            </Row>
            <p className="monitor-note">{data.app?.go_version} · {data.app?.version}</p>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('monitor.database')} loading={loading && !data.database}>
            {!data.database?.available ? (
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
            {!data.redis?.available ? (
              <Alert type="warning" message={t('monitor.redisDown')} />
            ) : (
              <Row gutter={[16, 16]}>
                <Col span={8}><Statistic title={t('monitor.redisClients')} value={data.redis.connected_clients} /></Col>
                <Col span={8}><Statistic title={t('monitor.redisMemory')} value={formatBytes(data.redis.used_memory)} /></Col>
                <Col span={8}><Statistic title={t('monitor.redisHitRate')} value={data.redis.hit_rate} suffix="%" /></Col>
                <Col span={8}><Statistic title={t('monitor.redisKeys')} value={data.redis.keys} /></Col>
                <Col span={8}><Statistic title={t('monitor.redisHits')} value={data.redis.keyspace_hits} /></Col>
                <Col span={8}><Statistic title={t('monitor.redisMisses')} value={data.redis.keyspace_misses} /></Col>
              </Row>
            )}
            <p className="monitor-note">{data.redis?.version}</p>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('monitor.api')} loading={loading && !data.api}>
            <Row gutter={[16, 16]}>
              <Col span={8}><Statistic title={t('monitor.apiTotal')} value={data.api?.request_total ?? 0} /></Col>
              <Col span={8}><Statistic title={t('monitor.apiErrors')} value={data.api?.error_total ?? 0} /></Col>
              <Col span={8}><Statistic title={t('monitor.apiErrorRate')} value={data.api?.error_rate ?? 0} suffix="%" /></Col>
            </Row>
            <p className="monitor-note">{t('monitor.apiNote')}</p>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default MonitorPage;
