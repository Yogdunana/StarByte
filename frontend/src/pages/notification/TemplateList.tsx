import React, { useState, useEffect, useCallback } from 'react';
import { Card, Table, Button, Input, Space, Form, message } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  getNotificationTemplateList,
  createNotificationTemplate,
  updateNotificationTemplate,
  deleteNotificationTemplate,
  testNotificationTemplate,
} from '@/api/notification';
import type { NotificationTemplate, CreateNotificationTemplateParams } from '@/types/api';
import { getTemplateColumns } from './templateColumns';
import TemplateFormModal from './TemplateFormModal';
import TemplateTestDrawer from './TemplateTestDrawer';

function isFormValidateError(error: unknown): boolean {
  return typeof error === 'object' && error !== null && 'errorFields' in error;
}

const TemplateList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<NotificationTemplate[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [modalVisible, setModalVisible] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<NotificationTemplate | null>(null);
  const [testDrawerVisible, setTestDrawerVisible] = useState(false);
  const [testingTemplate, setTestingTemplate] = useState<NotificationTemplate | null>(null);
  const [testResult, setTestResult] = useState<{ title: string; content: string } | null>(null);
  const [testLoading, setTestLoading] = useState(false);
  const [form] = Form.useForm();
  const [testForm] = Form.useForm();

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getNotificationTemplateList({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
      });
      setData(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载模板列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSearch = () => {
    setPage(1);
    loadData();
  };

  const handleReset = () => {
    setKeyword('');
    setPage(1);
    loadData();
  };

  const handleAdd = () => {
    setEditingTemplate(null);
    form.resetFields();
    form.setFieldsValue({
      channels: ['in_app'],
      status: 1,
    });
    setModalVisible(true);
  };

  const handleEdit = (record: NotificationTemplate) => {
    setEditingTemplate(record);
    form.setFieldsValue({
      code: record.code,
      name: record.name,
      title_template: record.title_template,
      body_template: record.body_template,
      channels: record.channels,
      category: record.category,
      status: record.status,
    });
    setModalVisible(true);
  };

  const handleDelete = async (record: NotificationTemplate) => {
    try {
      await deleteNotificationTemplate(record.id);
      message.success('删除成功');
      loadData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const params: CreateNotificationTemplateParams = {
        ...values,
        variables_schema: {},
      };
      if (editingTemplate) {
        await updateNotificationTemplate(editingTemplate.id, params);
        message.success('更新成功');
      } else {
        await createNotificationTemplate(params);
        message.success('创建成功');
      }
      setModalVisible(false);
      loadData();
    } catch (error: unknown) {
      if (isFormValidateError(error)) return;
      message.error(editingTemplate ? '更新失败' : '创建失败');
    }
  };

  const handleTest = (record: NotificationTemplate) => {
    setTestingTemplate(record);
    setTestResult(null);
    testForm.resetFields();
    setTestDrawerVisible(true);
  };

  const handleRunTest = async () => {
    if (!testingTemplate) return;
    setTestLoading(true);
    try {
      const values = await testForm.validateFields();
      let variables: Record<string, unknown> = {};
      const rawStr = values.variables_raw as string;
      if (rawStr && rawStr.trim()) {
        try {
          variables = JSON.parse(rawStr) as Record<string, unknown>;
        } catch {
          message.error('变量 JSON 格式不正确');
          setTestLoading(false);
          return;
        }
      }
      const result = await testNotificationTemplate(testingTemplate.id, {
        variables,
      });
      setTestResult(result);
    } catch (error: unknown) {
      if (isFormValidateError(error)) return;
      message.error('测试失败');
    } finally {
      setTestLoading(false);
    }
  };

  const columns = getTemplateColumns({
    onTest: handleTest,
    onEdit: handleEdit,
    onDelete: handleDelete,
  });

  return (
    <Card>
      <div style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
        <Input
          placeholder="搜索模板编码/名称"
          prefix={<SearchOutlined />}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 240 }}
          onPressEnter={handleSearch}
        />
        <Space>
          <Button type="primary" onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
        </Space>
        <div style={{ flex: 1 }} />
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          新增模板
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1100 }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />

      <TemplateFormModal
        open={modalVisible}
        editing={editingTemplate}
        form={form}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      />

      <TemplateTestDrawer
        open={testDrawerVisible}
        template={testingTemplate}
        form={testForm}
        loading={testLoading}
        result={testResult}
        onClose={() => setTestDrawerVisible(false)}
        onRun={handleRunTest}
      />
    </Card>
  );
};

export default TemplateList;
