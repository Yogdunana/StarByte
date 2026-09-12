import { tx, useLocale } from '@/i18n/text';
import React, { useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  createDictItem,
  deleteDictItem,
  updateDictItem,
  type DictItem,
  type DictType,
} from '@/api/dict';

interface Props {
  type?: DictType;
  items: DictItem[];
  loading: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
  onChanged: () => void;
}

const ItemPanel: React.FC<Props> = ({
  type,
  items,
  loading,
  canCreate,
  canUpdate,
  canDelete,
  onChanged,
}) => {
  useLocale();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<DictItem | null>(null);
  const [form] = Form.useForm<Record<string, unknown>>();

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ sort_order: items.length, status: true });
    setOpen(true);
  };

  const openEdit = (row: DictItem) => {
    setEditing(row);
    form.setFieldsValue({
      item_label: row.item_label,
      sort_order: row.sort_order,
      status: row.status === 0,
    });
    setOpen(true);
  };

  const submit = async () => {
    if (!type) return;
    const values = await form.validateFields();
    const status = values.status ? 0 : 1;
    if (editing) {
      await updateDictItem(editing.id, {
        item_label: String(values.item_label),
        sort_order: Number(values.sort_order ?? 0),
        status: status as 0 | 1,
      });
      message.success(tx('字典项已更新'));
    } else {
      await createDictItem({
        type_code: type.code,
        item_value: String(values.item_value),
        item_label: String(values.item_label),
        sort_order: Number(values.sort_order ?? 0),
        status: status as 0 | 1,
      });
      message.success(tx('字典项已创建'));
    }
    setOpen(false);
    onChanged();
  };

  const columns: ColumnsType<DictItem> = [
    { title: tx('标签'), dataIndex: 'item_label' },
    { title: tx('值'), dataIndex: 'item_value', width: 120 },
    { title: tx('排序'), dataIndex: 'sort_order', width: 70 },
    {
      title: tx('状态'),
      dataIndex: 'status',
      width: 80,
      render: (v: number) =>
        v === 0 ? <Tag color="success">{tx('启用')}</Tag> : <Tag>{tx('禁用')}</Tag>,
    },
    {
      title: tx('操作'),
      width: 120,
      render: (_, r) => (
        <Space>
          {canUpdate && (
            <Button type="link" size="small" onClick={() => openEdit(r)}>
              {tx('编辑')}
            </Button>
          )}
          {canDelete && (
            <Button
              type="link"
              size="small"
              danger
              onClick={() => {
                void deleteDictItem(r.id).then(() => {
                  message.success(tx('已删除'));
                  onChanged();
                });
              }}
            >
              {tx('删除')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <Space style={{ marginBottom: 12 }}>
        {canCreate && type && (
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
            {tx('新建项')}
          </Button>
        )}
        {!type && <span>{tx('请选择左侧字典类型')}</span>}
      </Space>
      <Table<DictItem>
        rowKey="id"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={items}
        pagination={false}
      />
      <Modal
        title={editing ? tx('编辑字典项') : tx('新建字典项')}
        open={open}
        onOk={() => void submit()}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          {!editing && (
            <Form.Item
              name="item_value"
              label={tx('值')}
              rules={[{ required: true, message: tx('请输入值') }]}
            >
              <Input placeholder={tx('写入业务枚举值，如 0')} />
            </Form.Item>
          )}
          <Form.Item
            name="item_label"
            label={tx('标签')}
            rules={[{ required: true, message: tx('请输入标签') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="sort_order" label={tx('排序')}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label={tx('启用')} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default ItemPanel;
