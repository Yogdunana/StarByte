import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Badge, Button, Calendar, Card, Checkbox, DatePicker, Form, Input, Modal, Radio, Select, Space, Table, Tag, Upload, message,
} from 'antd';
import type { CalendarProps, UploadFile } from 'antd';
import { ImportOutlined, PlusOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { usePermission } from '@/hooks/usePermission';
import PageIntro from '@/components/PageIntro/PageIntro';
import {
  createCalendar,
  createEvent,
  deleteEvent,
  googleCallback,
  googleConnect,
  googleDisconnect,
  googleStatus,
  googleSync,
  importICS,
  importTimetable,
  listCalendars,
  rangeEvents,
  updateEvent,
  type CalendarItem,
  type GoogleStatus,
  type ScheduleEvent,
} from '@/api/schedule';
import './schedule.css';

type ViewMode = 'month' | 'week' | 'day' | 'agenda';

const VIS_KEY = 'schedule.layerVisibility';

function loadVisibility(): Record<string, boolean> {
  try {
    return JSON.parse(localStorage.getItem(VIS_KEY) || '{}') as Record<string, boolean>;
  } catch {
    return {};
  }
}

const CalendarPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const canCreate = usePermission('schedule:create');
  const canUpdate = usePermission('schedule:update');
  const canDelete = usePermission('schedule:delete');
  const [cals, setCals] = useState<CalendarItem[]>([]);
  const [events, setEvents] = useState<ScheduleEvent[]>([]);
  const [view, setView] = useState<ViewMode>('month');
  const [cursor, setCursor] = useState(dayjs());
  const [visible, setVisible] = useState<Record<string, boolean>>(loadVisibility);
  const [loading, setLoading] = useState(false);
  const [openCal, setOpenCal] = useState(false);
  const [openEv, setOpenEv] = useState(false);
  const [openImport, setOpenImport] = useState(false);
  const [editing, setEditing] = useState<ScheduleEvent | null>(null);
  const [gStatus, setGStatus] = useState<GoogleStatus | null>(null);
  const [calForm] = Form.useForm();
  const [evForm] = Form.useForm();
  const [importForm] = Form.useForm();

  const windowRange = useMemo(() => {
    if (view === 'day') {
      return { start: cursor.startOf('day'), end: cursor.endOf('day') };
    }
    if (view === 'week') {
      return { start: cursor.startOf('week'), end: cursor.endOf('week') };
    }
    return { start: cursor.startOf('month').subtract(7, 'day'), end: cursor.endOf('month').add(7, 'day') };
  }, [cursor, view]);

  const loadCals = useCallback(async () => {
    const res = await listCalendars({ page: 1, page_size: 50 });
    const list = res.list || [];
    setCals(list);
    setVisible((prev) => {
      const next = { ...prev };
      list.forEach((c) => {
        if (next[c.id] === undefined) next[c.id] = true;
      });
      localStorage.setItem(VIS_KEY, JSON.stringify(next));
      return next;
    });
  }, []);

  const loadEvents = useCallback(async () => {
    setLoading(true);
    try {
      const list = await rangeEvents({
        start: windowRange.start.toISOString(),
        end: windowRange.end.toISOString(),
      });
      setEvents(list || []);
    } finally {
      setLoading(false);
    }
  }, [windowRange.end, windowRange.start]);

  const loadGoogle = useCallback(async () => {
    try {
      setGStatus(await googleStatus());
    } catch {
      setGStatus({ configured: false, connected: false });
    }
  }, []);

  useEffect(() => { void loadCals().catch(() => undefined); }, [loadCals]);
  useEffect(() => { void loadEvents().catch(() => undefined); }, [loadEvents]);
  useEffect(() => { void loadGoogle(); }, [loadGoogle]);

  useEffect(() => {
    const q = new URLSearchParams(window.location.search);
    const code = q.get('code');
    const state = q.get('state') || '';
    if (!code || (q.get('google') !== 'callback' && !state)) return;
    const lockKey = `schedule.google.bind:${code}`;
    if (sessionStorage.getItem(lockKey)) return;
    sessionStorage.setItem(lockKey, 'pending');
    const clearParams = () => {
      q.delete('code');
      q.delete('state');
      q.delete('google');
      const next = q.toString();
      window.history.replaceState({}, '', `${window.location.pathname}${next ? `?${next}` : ''}`);
    };
    void googleCallback({ code, state }).then(() => {
      sessionStorage.setItem(lockKey, 'done');
      message.success(t('schedule.google.connected'));
      void loadGoogle();
      void loadCals();
      void loadEvents();
      clearParams();
    }).catch(() => {
      sessionStorage.removeItem(lockKey);
      message.error(t('schedule.google.bindFailed'));
      clearParams();
    });
    // request 拦截器也会 toast 具体错误；这里补充「请重新连接」。
  }, [loadCals, loadEvents, loadGoogle, t]);

  const toggleLayer = (id: string, checked: boolean) => {
    setVisible((prev) => {
      const next = { ...prev, [id]: checked };
      localStorage.setItem(VIS_KEY, JSON.stringify(next));
      return next;
    });
  };

  const layerEvents = useMemo(
    () => events.filter((e) => visible[e.calendar_id] !== false),
    [events, visible],
  );

  const eventsOn = (day: Dayjs) => layerEvents.filter((e) => dayjs(e.start_at).isSame(day, 'day'));

  const monthCell: CalendarProps<Dayjs>['cellRender'] = (date) => {
    const items = eventsOn(date);
    if (!items.length) return null;
    return (
      <ul className="schedule-dots">
        {items.slice(0, 3).map((e) => (
          <li key={`${e.id}-${e.start_at}`}>
            <button
              type="button"
              className="schedule-event-chip"
              data-testid={`schedule-event-${e.id}`}
              onClick={(ev) => { ev.stopPropagation(); openEvent(e); }}
            >
              <Badge color={e.color || e.calendar_color || '#2563eb'} text={e.title} />
            </button>
          </li>
        ))}
        {items.length > 3 && <li className="schedule-more">+{items.length - 3}</li>}
      </ul>
    );
  };

  const openCreate = (day?: Dayjs) => {
    setEditing(null);
    evForm.resetFields();
    const start = (day || cursor).hour(10).minute(0).second(0);
    evForm.setFieldsValue({
      calendar_id: editableCals.find((c) => (c.source || 'personal') === 'personal')?.id || editableCals[0]?.id,
      start_at: start,
      end_at: start.add(1, 'hour'),
      recurrence: 'none',
      remind_minutes: [15],
    });
    setOpenEv(true);
  };

  const editableCals = useMemo(() => cals.filter((c) => c.can_edit), [cals]);

  const openEvent = (ev: ScheduleEvent) => {
    if (ev.link) {
      navigate(ev.link);
      return;
    }
    if (!ev.can_edit || !canUpdate) return;
    setEditing(ev);
    evForm.setFieldsValue({
      calendar_id: ev.calendar_id,
      title: ev.title,
      location: ev.location,
      start_at: dayjs(ev.start_at),
      end_at: dayjs(ev.end_at),
      recurrence: ev.recurrence || 'none',
    });
    setOpenEv(true);
  };

  const visibleList = useMemo(() => {
    if (view === 'agenda') return layerEvents;
    return layerEvents.filter((e) => {
      const d = dayjs(e.start_at);
      if (view === 'day') return d.isSame(cursor, 'day');
      if (view === 'week') return d.isAfter(cursor.startOf('week').subtract(1, 'second')) && d.isBefore(cursor.endOf('week').add(1, 'second'));
      return d.isSame(cursor, 'month');
    });
  }, [cursor, layerEvents, view]);

  return (
    <div className="schedule-page">
      <PageIntro
        eyebrow={t('schedule.eyebrow')}
        title={t('schedule.title')}
        description={t('schedule.desc')}
        actions={canCreate ? (
          <Space>
            <Button icon={<ImportOutlined />} onClick={() => { importForm.resetFields(); setOpenImport(true); }} data-testid="schedule-import">
              {t('schedule.import')}
            </Button>
            <Button onClick={() => { calForm.resetFields(); setOpenCal(true); }}>{t('schedule.newCalendar')}</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()}>{t('schedule.newEvent')}</Button>
          </Space>
        ) : undefined}
      />
      <div className="schedule-body">
        <aside className="schedule-layers" data-testid="schedule-layers">
          <div className="schedule-layers-title">{t('schedule.layers')}</div>
          {cals.map((c) => (
            <label key={c.id} className="schedule-layer">
              <span className="schedule-dot" style={{ background: c.color || '#2563eb' }} />
              <Checkbox
                checked={visible[c.id] !== false}
                onChange={(e) => toggleLayer(c.id, e.target.checked)}
              />
              <span className="schedule-layer-name">{c.name}</span>
              <Tag className="schedule-layer-tag">{t(`schedule.source.${c.source || 'personal'}`)}</Tag>
            </label>
          ))}
          {canCreate && (
            <div className="schedule-google">
              <div className="schedule-layers-title">{t('schedule.google.title')}</div>
              {!gStatus?.configured && <p className="schedule-hint">{t('schedule.google.notConfigured')}</p>}
              {gStatus?.configured && !gStatus.connected && (
                <Button size="small" onClick={() => {
                  void googleConnect().then((r) => { window.location.href = r.auth_url; });
                }}
                >{t('schedule.google.connect')}</Button>
              )}
              {gStatus?.connected && (
                <Space direction="vertical" size={6}>
                  <span>{gStatus.email || t('schedule.google.connected')}</span>
                  <Space>
                    <Button size="small" onClick={() => { void googleSync().then((r) => { message.success(t('schedule.imported', { n: r.event_count })); void loadCals(); void loadEvents(); }); }}>{t('schedule.google.sync')}</Button>
                    <Button size="small" onClick={() => { void googleDisconnect().then(() => { message.success(t('common.deleted')); void loadGoogle(); void loadCals(); }); }}>{t('schedule.google.disconnect')}</Button>
                  </Space>
                </Space>
              )}
            </div>
          )}
        </aside>
        <Card className="page-shell">
          <Space wrap style={{ marginBottom: 16 }}>
            <Radio.Group
              value={view}
              optionType="button"
              data-testid="schedule-view"
              onChange={(e) => setView(e.target.value)}
              options={[
                { value: 'month', label: t('schedule.view.month') },
                { value: 'week', label: t('schedule.view.week') },
                { value: 'day', label: t('schedule.view.day') },
                { value: 'agenda', label: t('schedule.view.agenda') },
              ]}
            />
            <DatePicker value={cursor} onChange={(v) => v && setCursor(v)} />
          </Space>
          {view === 'month' && (
            <Calendar
              value={cursor}
              onChange={setCursor}
              cellRender={monthCell}
            />
          )}
          {view !== 'month' && (
            <Table
              rowKey={(r) => `${r.id}-${r.start_at}`}
              loading={loading}
              dataSource={visibleList}
              pagination={view === 'agenda' ? { pageSize: 10 } : false}
              onRow={(record) => ({ onClick: () => openEvent(record) })}
              columns={[
                { title: t('schedule.eventTitle'), dataIndex: 'title' },
                { title: t('schedule.calendar'), dataIndex: 'calendar_name', width: 140 },
                {
                  title: t('schedule.time'),
                  width: 280,
                  render: (_, r) => `${dayjs(r.start_at).format('YYYY-MM-DD HH:mm')} – ${dayjs(r.end_at).format('HH:mm')}`,
                },
                { title: t('schedule.location'), dataIndex: 'location', width: 140, render: (v: string) => v || '-' },
                {
                  title: t('schedule.recurrenceLabel'),
                  dataIndex: 'recurrence',
                  width: 100,
                  render: (v: string) => <Tag>{t(`schedule.recurrence.${v || 'none'}`)}</Tag>,
                },
                {
                  title: t('common.actions'),
                  width: 80,
                  render: (_, r) => canDelete && r.can_edit ? (
                    <Button type="link" danger size="small" onClick={(e) => {
                      e.stopPropagation();
                      void deleteEvent(r.id).then(() => { message.success(t('common.deleted')); void loadEvents(); });
                    }}
                    >{t('common.delete')}</Button>
                  ) : null,
                },
              ]}
            />
          )}
        </Card>
      </div>

      <Modal
        open={openCal}
        title={t('schedule.newCalendar')}
        onCancel={() => setOpenCal(false)}
        onOk={() => calForm.submit()}
        destroyOnHidden
      >
        <Form
          form={calForm}
          layout="vertical"
          onFinish={(values: { name: string; calendar_type: number; description?: string }) => {
            void createCalendar(values).then(() => {
              message.success(t('common.saved'));
              setOpenCal(false);
              void loadCals();
            });
          }}
        >
          <Form.Item name="name" label={t('schedule.calendarName')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="calendar_type" label={t('schedule.calendarType')} initialValue={3} rules={[{ required: true }]}>
            <Select options={[
              { value: 2, label: t('schedule.type.2') },
              { value: 3, label: t('schedule.type.3') },
            ]}
            />
          </Form.Item>
          <Form.Item name="description" label={t('schedule.description')}><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>

      <Modal
        open={openEv}
        title={editing ? t('common.edit') : t('schedule.newEvent')}
        onCancel={() => setOpenEv(false)}
        onOk={() => evForm.submit()}
        destroyOnHidden
      >
        <Form
          form={evForm}
          layout="vertical"
          onFinish={(values: {
            calendar_id?: string; title: string; location?: string;
            start_at: Dayjs; end_at: Dayjs; recurrence?: string; remind_minutes?: number[];
          }) => {
            const startISO = values.start_at.toISOString();
            const endISO = values.end_at.toISOString();
            const timesUnchanged = !!editing
              && values.start_at.isSame(dayjs(editing.start_at))
              && values.end_at.isSame(dayjs(editing.end_at));
            const payload = {
              calendar_id: values.calendar_id,
              title: values.title,
              location: values.location,
              recurrence: values.recurrence,
              remind_minutes: values.remind_minutes,
              ...(timesUnchanged ? {} : { start_at: startISO, end_at: endISO }),
            };
            const run = editing
              ? updateEvent(editing.id, payload)
              : createEvent({ ...payload, start_at: startISO, end_at: endISO });
            void run.then(() => {
              message.success(t('common.saved'));
              setOpenEv(false);
              void loadEvents();
            });
          }}
        >
          <Form.Item name="calendar_id" label={t('schedule.calendar')}>
            <Select options={editableCals.map((c) => ({ value: c.id, label: c.name }))} />
          </Form.Item>
          <Form.Item name="title" label={t('schedule.eventTitle')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="location" label={t('schedule.location')}><Input /></Form.Item>
          <Form.Item name="start_at" label={t('schedule.start')} rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="end_at" label={t('schedule.end')} rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="recurrence" label={t('schedule.recurrenceLabel')} initialValue="none">
            <Select options={['none', 'daily', 'weekly', 'monthly'].map((v) => ({ value: v, label: t(`schedule.recurrence.${v}`) }))} />
          </Form.Item>
          {!editing && (
            <Form.Item name="remind_minutes" label={t('schedule.remind')}>
              <Select mode="multiple" options={[5, 15, 30, 60].map((m) => ({ value: m, label: t('schedule.minutes', { n: m }) }))} />
            </Form.Item>
          )}
        </Form>
      </Modal>

      <Modal
        open={openImport}
        title={t('schedule.import')}
        onCancel={() => setOpenImport(false)}
        onOk={() => importForm.submit()}
        destroyOnHidden
      >
        <Form
          form={importForm}
          layout="vertical"
          initialValues={{ kind: 'timetable' }}
          onFinish={(values: { kind: 'timetable' | 'ics'; semester_start?: Dayjs; calendar_id?: string; file?: UploadFile[] }) => {
            const file = values.file?.[0]?.originFileObj;
            if (!file) {
              message.error(t('schedule.importNeedFile'));
              return;
            }
            const run = values.kind === 'timetable'
              ? importTimetable(file, values.semester_start ? values.semester_start.format('YYYY-MM-DD') : '')
              : importICS(file, values.calendar_id);
            void run.then((r) => {
              message.success(t('schedule.imported', { n: r.event_count }));
              setOpenImport(false);
              void loadCals();
              void loadEvents();
            });
          }}
        >
          <Form.Item name="kind" label={t('schedule.importKind')}>
            <Radio.Group>
              <Radio.Button value="timetable">{t('schedule.importTimetable')}</Radio.Button>
              <Radio.Button value="ics">{t('schedule.importICS')}</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(p, c) => p.kind !== c.kind}>
            {({ getFieldValue }) => getFieldValue('kind') === 'timetable' ? (
              <Form.Item
                name="semester_start"
                label={t('schedule.semesterStart')}
                extra={t('schedule.semesterStartHint')}
                rules={[{ required: true }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            ) : (
              <Form.Item name="calendar_id" label={t('schedule.importTarget')} extra={t('schedule.importTargetHint')}>
                <Select allowClear options={editableCals.map((c) => ({ value: c.id, label: c.name }))} />
              </Form.Item>
            )}
          </Form.Item>
          <Form.Item name="file" label={t('schedule.importFile')} valuePropName="fileList" getValueFromEvent={(e: { fileList?: UploadFile[] }) => e?.fileList}>
            <Upload beforeUpload={() => false} maxCount={1} accept=".xlsx,.xls,.ics">
              <Button>{t('schedule.chooseFile')}</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CalendarPage;
