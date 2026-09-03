import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Input,
  Space,
  Modal,
  Form,
  message,
  Tag,
  Select,
  Drawer,
  Typography,
  Popconfirm,
} from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  ExperimentOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  getNotificationTemplateList,
  createNotificationTemplate,
  updateNotificationTemplate,
  deleteNotificationTemplate,
  testNotificationTemplate,
} from '@/api/notification';
import type { NotificationTemplate, CreateNotificationTemplateParams } from '@/types/api';

const { Text, Paragraph } = Typography;
const { TextArea } = Input;

/** 通知渠道选项 */
const channelOptions = [
  { label: '站内消息', value: 'in_app' },
  { label: '邮件', value: 'email' },
  { label: 'WebSocket', value: 'websocket' },
];

/** 模板状态映射 */
const statusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '禁用' },
  1: { color: 'success', text: '启用' },
};

/** 渠道标签颜色映射 */
const channelColorMap: Record<string, string> = {
  in_app: 'blue',
  email: 'green',
  websocket: 'purple',
};

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

  // 搜索
  const handleSearch = () => {
    setPage(1);
    loadData();
  };

  // 重置
  const handleReset = () => {
    setKeyword('');
    setPage(1);
    loadData();
  };

  // 新增
  const handleAdd = () => {
    setEditingTemplate(null);
    form.resetFields();
    form.setFieldsValue({
      channels: ['in_app'],
      status: 1,
    });
    setModalVisible(true);
  };

  // 编辑
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

  // 删除
  const handleDelete = async (record: NotificationTemplate) => {
    try {
      await deleteNotificationTemplate(record.id);
      message.success('删除成功');
      loadData();
    } catch {
      message.error('删除失败');
    }
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      // variables_schema 需要转为对象
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
    } catch (error: any) {
      if (error.errorFields) return;
      message.error(editingTemplate ? '更新失败' : '创建失败');
    }
  };

  // 打开测试弹窗
  const handleTest = (record: NotificationTemplate) => {
    setTestingTemplate(record);
    setTestResult(null);
    testForm.resetFields();
    setTestDrawerVisible(true);
  };

  // 执行测试
  const handleRunTest = async () => {
    if (!testingTemplate) return;
    setTestLoading(true);
    try {
      const values = await testForm.validateFields();
      // 解析 JSON 格式的变量
      let variables: Record<string, unknown> = {};
      const rawStr = values.variables_raw as string;
      if (rawStr && rawStr.trim()) {
        try {
          variables = JSON.parse(rawStr);
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
    } catch (error: any) {
      if (error.errorFields) return;
      message.error('测试失败');
    } finally {
      setTestLoading(false);
    }
  };

  const columns: ColumnsType<NotificationTemplate> = [
    {
      title: '编码',
      dataIndex: 'code',
      key: 'code',
      width: 180,
      render: (code: string) => <Text code>{code}</Text>,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 160,
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (cat: string) => cat || '-',
    },
    {
      title: '渠道',
      dataIndex: 'channels',
      key: 'channels',
      width: 200,
      render: (channels: string[]) => (
        <Space wrap size={[4, 4]}>
          {channels?.map((ch) => (
            <Tag key={ch} color={channelColorMap[ch] || 'default'}>
              {ch}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        const info = statusMap[status] || statusMap[0];
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) =>
        new Date(time).toLocaleString('zh-CN', { hour12: false }),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: NotificationTemplate) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<ExperimentOutlined />}
            onClick={() => handleTest(record)}
          >
            测试
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除此模板？"
            onConfirm={() => handleDelete(record)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      {/* 搜索栏 */}
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

      {/* 表格 */}
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

      {/* 新增/编辑弹窗 */}
      <Modal
        title={editingTemplate ? '编辑模板' : '新增模板'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
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
            <Input
              placeholder="如: task_assigned"
              disabled={!!editingTemplate}
            />
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
              options={[
                { label: '系统', value: 'system' },
                { label: '任务', value: 'task' },
                { label: '会议', value: 'meeting' },
                { label: '审批', value: 'approval' },
                { label: '面试', value: 'interview' },
                { label: '其他', value: 'other' },
              ]}
            />
          </Form.Item>
          {editingTemplate && (
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

      {/* 测试模板抽屉 */}
      <Drawer
        title="测试模板渲染"
        open={testDrawerVisible}
        onClose={() => setTestDrawerVisible(false)}
        width={560}
        extra={
          <Button
            type="primary"
            loading={testLoading}
            onClick={handleRunTest}
            icon={<ExperimentOutlined />}
          >
            执行测试
          </Button>
        }
      >
        {testingTemplate && (
          <div>
            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">模板：</Text>
              <Text code>{testingTemplate.code}</Text>
              <Text>（{testingTemplate.name}）</Text>
            </div>
            <Form form={testForm} layout="vertical">
              <Form.Item label="测试变量（JSON 格式）" name="variables_raw">
                <TextArea
                  rows={6}
                  placeholder={'输入 JSON 格式的变量，如：\n{"TaskName": "完成需求文档", "DueDate": "2024-12-31"}'}
                />
              </Form.Item>
            </Form>
            {testResult && (
              <div style={{ marginTop: 16 }}>
                <Text strong>渲染结果：</Text>
                <div style={{ marginTop: 8, padding: 16, background: '#f5f5f5', borderRadius: 6 }}>
                  <Text strong>标题：</Text>
                  <Paragraph>{testResult.title}</Paragraph>
                  <Text strong>正文：</Text>
                  <Paragraph style={{ whiteSpace: 'pre-wrap' }}>{testResult.content}</Paragraph>
                </div>
              </div>
            )}
          </div>
        )}
      </Drawer>
    </Card>
  );
};

export default TemplateList;
