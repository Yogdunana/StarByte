import React, { useCallback, useEffect, useState } from 'react';
import {
  Button, Card, Descriptions, Form, Input, InputNumber, Modal, Space, Table, message,
} from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  createDimension, deleteDimension, endInterview, getDimensions, getEvaluationSummary,
  getInterviewList, startInterview, submitEvaluations, submitInterviewResult, updateDimension,
} from '@/api/interview';
import type { EvaluationDimension, EvaluationSummary, Interview } from '@/types/api';
import { InterviewStatusMap, ResultMap } from './meta';

const ScorePage: React.FC = () => {
  const canEval = usePermission('interview:evaluate');
  const canManage = usePermission('interview:manage');
  const [list, setList] = useState<Interview[]>([]);
  const [dims, setDims] = useState<EvaluationDimension[]>([]);
  const [current, setCurrent] = useState<Interview | null>(null);
  const [summary, setSummary] = useState<EvaluationSummary | null>(null);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    const [iv, d] = await Promise.all([
      getInterviewList({ page: 1, page_size: 50 }),
      getDimensions(),
    ]);
    setList(iv.list);
    setDims(d);
  }, []);

  useEffect(() => { void load(); }, [load]);

  const openScore = async (record: Interview) => {
    setCurrent(record);
    const s = await getEvaluationSummary(record.id);
    setSummary(s);
    form.setFieldsValue({
      items: dims.map((d) => ({ dimension: d.name, score: undefined, comment: '' })),
    });
  };

  return (
    <Card title="面试评分">
      <Table
        rowKey="id"
        dataSource={list}
        columns={[
          { title: '面试者', render: (_, r) => r.applicant.name },
          { title: '场次', dataIndex: 'session_title' },
          { title: '状态', dataIndex: 'status', render: (v: number) => <StatusTag status={v} mapping={InterviewStatusMap} /> },
          { title: '得分', dataIndex: 'score', render: (v?: number) => (v == null ? '-' : v) },
          { title: '结果', dataIndex: 'result', render: (v: number) => <StatusTag status={v} mapping={ResultMap} /> },
          {
            title: '操作',
            render: (_, r) => (
              <Space>
                {canEval && r.status === 1 && <Button type="link" size="small" onClick={() => startInterview(r.id).then(load)}>开始</Button>}
                {canEval && r.status === 2 && <Button type="link" size="small" onClick={() => endInterview(r.id).then(load)}>结束</Button>}
                {canEval && (r.status === 2 || r.status === 3) && (
                  <Button type="link" size="small" onClick={() => void openScore(r)}>评分</Button>
                )}
                {canManage && r.status === 3 && r.result === 0 && (
                  <Button type="link" size="small" onClick={() => {
                    Modal.confirm({
                      title: '提交结果',
                      content: (
                        <SelectResult onPick={async (result, comment) => {
                          await submitInterviewResult(r.id, result, comment);
                          message.success('结果已提交');
                          await load();
                        }} />
                      ),
                      okButtonProps: { style: { display: 'none' } },
                    });
                  }}>出结果</Button>
                )}
              </Space>
            ),
          },
        ]}
      />
      {canManage && (
        <Card size="small" title="评分维度" style={{ marginTop: 16 }}>
          <Space wrap>
            {dims.map((d) => (
              <span key={d.id}>
                {d.name}（权重 {d.weight}，满分 {d.max_score}）
                <Button type="link" size="small" onClick={() => {
                  const name = window.prompt('维度名称', d.name);
                  if (!name) return;
                  void updateDimension(d.id, { name }).then(load);
                }}>改</Button>
                <Button type="link" size="small" danger onClick={() => deleteDimension(d.id).then(load)}>删</Button>
              </span>
            ))}
            <Button size="small" onClick={() => {
              const name = window.prompt('新维度名称');
              if (!name) return;
              void createDimension({ name, weight: 0.2, max_score: 100 }).then(load);
            }}>新增维度</Button>
          </Space>
        </Card>
      )}
      <Modal
        title={current ? `评分 · ${current.applicant.name}` : '评分'}
        open={!!current}
        onCancel={() => setCurrent(null)}
        onOk={() => form.submit()}
        width={720}
      >
        {summary && (
          <Descriptions size="small" column={3} style={{ marginBottom: 12 }}>
            <Descriptions.Item label="平均分">{summary.average_score}</Descriptions.Item>
            <Descriptions.Item label="加权分">{summary.weighted_score}</Descriptions.Item>
            <Descriptions.Item label="已评人数">{summary.evaluations.length}</Descriptions.Item>
          </Descriptions>
        )}
        <Form
          form={form}
          onFinish={async (values: { items: Array<{ dimension: string; score: number; comment: string }> }) => {
            if (!current) return;
            await submitEvaluations(current.id, values.items);
            message.success('评分已提交');
            setCurrent(null);
            await load();
          }}
        >
          <Form.List name="items">
            {(fields) => fields.map((field) => (
              <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
                <Form.Item {...field} name={[field.name, 'dimension']} rules={[{ required: true }]}>
                  <Input disabled style={{ width: 120 }} />
                </Form.Item>
                <Form.Item {...field} name={[field.name, 'score']} rules={[{ required: true, message: '请打分' }]}>
                  <InputNumber min={0} max={100} />
                </Form.Item>
                <Form.Item {...field} name={[field.name, 'comment']}>
                  <Input placeholder="评语" style={{ width: 280 }} />
                </Form.Item>
              </Space>
            ))}
          </Form.List>
        </Form>
      </Modal>
    </Card>
  );
};

const SelectResult: React.FC<{ onPick: (r: 1 | 2 | 3, c: string) => Promise<void> }> = ({ onPick }) => {
  const [comment, setComment] = useState('');
  return (
    <Space direction="vertical" style={{ width: '100%', marginTop: 8 }}>
      <Input.TextArea rows={2} placeholder="评语" value={comment} onChange={(e) => setComment(e.target.value)} />
      <Space>
        <Button type="primary" onClick={() => void onPick(1, comment)}>通过</Button>
        <Button danger onClick={() => void onPick(2, comment)}>不通过</Button>
        <Button onClick={() => void onPick(3, comment)}>待定</Button>
      </Space>
    </Space>
  );
};

export default ScorePage;
