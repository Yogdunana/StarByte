import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Empty, Space, Tag } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { createInternship, getMyInternships, updateInternship } from '@/api/internship';
import type { Internship } from '@/types/api';
import DetailDrawer from './DetailDrawer';
import FormDrawer from './FormDrawer';
import { InternshipStatusMap, InternshipTypeMap } from './meta';
import './internship.css';

const MyPage: React.FC = () => {
  const canCreate = usePermission('internship:create');
  const [list, setList] = useState<Internship[]>([]);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Internship | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);

  const load = useCallback(async () => {
    const rows = await getMyInternships();
    setList(rows || []);
  }, []);

  useEffect(() => { void load(); }, [load]);

  return (
    <div>
      <div className="intern-hero">
        <div>
          <h2>我的实习</h2>
          <p>记录自己的社团实习与校外实践，提交报告后即可进入培养档案。</p>
        </div>
        {canCreate && (
          <Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
            登记我的实习
          </Button>
        )}
      </div>
      {list.length === 0 ? (
        <Card className="page-shell"><Empty description="还没有实习记录" /></Card>
      ) : (
        <div className="intern-card-grid">
          {list.map((row) => (
            <Card
              key={row.id}
              className="intern-card"
              hoverable
              title={row.title}
              extra={<StatusTag status={row.status} mapping={InternshipStatusMap} />}
              onClick={() => setDetailId(row.id)}
            >
              <Space direction="vertical" size={8} style={{ width: '100%' }}>
                <div>{row.organization}</div>
                <div style={{ color: '#64748b' }}>
                  {row.start_date.slice(0, 10)} ~ {row.end_date ? row.end_date.slice(0, 10) : '进行中'} · {row.duration_days} 天
                </div>
                <Space wrap>
                  <StatusTag status={row.type} mapping={InternshipTypeMap} />
                  {(row.skills || []).map((s) => <Tag key={s}>{s}</Tag>)}
                </Space>
                {row.status === 0 && (
                  <Button size="small" onClick={(e) => { e.stopPropagation(); setEditing(row); setOpen(true); }}>
                    编辑
                  </Button>
                )}
              </Space>
            </Card>
          ))}
        </div>
      )}
      <FormDrawer
        open={open}
        editing={editing}
        onClose={() => setOpen(false)}
        onSubmit={async (values) => {
          if (editing) await updateInternship(editing.id, values);
          else await createInternship(values);
          setOpen(false);
          void load();
        }}
      />
      <DetailDrawer id={detailId} onClose={() => setDetailId(null)} onChanged={() => void load()} />
    </div>
  );
};

export default MyPage;
