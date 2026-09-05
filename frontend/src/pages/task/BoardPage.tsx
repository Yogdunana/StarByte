import React, { useCallback, useEffect, useState } from 'react';
import { Card, Tag, message } from 'antd';
import StatusTag from '@/components/StatusTag/StatusTag';
import { getTaskList, updateTaskStatus } from '@/api/task';
import type { Task } from '@/types/api';
import { BOARD_COLUMNS, TaskPriorityMap } from './meta';
import DetailDrawer from './DetailDrawer';

const BoardPage: React.FC = () => {
  const [list, setList] = useState<Task[]>([]);
  const [dragId, setDragId] = useState<string | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);

  const load = useCallback(async () => {
    const res = await getTaskList({ page: 1, page_size: 100, sort_by: 'sort_order', sort_order: 'asc' });
    setList(res.list || []);
  }, []);

  useEffect(() => { void load(); }, [load]);

  const dropTo = async (status: number) => {
    if (!dragId) return;
    try {
      await updateTaskStatus(dragId, status);
      message.success('已更新状态');
      await load();
    } catch {
      /* interceptor already toasts */
    } finally {
      setDragId(null);
    }
  };

  return (
    <Card title="任务看板">
      <div style={{ display: 'flex', gap: 12, overflowX: 'auto', minHeight: 480 }}>
        {BOARD_COLUMNS.map((col) => {
          const cards = list.filter((t) => t.status === col.status);
          return (
            <div
              key={col.status}
              onDragOver={(e) => e.preventDefault()}
              onDrop={() => { void dropTo(col.status); }}
              style={{
                width: 240, flexShrink: 0, background: '#f5f5f5', borderRadius: 8, padding: 8,
              }}
            >
              <div style={{ fontWeight: 600, marginBottom: 8 }}>{col.title} ({cards.length})</div>
              {cards.map((t) => (
                <div
                  key={t.id}
                  draggable
                  onDragStart={() => setDragId(t.id)}
                  onClick={() => setDetailId(t.id)}
                  style={{
                    background: '#fff', borderRadius: 6, padding: 10, marginBottom: 8,
                    cursor: 'grab', boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                  }}
                >
                  <div style={{ fontWeight: 500 }}>{t.title}</div>
                  <div style={{ marginTop: 6 }}>
                    <StatusTag status={t.priority} mapping={TaskPriorityMap} />
                  </div>
                  <div style={{ color: '#888', fontSize: 12, marginTop: 6 }}>
                    {t.assignee?.name || '未分配'}
                    {t.due_date ? ` · ${t.due_date.slice(0, 10)}` : ''}
                  </div>
                  {(t.tags || []).slice(0, 3).map((tag) => <Tag key={tag} style={{ marginTop: 4 }}>{tag}</Tag>)}
                </div>
              ))}
            </div>
          );
        })}
      </div>
      <DetailDrawer taskId={detailId} open={!!detailId} onClose={() => setDetailId(null)} onChanged={() => { void load(); }} />
    </Card>
  );
};

export default BoardPage;
