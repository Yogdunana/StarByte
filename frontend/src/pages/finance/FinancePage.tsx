import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Col, DatePicker, Form, Input, InputNumber, Modal, Row, Select, Space, Statistic, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { usePermission } from '@/hooks/usePermission';
import {
  createFinanceRecord,
  deleteFinanceRecord,
  exportFinance,
  getFinanceCategories,
  getFinanceRecords,
  getFinanceSummary,
  updateFinanceRecord,
  type FinanceCategory,
  type FinanceRecord,
  type FinanceSummary,
} from '@/api/finance';

const FinancePage: React.FC = () => {
  const { t } = useTranslation();
  const canCreate = usePermission('finance:create');
  const canManage = usePermission('finance:manage');
  const [list, setList] = useState<FinanceRecord[]>([]);
  const [cats, setCats] = useState<FinanceCategory[]>([]);
  const [sum, setSum] = useState<FinanceSummary | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<FinanceRecord | null>(null);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [res, summary] = await Promise.all([
        getFinanceRecords({ page, page_size: 10 }),
        getFinanceSummary(),
      ]);
      setList(res.list || []);
      setTotal(res.total);
      setSum(summary);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => { void getFinanceCategories().then(setCats).catch(() => undefined); }, []);

  const dirLabel = (d: number) => (d === 2 ? t('finance.income') : t('finance.expense'));

  const columns: ColumnsType<FinanceRecord> = [
    { title: t('finance.title'), dataIndex: 'title' },
    { title: t('finance.category'), dataIndex: 'category_name', width: 120 },
    { title: t('finance.direction'), dataIndex: 'direction', width: 80, render: dirLabel },
    { title: t('finance.amount'), dataIndex: 'amount', width: 120, render: (v: number) => v.toFixed(2) },
    { title: t('finance.occurredAt'), dataIndex: 'occurred_at', width: 120, render: (v: string) => v?.slice(0, 10) },
    {
      title: t('common.actions'),
      width: 140,
      render: (_, record) => (
        <Space>
          {canManage && (
            <Button type="link" size="small" onClick={() => {
              setEditing(record);
              form.setFieldsValue({
                ...record,
                occurred_at: record.occurred_at ? dayjs(record.occurred_at.slice(0, 10)) : undefined,
              });
              setOpen(true);
            }}
            >{t('common.edit')}</Button>
          )}
          {canManage && (
            <Button type="link" size="small" danger onClick={() => {
              void deleteFinanceRecord(record.id).then(() => { message.success(t('common.deleted')); void load(); });
            }}
            >{t('common.delete')}</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} md={8}><Card><Statistic title={t('finance.incomeTotal')} value={sum?.income_total ?? 0} precision={2} /></Card></Col>
        <Col xs={24} md={8}><Card><Statistic title={t('finance.expenseTotal')} value={sum?.expense_total ?? 0} precision={2} /></Card></Col>
        <Col xs={24} md={8}><Card><Statistic title={t('finance.balance')} value={sum?.balance ?? 0} precision={2} /></Card></Col>
      </Row>
      <Card
        title={t('finance.records')}
        extra={(
          <Space>
            <Button onClick={() => { void exportFinance().catch(() => message.info(t('finance.exportReserved'))); }}>
              {t('finance.export')}
            </Button>
            {canCreate && (
              <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true); }}>
                {t('finance.create')}
              </Button>
            )}
          </Space>
        )}
      >
        <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={{ current: page, pageSize: 10, total, onChange: setPage }} />
      </Card>
      <Modal
        open={open}
        title={editing ? t('common.edit') : t('finance.create')}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values: { title: string; category_id: string; direction: number; amount: number; occurred_at: dayjs.Dayjs; remark?: string }) => {
            const payload = {
              title: values.title,
              category_id: values.category_id,
              direction: values.direction,
              amount: values.amount,
              occurred_at: `${values.occurred_at.format('YYYY-MM-DD')}T00:00:00Z`,
              remark: values.remark,
            };
            const run = editing
              ? updateFinanceRecord(editing.id, payload)
              : createFinanceRecord(payload);
            void run.then(() => { message.success(t('common.saved')); setOpen(false); void load(); });
          }}
        >
          <Form.Item name="title" label={t('finance.title')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category_id" label={t('finance.category')} rules={[{ required: true }]}>
            <Select options={cats.map((c) => ({ value: c.id, label: c.name }))} />
          </Form.Item>
          <Form.Item name="direction" label={t('finance.direction')} rules={[{ required: true }]}>
            <Select options={[{ value: 2, label: t('finance.income') }, { value: 1, label: t('finance.expense') }]} />
          </Form.Item>
          <Form.Item name="amount" label={t('finance.amount')} rules={[{ required: true }]}><InputNumber min={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="occurred_at" label={t('finance.occurredAt')} rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="remark" label={t('finance.remark')}><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FinancePage;
