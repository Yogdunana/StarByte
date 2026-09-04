import React from 'react';
import { Modal, Form, Input, Select } from 'antd';
import type { FormInstance } from 'antd/es/form';
import type { NotificationTemplate } from '@/types/api';
import { channelOptions, categorySelectOptions } from './templateMeta';

const { TextArea } = Input;

export interface TemplateFormModalProps {
  open: boolean;
  editing: NotificationTemplate | null;
  form: FormInstance;
  onOk: () => void;
  onCancel: () => void;
}

const TemplateFormModal: React.FC<TemplateFormModalProps> = ({
  open,
  editing,
  form,
  onOk,
  onCancel,
}) => (
  <Modal
    title={editing ? '编辑模板' : '新增模板'}
    open={open}
    onOk={onOk}
    onCancel={onCancel}
    width={640}
    destroyOnClose
  >
    <Form form={form} layout="vertical">
      <Form.Item
        name="code"
        label="模板编码"
        rules={[
          { required: true, message: '请输入模板编码' },
          { pattern: /^[a-z][a-z0-9_]*$/, message: '只能包含小写字母、数字和下划线' },
        ]}
      >
        <Input placeholder="如: task_assigned" disabled={!!editing} />
      </Form.Item>
      <Form.Item
        name="name"
        label="模板名称"
        rules={[{ required: true, message: '请输入模板名称' }]}
      >
        <Input placeholder="如: 任务分配通知" />
      </Form.Item>
      <Form.Item
        name="title_template"
        label="标题模板"
        rules={[{ required: true, message: '请输入标题模板' }]}
        extra="使用 {{.变量名}} 引用变量，如：{{.TaskName}} 已分配给您"
      >
        <Input placeholder="如：{{.TaskName}} 已分配给您" />
      </Form.Item>
      <Form.Item
        name="body_template"
        label="正文模板"
        rules={[{ required: true, message: '请输入正文模板' }]}
        extra="支持多行文本，使用 {{.变量名}} 引用变量"
      >
        <TextArea
          rows={5}
          placeholder={'如：任务 "{{.TaskName}}" 已分配给您。\n截止时间：{{.DueDate}}\n请及时处理。'}
        />
      </Form.Item>
      <Form.Item
        name="channels"
        label="通知渠道"
        rules={[{ required: true, message: '请至少选择一个渠道' }]}
      >
        <Select
          mode="multiple"
          placeholder="选择通知渠道"
          options={channelOptions}
        />
      </Form.Item>
      <Form.Item name="category" label="分类">
        <Select
          placeholder="选择分类"
          allowClear
          options={categorySelectOptions}
        />
      </Form.Item>
      {editing && (
        <Form.Item name="status" label="状态">
          <Select
            options={[
              { label: '禁用', value: 0 },
              { label: '启用', value: 1 },
            ]}
          />
        </Form.Item>
      )}
    </Form>
  </Modal>
);

export default TemplateFormModal;
