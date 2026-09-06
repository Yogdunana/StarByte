import React, { useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  createDictType,
  deleteDictType,
  updateDictType,
  type DictType,
} from '@/api/dict';

interface Props {
  types: DictType[];
  loading: boolean;
  selectedId?: string;
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
  onSelect: (row: DictType) => void;
  onChanged: () => void;
}

const TypePanel: React.FC<Props> = ({
  types, loading, selectedId, canCreate, canUpdate, canDelete, onSelect, onChanged,
}) => {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<DictType | null>(null);
  const [form] = Form.useForm<Record<string, unknown>>();

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ sort_order: types.length + 1, status: true });
    setOpen(true);
  };

  const openEdit = (row: DictType) => {
    setEditing(row);
    form.setFieldsValue({
      name: row.name, description: row.description, sort_order: row.sort_order, status: row.status === 0,
    });
    setOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    const status = values.status ? 0 : 1;
    if (editing) {
      await updateDictType(editing.id, {
        name: String(values.name),
        description: String(values.description ?? ''),
        sort_order: Number(values.sort_order ?? 0),
        status: status as 0 | 1,
      });
      message.success('类型已更新');
    } else {
      await createDictType({
        code: String(values.code),
        name: String(values.name),
        description: String(values.description ?? ''),
        sort_order: Number(values.sort_order ?? 0),
        status: status as 0 | 1,
      });
      message.success('类型已创建');
    }
    setOpen(false);
    onChanged();
  };

  const columns: ColumnsType<DictType> = [
    { title: '名称', dataIndex: 'name', render: (v: string, r) => <Button type="link" onClick={() => onSelect(r)}>{v}</Button> },
    { title: '编码', dataIndex: 'code', width: 150 },
    {
      title: '标记',
      width: 90,
      render: (_, r) => (
        <Space size={4}>
          {r.is_system && <Tag color="blue">系统</Tag>}
          {r.status === 1 && <Tag>禁用</Tag>}
        </Space>
      ),
    },
    {
      title: '操作',
      width: 120,
      render: (_, r) => (
        <Space>
          {canUpdate && <Button type="link" size="small" onClick={() => openEdit(r)}>编辑</Button>}
          {canDelete && !r.is_system && (
            <Button type="link" size="small" danger onClick={() => {
              void deleteDictType(r.id).then(() => { message.success('已删除'); onChanged(); });
            }}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <Space style={{ marginBottom: 12 }}>
        {canCreate && <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建类型</Button>}
      </Space>
      <Table<DictType>
        rowKey="id"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={types}
        pagination={false}
        rowClassName={(r) => (r.id === selectedId ? 'ant-table-row-selected' : '')}
        onRow={(r) => ({ onClick: () => onSelect(r) })}
      />
      <Modal
        title={editing ? '编辑字典类型' : '新建字典类型'}
        open={open}
        onOk={() => void submit()}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          {!editing && (
            <Form.Item name="code" label="编码" rules={[{ required: true, message: '请输入编码' }]}>
              <Input placeholder="例如 task_priority" />
            </Form.Item>
          )}
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default TypePanel;
