import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button, Card, Descriptions, Form, Input, Progress, Select, Space, Table, Tag, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DownloadOutlined, PrinterOutlined } from '@ant-design/icons';
import {
  downloadExportFile, exportCsv, exportExcel, exportJson, exportPdf, exportTemplate,
  getExportTask, getExportTemplates,
} from '@/api/export';
import { usePermissions } from '@/hooks/usePermission';
import type {
  ExportFormat, ExportTableRequest, ExportTask, ExportTemplateInfo,
} from '@/types/api';
import './export.css';

const formatOptions: { value: ExportFormat; label: string; perm: string }[] = [
  { value: 'excel', label: 'Excel', perm: 'export:excel' },
  { value: 'csv', label: 'CSV', perm: 'export:csv' },
  { value: 'pdf', label: 'PDF', perm: 'export:pdf' },
  { value: 'json', label: 'JSON', perm: 'export:json' },
];

const sampleTemplateVars: Record<string, Record<string, string>> = {
  member_application: {
    Title: '入会申请表', Date: '2026-09-06', RealName: '张三',
    StudentNo: '20210001', Department: '技术部', Phone: '13800000000', Reason: '希望加入协会',
  },
  meeting_notice: {
    Title: '会议通知', Date: '2026-09-06', RealName: '李四',
    Subject: '招新评审', StartAt: '2026-09-07 19:00', Location: 'A101', Agenda: '面试安排',
  },
  internship_record: {
    Title: '实习记录', Date: '2026-09-06', RealName: '王五',
    Department: '技术部', Mentor: '赵六', Period: '2026-07 ~ 2026-08', Content: '后端开发', Comment: '表现良好',
  },
};

interface FormValues {
  format: ExportFormat;
  filename: string;
  title: string;
  columns: string;
  rows: string;
  sheet?: string;
  delimiter?: string;
}

function parseRows(text: string): string[][] {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => line.split(',').map((cell) => cell.trim()));
}

async function startExport(format: ExportFormat, body: ExportTableRequest): Promise<ExportTask> {
  switch (format) {
    case 'excel':
      return exportExcel(body);
    case 'csv':
      return exportCsv(body);
    case 'pdf':
      return exportPdf(body);
    default:
      return exportJson(body);
  }
}

const ExportPage: React.FC = () => {
  const [canExcel, canCsv, canPdf, canJson, canTpl, canDownload] = usePermissions([
    'export:excel', 'export:csv', 'export:pdf', 'export:json', 'export:template', 'export:download',
  ]);
  const allowedFormats = useMemo(
    () => formatOptions.filter((o) => (
      (o.value === 'excel' && canExcel)
      || (o.value === 'csv' && canCsv)
      || (o.value === 'pdf' && canPdf)
      || (o.value === 'json' && canJson)
    )),
    [canExcel, canCsv, canPdf, canJson],
  );
  const [form] = Form.useForm<FormValues>();
  const [templates, setTemplates] = useState<ExportTemplateInfo[]>([]);
  const [task, setTask] = useState<ExportTask | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    void getExportTemplates().then((list) => setTemplates(list || [])).catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!task?.task_id || task.status === 'done' || task.status === 'failed') return undefined;
    const id = task.task_id;
    const timer = window.setInterval(() => {
      void getExportTask(id).then(setTask).catch(() => undefined);
    }, 1000);
    return () => window.clearInterval(timer);
  }, [task?.task_id, task?.status]);

  const onSubmit = useCallback(async () => {
    const values = await form.validateFields();
    const columns = values.columns.split(',').map((s) => s.trim()).filter(Boolean);
    const rows = parseRows(values.rows);
    setSubmitting(true);
    try {
      const res = await startExport(values.format, {
        filename: values.filename,
        title: values.title,
        columns,
        rows,
        sheet: values.sheet,
        delimiter: values.delimiter,
      });
      setTask(res);
      if (res.status === 'done') message.success('导出完成');
      else message.info('已提交异步导出任务');
    } finally {
      setSubmitting(false);
    }
  }, [form]);

  const onPrint = async (tpl: ExportTemplateInfo) => {
    setSubmitting(true);
    try {
      const res = await exportTemplate(tpl.id, {
        filename: tpl.id,
        watermark: 'StarByte',
        vars: sampleTemplateVars[tpl.id] || { Title: tpl.name, Date: '2026-09-06', RealName: '张三' },
      });
      setTask(res);
      message.success('模板已生成');
    } finally {
      setSubmitting(false);
    }
  };

  const onDownload = async () => {
    if (!task?.file_id) return;
    await downloadExportFile(task.file_id, task.filename || 'export');
  };

  const tplColumns: ColumnsType<ExportTemplateInfo> = [
    { title: '模板', dataIndex: 'name', width: 140 },
    { title: '说明', dataIndex: 'description' },
    {
      title: '操作',
      width: 100,
      render: (_, row) => (
        <Button
          type="link"
          size="small"
          icon={<PrinterOutlined />}
          disabled={!canTpl}
          onClick={() => void onPrint(row)}
        >
          打印
        </Button>
      ),
    },
  ];

  return (
    <div>
      <div className="export-hero">
        <div>
          <h2>打印 / 报表导出</h2>
          <p>将列表数据导出为 Excel、CSV、PDF、JSON，或使用内置模板打印。大表异步任务可在此查询进度。</p>
        </div>
      </div>
      <div className="export-layout">
        <Card className="page-shell" title="表格导出">
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              format: allowedFormats[0]?.value || 'excel',
              filename: 'members',
              title: '会员名单',
              columns: '姓名,部门',
              rows: '张三,技术部\n李四,宣传部',
              sheet: 'Sheet1',
              delimiter: ',',
            }}
          >
            <Form.Item name="format" label="格式" rules={[{ required: true }]}>
              <Select options={allowedFormats} />
            </Form.Item>
            <Space wrap style={{ width: '100%' }}>
              <Form.Item name="filename" label="文件名" rules={[{ required: true }]}>
                <Input style={{ width: 180 }} />
              </Form.Item>
              <Form.Item name="title" label="标题">
                <Input style={{ width: 200 }} />
              </Form.Item>
              <Form.Item name="sheet" label="Sheet">
                <Input style={{ width: 140 }} />
              </Form.Item>
              <Form.Item name="delimiter" label="CSV 分隔符">
                <Input style={{ width: 100 }} />
              </Form.Item>
            </Space>
            <Form.Item name="columns" label="列（逗号分隔）" rules={[{ required: true }]}>
              <Input placeholder="姓名,部门" />
            </Form.Item>
            <Form.Item name="rows" label="行（每行一条，逗号分隔）" rules={[{ required: true }]}>
              <Input.TextArea rows={6} placeholder={'张三,技术部'} />
            </Form.Item>
            <Button type="primary" loading={submitting} disabled={!allowedFormats.length} onClick={() => void onSubmit()}>
              开始导出
            </Button>
          </Form>
        </Card>
        <Card className="page-shell" title="任务进度 / 模板打印">
          {task ? (
            <Descriptions column={1} size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label="任务">{task.task_id}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={task.status === 'done' ? 'green' : task.status === 'failed' ? 'red' : 'blue'}>
                  {task.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="进度">
                <Progress percent={task.progress} size="small" />
              </Descriptions.Item>
              {task.error && <Descriptions.Item label="错误">{task.error}</Descriptions.Item>}
              {task.file_id && canDownload && (
                <Descriptions.Item label="下载">
                  <Button type="primary" icon={<DownloadOutlined />} onClick={() => void onDownload()}>
                    下载 {task.filename}
                  </Button>
                </Descriptions.Item>
              )}
            </Descriptions>
          ) : (
            <p style={{ color: '#64748b' }}>提交导出后可在此查看进度。</p>
          )}
          <Table
            rowKey="id"
            size="small"
            pagination={false}
            columns={tplColumns}
            dataSource={templates}
          />
        </Card>
      </div>
    </div>
  );
};

export default ExportPage;
