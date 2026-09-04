import React, { useEffect, useState } from 'react';
import { Form, Input, Modal, Select, Table } from 'antd';
import { getFlowDefinitionList } from '@/api/workflow';
import type { CreateDefinitionPayload, FlowDefinitionDTO } from '@/types/workflow';

interface CreateModalProps {
  open: boolean;
  confirmLoading?: boolean;
  onCancel: () => void;
  onOk: (payload: CreateDefinitionPayload) => void;
}

export const CreateDefinitionModal: React.FC<CreateModalProps> = ({
  open,
  confirmLoading,
  onCancel,
  onOk,
}) => {
  const [form] = Form.useForm<CreateDefinitionPayload>();

  useEffect(() => {
    if (open) {
      form.setFieldsValue({
        key: `flow_${Date.now()}`,
        name: '',
        description: '',
        category: 'custom',
      });
    }
  }, [open, form]);

  return (
    <Modal
      title="创建流程定义"
      open={open}
      confirmLoading={confirmLoading}
      onCancel={onCancel}
      onOk={() => form.validateFields().then(onOk)}
    >
      <Form form={form} layout="vertical">
        <Form.Item name="name" label="流程名称" rules={[{ required: true, message: '请输入名称' }]}>
          <Input />
        </Form.Item>
        <Form.Item name="key" label="流程标识" rules={[{ required: true, message: '请输入标识' }]}>
          <Input />
        </Form.Item>
        <Form.Item name="category" label="分类">
          <Select
            options={[
              { label: '自定义', value: 'custom' },
              { label: '面试', value: 'interview' },
              { label: '会员', value: 'member' },
              { label: '任务', value: 'task' },
            ]}
          />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

interface OpenModalProps {
  open: boolean;
  onCancel: () => void;
  onSelect: (id: string) => void;
}

export const OpenDefinitionModal: React.FC<OpenModalProps> = ({ open, onCancel, onSelect }) => {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<FlowDefinitionDTO[]>([]);
  const [keyword, setKeyword] = useState('');

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    getFlowDefinitionList({ page: 1, page_size: 50, keyword })
      .then((res) => setList(res.list ?? []))
      .catch(() => setList([]))
      .finally(() => setLoading(false));
  }, [open, keyword]);

  return (
    <Modal title="打开已有流程" open={open} onCancel={onCancel} footer={null} width={720}>
      <Input.Search
        placeholder="搜索名称或标识"
        allowClear
        onSearch={setKeyword}
        style={{ marginBottom: 12 }}
      />
      <Table<FlowDefinitionDTO>
        rowKey="id"
        loading={loading}
        dataSource={list}
        size="small"
        pagination={false}
        onRow={(record) => ({
          onClick: () => onSelect(record.id),
        })}
        columns={[
          { title: '名称', dataIndex: 'name' },
          { title: '标识', dataIndex: 'key' },
          {
            title: '状态',
            dataIndex: 'status',
            render: (status: number) => (status === 1 ? '已发布' : status === 2 ? '已停用' : '草稿'),
          },
        ]}
      />
    </Modal>
  );
};
