import React, { useEffect, useState } from 'react';
import { Button, Descriptions, Drawer, Input, Space, Tabs, Tag, message } from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  commentInternship,
  completeInternship,
  getInternshipDetail,
  submitInternshipReport,
} from '@/api/internship';
import type { Internship } from '@/types/api';
import { InternshipStatusMap, InternshipTypeMap } from './meta';

interface Props {
  id: string | null;
  onClose: () => void;
  onChanged: () => void;
}

const DetailDrawer: React.FC<Props> = ({ id, onClose, onChanged }) => {
  const canUpdate = usePermission('internship:update');
  const canEvaluate = usePermission('internship:evaluate');
  const [row, setRow] = useState<Internship | null>(null);
  const [report, setReport] = useState('');
  const [comment, setComment] = useState('');

  useEffect(() => {
    if (!id) {
      setRow(null);
      return;
    }
    void getInternshipDetail(id).then((data) => {
      setRow(data);
      setReport(data.report || '');
      setComment(data.mentor_comment || '');
    });
  }, [id]);

  const reload = async () => {
    if (!id) return;
    const data = await getInternshipDetail(id);
    setRow(data);
    onChanged();
  };

  return (
    <Drawer title="实习详情" open={!!id} onClose={onClose} width={560} destroyOnClose>
      {row && (
        <>
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="项目">{row.title}</Descriptions.Item>
            <Descriptions.Item label="成员">{row.user.name}</Descriptions.Item>
            <Descriptions.Item label="单位">{row.organization}</Descriptions.Item>
            <Descriptions.Item label="部门">{row.department?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="类型">
              <StatusTag status={row.type} mapping={InternshipTypeMap} />
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag status={row.status} mapping={InternshipStatusMap} />
            </Descriptions.Item>
            <Descriptions.Item label="周期">
              {row.start_date.slice(0, 10)} ~ {row.end_date ? row.end_date.slice(0, 10) : '进行中'}
            </Descriptions.Item>
            <Descriptions.Item label="时长">{row.duration_days} 天</Descriptions.Item>
            <Descriptions.Item label="导师">{row.mentor?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="技能">
              {(row.skills || []).map((s) => <Tag key={s}>{s}</Tag>)}
            </Descriptions.Item>
            <Descriptions.Item label="说明">{row.description || '-'}</Descriptions.Item>
            <Descriptions.Item label="成果">{row.achievements || '-'}</Descriptions.Item>
          </Descriptions>
          <Tabs
            style={{ marginTop: 16 }}
            items={[
              {
                key: 'report',
                label: '实习报告',
                children: (
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Input.TextArea rows={6} value={report} onChange={(e) => setReport(e.target.value)} />
                    {canUpdate && (
                      <Space>
                        <Button onClick={() => submitInternshipReport(row.id, report).then(() => {
                          message.success('报告已保存');
                          void reload();
                        })}>
                          保存报告
                        </Button>
                        {row.status === 0 && (
                          <Button type="primary" onClick={() => completeInternship(row.id, {
                            report, achievements: row.achievements,
                          }).then(() => {
                            message.success('已完成实习');
                            void reload();
                          })}>
                            完成实习
                          </Button>
                        )}
                      </Space>
                    )}
                  </Space>
                ),
              },
              {
                key: 'comment',
                label: '导师评价',
                children: (
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Input.TextArea rows={5} value={comment} onChange={(e) => setComment(e.target.value)} />
                    {canEvaluate && (
                      <Button type="primary" onClick={() => commentInternship(row.id, comment).then(() => {
                        message.success('评价已保存');
                        void reload();
                      })}>
                        提交评价
                      </Button>
                    )}
                  </Space>
                ),
              },
            ]}
          />
        </>
      )}
    </Drawer>
  );
};

export default DetailDrawer;
