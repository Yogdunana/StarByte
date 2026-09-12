import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Badge, Button, Calendar, Card, Col, Empty, Form, Input, InputNumber, Modal, Row, Select, Space, Switch, Table, Tag, message } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import ReactECharts from 'echarts-for-react';
import { useTranslation } from 'react-i18next';
import { createLeaveType, getLeaveCalendar, getLeaveStats, getLeaveTypes, updateLeaveType } from '@/api/leave';
import type { LeaveApplication, LeaveStats, LeaveType, UpsertLeaveTypeParams } from '@/api/leave';
import { usePermission } from '@/hooks/usePermission';
import { leaveTypeLabel } from './meta';
import './leave.css';

const StatsPage: React.FC = () => {
  const { t } = useTranslation();
  const canApprove = usePermission('leave:approve');
  const [year, setYear] = useState(dayjs().year());
  const [month, setMonth] = useState(dayjs());
  const [stats, setStats] = useState<LeaveStats | null>(null);
  const [events, setEvents] = useState<LeaveApplication[]>([]);
  const [types, setTypes] = useState<LeaveType[]>([]);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<LeaveType | null>(null);
  const [form] = Form.useForm<UpsertLeaveTypeParams>();

  const load = useCallback(async () => {
    const start = month.startOf('month').format('YYYY-MM-DD');
    const end = month.endOf('month').format('YYYY-MM-DD');
    const [stat, cal, typeRows] = await Promise.all([
      getLeaveStats({ year }),
      getLeaveCalendar({ from: start, to: end }),
      getLeaveTypes(),
    ]);
    setStats(stat);
    setEvents(cal || []);
    setTypes(typeRows || []);
  }, [month, year]);

  useEffect(() => { void load(); }, [load]);

  const byDate = useMemo(() => {
    const map = new Map<string, LeaveApplication[]>();
    events.forEach((row) => {
      let cursor = dayjs(row.start_time).startOf('day');
      const last = dayjs(row.end_time).startOf('day');
      while (cursor.isBefore(last) || cursor.isSame(last, 'day')) {
        const key = cursor.format('YYYY-MM-DD');
        const list = map.get(key) || [];
        list.push(row);
        map.set(key, list);
        cursor = cursor.add(1, 'day');
      }
    });
    return map;
  }, [events]);

  return (
    <div>
      <div className="leave-hero">
        <div>
          <h2>{t('leave.statsTitle')}</h2>
          <p>{t('leave.statsHint')}</p>
        </div>
        <Space>
          <Select
            value={year}
            style={{ width: 120 }}
            onChange={setYear}
            options={[year - 1, year, year + 1].map((y) => ({ value: y, label: String(y) }))}
          />
          {canApprove && (
            <Button onClick={() => { setEditing(null); form.resetFields(); form.setFieldsValue({ enabled: true, deductible: false, default_days: 0 }); setOpen(true); }}>
              {t('leave.addType')}
            </Button>
          )}
        </Space>
      </div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} md={8}>
          <Card title={t('leave.personalStats')}>
            <div className="leave-stat-number">{stats?.personal.total ?? 0}</div>
            <div>{t('leave.personalDays', { days: stats?.personal.days ?? 0 })}</div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title={t('leave.scopeStats')}>
            <div className="leave-stat-number">{stats?.total ?? 0}</div>
            <div>{t('leave.listHint', { total: stats?.total ?? 0, pending: stats?.by_status?.pending ?? 0 })}</div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title={t('leave.deptCount')}>
            <div className="leave-stat-number">{stats?.departments?.length ?? 0}</div>
            <div>{t('leave.deptHint')}</div>
          </Card>
        </Col>
      </Row>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} lg={14}>
          <Card title={t('leave.monthChart')}>
            {!stats?.by_month?.length ? <Empty /> : (
              <ReactECharts
                style={{ height: 280 }}
                option={{
                  tooltip: { trigger: 'axis' },
                  xAxis: { type: 'category', data: stats.by_month.map((i) => i.month) },
                  yAxis: { type: 'value', name: t('leave.days') },
                  series: [{
                    type: 'bar',
                    data: stats.by_month.map((i) => i.days),
                    itemStyle: { color: '#2563eb', borderRadius: [8, 8, 0, 0] },
                  }],
                }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title={t('leave.deptChart')}>
            {!stats?.departments?.length ? <Empty /> : (
              <ReactECharts
                style={{ height: 280 }}
                option={{
                  tooltip: { trigger: 'item' },
                  series: [{
                    type: 'pie',
                    radius: ['36%', '68%'],
                    data: stats.departments.map((item) => ({
                      name: item.department_name || t('leave.noDept'),
                      value: item.days,
                    })),
                  }],
                }}
              />
            )}
          </Card>
        </Col>
      </Row>
      <Card title={t('leave.calendarTitle')} style={{ marginBottom: 16 }}>
        <Calendar
          value={month}
          onPanelChange={(value) => setMonth(value)}
          cellRender={(value: Dayjs) => {
            const rows = byDate.get(value.format('YYYY-MM-DD')) || [];
            return (
              <ul className="leave-cal-list">
                {rows.slice(0, 3).map((row) => (
                  <li key={row.id}>
                    <Badge
                      status={row.status === 'approved' ? 'success' : 'processing'}
                      text={`${row.applicant.name || ''} ${leaveTypeLabel(t, row.leave_type.code, row.leave_type.name)}`}
                    />
                  </li>
                ))}
              </ul>
            );
          }}
        />
      </Card>
      {canApprove && (
        <Card title={t('leave.typesTitle')}>
          <Table
            rowKey="id"
            dataSource={types}
            pagination={false}
            columns={[
              { title: t('leave.typeLabel'), render: (_, row) => leaveTypeLabel(t, row.code, row.name) },
              { title: t('leave.code'), dataIndex: 'code' },
              { title: t('leave.deductible'), render: (_, row) => row.deductible ? t('leave.yes') : t('leave.no') },
              { title: t('leave.defaultDays'), dataIndex: 'default_days' },
              { title: t('leave.enabled'), render: (_, row) => <Tag color={row.enabled ? 'success' : 'default'}>{row.enabled ? t('leave.on') : t('leave.off')}</Tag> },
              {
                title: t('common.actions'),
                render: (_, row) => (
                  <Button type="link" size="small" onClick={() => {
                    setEditing(row);
                    form.setFieldsValue(row);
                    setOpen(true);
                  }}>{t('common.edit')}</Button>
                ),
              },
            ]}
          />
        </Card>
      )}
      <Modal
        title={editing ? t('leave.editType') : t('leave.addType')}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={async () => {
          const values = await form.validateFields();
          if (editing) await updateLeaveType(editing.id, values);
          else await createLeaveType(values);
          message.success(t('leave.typeSaved'));
          setOpen(false);
          void load();
        }}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label={t('leave.typeName')} rules={[{ required: true, max: 50 }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label={t('leave.code')} rules={[{ required: true, max: 20 }]}>
            <Input disabled={Boolean(editing)} />
          </Form.Item>
          <Form.Item name="default_days" label={t('leave.defaultDays')}>
            <InputNumber min={0} max={366} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label={t('leave.description')}>
            <Input.TextArea rows={2} maxLength={255} />
          </Form.Item>
          <Form.Item name="deductible" label={t('leave.deductible')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="enabled" label={t('leave.enabled')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="sort_order" label={t('leave.sortOrder')}>
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default StatsPage;
