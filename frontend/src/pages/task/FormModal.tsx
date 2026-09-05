import React, { useEffect, useState } from 'react';
import { DatePicker, Form, Input, Modal, Select } from 'antd';
import dayjs from 'dayjs';
import { getUserList } from '@/api/user';
import { getMemberDepartments } from '@/api/member';
import type { MemberDepartmentOption, Task } from '@/types/api';
import type { UserListItem } from '@/api/user';

interface Props {
  open: boolean;
  editing: Task | null;
  onCancel: () => void;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

const FormModal: React.FC<Props> = ({ open, editing, onCancel, onSubmit }) => {
  const [form] = Form.useForm();
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [depts, setDepts] = useState<MemberDepartmentOption[]>([]);

  useEffect(() => {
    void getUserList({ page: 1, page_size: 50 }).then((res) => setUsers(res.list || []));
    void getMemberDepartments().then(setDepts).catch(() => setDepts([]));
  }, []);

  useEffect(() => {
    if (!open) return;
    if (editing) {
      form.setFieldsValue({
        title: editing.title,
        description: editing.description,
        priority: editing.priority,
        tags: editing.tags,
        due_date: editing.due_date ? dayjs(editing.due_date) : undefined,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({ priority: 1 });
    }
  }, [open, editing, form]);

  return (
    <Modal
      title={editing ? '编辑任务' : '新建任务'}
      open={open}
      onCancel={onCancel}
      onOk={() => form.submit()}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          await onSubmit({
            title: values.title,
            description: values.description,
            priority: values.priority,
            tags: values.tags,
            assignee_id: values.assignee_id,
            department_id: values.department_id,
            parent_id: values.parent_id,
            due_date: values.due_date ? (values.due_date as dayjs.Dayjs).toISOString() : undefined,
          });
        }}
      >
        <Form.Item name="title" label="标题" rules={[{ required: true }]}>
          <Input maxLength={200} />
        </Form.Item>
        <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 0, label: '低' },
              { value: 1, label: '中' },
              { value: 2, label: '高' },
              { value: 3, label: '紧急' },
            ]}
          />
        </Form.Item>
        {!editing && (
          <Form.Item name="assignee_id" label="负责人">
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              options={users.map((u) => ({ value: u.id, label: u.real_name || u.username }))}
            />
          </Form.Item>
        )}
        {!editing && (
          <Form.Item name="department_id" label="部门">
            <Select allowClear options={depts.map((d) => ({ value: d.id, label: d.name }))} />
          </Form.Item>
        )}
        <Form.Item name="due_date" label="截止日期">
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="tags" label="标签">
          <Select mode="tags" placeholder="输入后回车" />
        </Form.Item>
        <Form.Item name="description" label="说明">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default FormModal;
