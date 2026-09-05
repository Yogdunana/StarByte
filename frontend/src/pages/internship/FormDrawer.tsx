import React, { useEffect, useState } from 'react';
import { Button, DatePicker, Drawer, Form, Input, Select, Space } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { getUserList } from '@/api/user';
import type { UserListItem } from '@/api/user';
import type { CreateInternshipParams, Internship } from '@/types/api';
import { InternshipTypeMap } from './meta';

interface FormValues {
  title: string;
  organization: string;
  description?: string;
  range?: [Dayjs, Dayjs?];
  type: number;
  mentor_id?: string;
  skills?: string[];
  achievements?: string;
}

interface InternshipFormPayload extends CreateInternshipParams {
  clear_end_date?: boolean;
}

interface Props {
  open: boolean;
  editing: Internship | null;
  onClose: () => void;
  onSubmit: (values: InternshipFormPayload) => Promise<void>;
}

function toCalendarDate(d: Dayjs): string {
  return `${d.format('YYYY-MM-DD')}T00:00:00Z`;
}

function fromCalendarDate(raw?: string): Dayjs | undefined {
  if (!raw) return undefined;
  const d = dayjs(raw.slice(0, 10));
  return d.isValid() ? d : undefined;
}

const FormDrawer: React.FC<Props> = ({ open, editing, onClose, onSubmit }) => {
  const [form] = Form.useForm<FormValues>();
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    void getUserList({ page: 1, page_size: 50 }).then((res) => setUsers(res.list || []));
  }, []);

  useEffect(() => {
    if (!open) return;
    if (editing) {
      form.setFieldsValue({
        title: editing.title,
        organization: editing.organization,
        description: editing.description,
        type: editing.type,
        mentor_id: editing.mentor?.id,
        skills: editing.skills,
        achievements: editing.achievements,
        range: [
          fromCalendarDate(editing.start_date),
          fromCalendarDate(editing.end_date),
        ],
      });
    } else {
      form.resetFields();
      form.setFieldsValue({ type: 0 });
    }
  }, [open, editing, form]);

  const handleOk = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const start = values.range?.[0];
      const end = values.range?.[1];
      if (!start) return;
      await onSubmit({
        title: values.title,
        organization: values.organization,
        description: values.description,
        type: values.type as 0 | 1 | 2,
        mentor_id: values.mentor_id,
        skills: values.skills || [],
        achievements: values.achievements,
        start_date: toCalendarDate(start),
        end_date: end ? toCalendarDate(end) : null,
        clear_end_date: !end,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Drawer
      title={editing ? '编辑实习' : '登记实习'}
      open={open}
      onClose={onClose}
      width={480}
      destroyOnClose
      footer={
        <Space style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" loading={loading} onClick={() => void handleOk()}>保存</Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Form.Item name="title" label="实习项目" rules={[{ required: true, max: 200 }]}>
          <Input placeholder="例如：StarByte 后端开发实习" />
        </Form.Item>
        <Form.Item name="organization" label="实习单位" rules={[{ required: true, max: 200 }]}>
          <Input placeholder="例如：计算机协会技术部" />
        </Form.Item>
        <Form.Item name="range" label="实习时间" rules={[{ required: true, message: '请选择开始日期' }]}>
          <DatePicker.RangePicker allowEmpty={[false, true]} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="type" label="类型" rules={[{ required: true }]}>
          <Select
            options={Object.entries(InternshipTypeMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
          />
        </Form.Item>
        <Form.Item name="mentor_id" label="指导老师">
          <Select
            allowClear
            showSearch
            optionFilterProp="label"
            options={users.map((u) => ({ value: u.id, label: u.real_name || u.username }))}
          />
        </Form.Item>
        <Form.Item name="skills" label="技能标签">
          <Select mode="tags" placeholder="输入后回车" />
        </Form.Item>
        <Form.Item name="description" label="实习说明">
          <Input.TextArea rows={3} />
        </Form.Item>
        <Form.Item name="achievements" label="实习成果">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Drawer>
  );
};

export default FormDrawer;
