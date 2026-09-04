import React from 'react';
import { Drawer, Button, Form, Input, Typography } from 'antd';
import { ExperimentOutlined } from '@ant-design/icons';
import type { FormInstance } from 'antd/es/form';
import type { NotificationTemplate } from '@/types/api';

const { Text, Paragraph } = Typography;
const { TextArea } = Input;

export interface TemplateTestDrawerProps {
  open: boolean;
  template: NotificationTemplate | null;
  form: FormInstance;
  loading: boolean;
  result: { title: string; content: string } | null;
  onClose: () => void;
  onRun: () => void;
}

const TemplateTestDrawer: React.FC<TemplateTestDrawerProps> = ({
  open,
  template,
  form,
  loading,
  result,
  onClose,
  onRun,
}) => (
  <Drawer
    title="测试模板渲染"
    open={open}
    onClose={onClose}
    width={560}
    extra={
      <Button
        type="primary"
        loading={loading}
        onClick={onRun}
        icon={<ExperimentOutlined />}
      >
        执行测试
      </Button>
    }
  >
    {template && (
      <div>
        <div style={{ marginBottom: 16 }}>
          <Text type="secondary">模板：</Text>
          <Text code>{template.code}</Text>
          <Text>（{template.name}）</Text>
        </div>
        <Form form={form} layout="vertical">
          <Form.Item label="测试变量（JSON 格式）" name="variables_raw">
            <TextArea
              rows={6}
              placeholder={'输入 JSON 格式的变量，如：\n{"TaskName": "完成需求文档", "DueDate": "2024-12-31"}'}
            />
          </Form.Item>
        </Form>
        {result && (
          <div style={{ marginTop: 16 }}>
            <Text strong>渲染结果：</Text>
            <div style={{ marginTop: 8, padding: 16, background: '#f5f5f5', borderRadius: 6 }}>
              <Text strong>标题：</Text>
              <Paragraph>{result.title}</Paragraph>
              <Text strong>正文：</Text>
              <Paragraph style={{ whiteSpace: 'pre-wrap' }}>{result.content}</Paragraph>
            </div>
          </div>
        )}
      </div>
    )}
  </Drawer>
);

export default TemplateTestDrawer;
