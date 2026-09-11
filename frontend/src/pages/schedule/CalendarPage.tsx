import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Badge, Button, Calendar, Card, DatePicker, Form, Input, Modal, Select, Space, Table, Tag, message,
} from 'antd';
import type { CalendarProps } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { usePermission } from '@/hooks/usePermission';
import PageIntro from '@/components/PageIntro/PageIntro';
import {
  createCalendar,
  createEvent,
  deleteEvent,
  listCalendars,
  rangeEvents,
  updateEvent,
  type CalendarItem,
  type ScheduleEvent,
} from '@/api/schedule';
import './schedule.css';

type ViewMode = 'month' | 'week' | 'day' | 'agenda';

const CalendarPage: React.FC = () => {
  const { t } = useTranslation();
  const canCreate = usePermission('schedule:create');
  const canUpdate = usePermission('schedule:update');
  const canDelete = usePermission('schedule:delete');
  const [cals, setCals] = useState<CalendarItem[]>([]);
  const [events, setEvents] = useState<ScheduleEvent[]>([]);
  const [view, setView] = useState<ViewMode>('month');
  const [cursor, setCursor] = useState(dayjs());
  const [calendarId, setCalendarId] = useState<string>();
  const [loading, setLoading] = useState(false);
  const [openCal, setOpenCal] = useState(false);
  const [openEv, setOpenEv] = useState(false);
  const [editing, setEditing] = useState<ScheduleEvent | null>(null);
  const [calForm] = Form.useForm();
  const [evForm] = Form.useForm();

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
    setCals(res.list || []);
  }, []);

  const loadEvents = useCallback(async () => {
    setLoading(true);
    try {
      const list = await rangeEvents({
        start: windowRange.start.toISOString(),
        end: windowRange.end.toISOString(),
        calendar_id: calendarId,
      });
      setEvents(list || []);
    } finally {
      setLoading(false);
    }
  }, [calendarId, windowRange.end, windowRange.start]);

  useEffect(() => { void loadCals().catch(() => undefined); }, [loadCals]);
  useEffect(() => { void loadEvents().catch(() => undefined); }, [loadEvents]);

  const eventsOn = (day: Dayjs) => events.filter((e) => dayjs(e.start_at).isSame(day, 'day'));

  const monthCell: CalendarProps<Dayjs>['cellRender'] = (date) => {
    const items = eventsOn(date);
    if (!items.length) return null;
    return (
      <ul className="schedule-dots">
        {items.slice(0, 3).map((e) => (
          <li key={`${e.id}-${e.start_at}`}>
            <Badge color={e.color || e.calendar_color || '#2563eb'} text={e.title} />
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
      calendar_id: calendarId || cals[0]?.id,
      start_at: start,
      end_at: start.add(1, 'hour'),
      recurrence: 'none',
      remind_minutes: [15],
    });
    setOpenEv(true);
  };

  const openEdit = (ev: ScheduleEvent) => {
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
    if (view === 'agenda') return events;
    return events.filter((e) => {
      const d = dayjs(e.start_at);
      if (view === 'day') return d.isSame(cursor, 'day');
      if (view === 'week') return d.isAfter(cursor.startOf('week').subtract(1, 'second')) && d.isBefore(cursor.endOf('week').add(1, 'second'));
      return d.isSame(cursor, 'month');
    });
  }, [cursor, events, view]);

  return (
    <div className="schedule-page">
      <PageIntro
        eyebrow={t('schedule.eyebrow')}
        title={t('schedule.title')}
        description={t('schedule.desc')}
        actions={canCreate ? (
          <Space>
            <Button onClick={() => { calForm.resetFields(); setOpenCal(true); }}>{t('schedule.newCalendar')}</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()}>{t('schedule.newEvent')}</Button>
          </Space>
        ) : undefined}
      />
      <Card className="page-shell">
        <Space wrap style={{ marginBottom: 16 }}>
          <Select
            value={view}
            style={{ width: 120 }}
            data-testid="schedule-view"
            onChange={(v: ViewMode) => setView(v)}
            options={[
              { value: 'month', label: t('schedule.view.month') },
              { value: 'week', label: t('schedule.view.week') },
              { value: 'day', label: t('schedule.view.day') },
              { value: 'agenda', label: t('schedule.view.agenda') },
            ]}
          />
          <DatePicker value={cursor} onChange={(v) => v && setCursor(v)} />
          <Select
            allowClear
            placeholder={t('schedule.allCalendars')}
            style={{ minWidth: 180 }}
            value={calendarId}
            onChange={setCalendarId}
            options={cals.map((c) => ({ value: c.id, label: c.name }))}
          />
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
            onRow={(record) => ({ onClick: () => openEdit(record) })}
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
            const payload = {
              calendar_id: values.calendar_id,
              title: values.title,
              location: values.location,
              start_at: values.start_at.toISOString(),
              end_at: values.end_at.toISOString(),
              recurrence: values.recurrence,
              remind_minutes: values.remind_minutes,
            };
            const run = editing ? updateEvent(editing.id, payload) : createEvent(payload);
            void run.then(() => {
              message.success(t('common.saved'));
              setOpenEv(false);
              void loadEvents();
            });
          }}
        >
          <Form.Item name="calendar_id" label={t('schedule.calendar')}>
            <Select options={cals.map((c) => ({ value: c.id, label: c.name }))} />
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
    </div>
  );
};

export default CalendarPage;
