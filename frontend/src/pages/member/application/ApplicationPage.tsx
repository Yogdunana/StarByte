import React, { useCallback, useEffect, useState } from 'react';
import { Card, Input, Select, Space, Table, Tabs } from 'antd';
import { getApplicationList, getMyApplications } from '@/api/member';
import type { MemberApplication, MemberApplicationStatus } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import ApplicationForm from './ApplicationForm';
import ReviewDrawer from './ReviewDrawer';
import { buildApplicationColumns } from './applicationColumns';
import MemberStats from '../stats/MemberStats';
import '@/pages/internship/internship.css';

const ApplicationPage: React.FC = () => {
  const canRead = usePermission('member:read');
  const canApprove = usePermission('member:approve');
  const [myList, setMyList] = useState<MemberApplication[]>([]);
  const [list, setList] = useState<MemberApplication[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<MemberApplicationStatus | undefined>();
  const [loading, setLoading] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'view' | 'review' | 'resubmit'>('view');
  const [current, setCurrent] = useState<MemberApplication | null>(null);

  const loadMine = useCallback(async () => {
    const rows = await getMyApplications();
    setMyList(rows);
  }, []);

  const loadList = useCallback(async () => {
    if (!canRead) return;
    setLoading(true);
    try {
      const res = await getApplicationList({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
        status,
      });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [canRead, page, pageSize, keyword, status]);

  useEffect(() => {
    void loadMine();
  }, [loadMine]);

  useEffect(() => {
    void loadList();
  }, [loadList]);

  const open = (record: MemberApplication, mode: 'view' | 'review' | 'resubmit') => {
    setCurrent(record);
    setDrawerMode(mode);
    setDrawerOpen(true);
  };

  const refresh = () => {
    void loadMine();
    void loadList();
    setDrawerOpen(false);
  };

  return (
    <div>
      <div className="recruit-hero">
        <div>
          <h2>2026 秋季招新 · 入会申请</h2>
          <p>欢迎加入计算机协会。会员可直接提交申请；干事需经过一面 / 二面。材料提交后可在「我的申请」跟踪进度。</p>
        </div>
      </div>
    <Card title="入会申请" styles={{ body: { paddingTop: 12 } }}>
      <Tabs
        items={[
          {
            key: 'form',
            label: '提交申请',
            children: <ApplicationForm onSubmitted={refresh} />,
          },
          {
            key: 'mine',
            label: '我的申请',
            children: (
              <Table
                rowKey="id"
                dataSource={myList}
                columns={buildApplicationColumns({
                  onView: (r) => open(r, 'view'),
                  onResubmit: (r) => open(r, 'resubmit'),
                })}
                pagination={false}
              />
            ),
          },
          ...(canRead
            ? [
                {
                  key: 'list',
                  label: '申请列表',
                  children: (
                    <>
                      <Space style={{ marginBottom: 16 }} wrap>
                        <Input.Search
                          placeholder="姓名/学号"
                          allowClear
                          onSearch={(v) => {
                            setKeyword(v);
                            setPage(1);
                          }}
                          style={{ width: 220 }}
                        />
                        <Select
                          allowClear
                          placeholder="状态"
                          style={{ width: 140 }}
                          value={status}
                          onChange={(v) => {
                            setStatus(v);
                            setPage(1);
                          }}
                          options={[
                            { value: 0, label: '待审核' },
                            { value: 1, label: '审核中' },
                            { value: 2, label: '面试中' },
                            { value: 3, label: '通过' },
                            { value: 4, label: '拒绝' },
                            { value: 5, label: '补充材料' },
                          ]}
                        />
                      </Space>
                      <Table
                        rowKey="id"
                        loading={loading}
                        dataSource={list}
                        columns={buildApplicationColumns({
                          showReview: canApprove,
                          onView: (r) => open(r, 'view'),
                          onReview: (r) => open(r, 'review'),
                        })}
                        pagination={{
                          current: page,
                          pageSize,
                          total,
                          onChange: (p, ps) => {
                            setPage(p);
                            setPageSize(ps);
                          },
                        }}
                      />
                    </>
                  ),
                },
                { key: 'stats', label: '统计', children: <MemberStats /> },
              ]
            : []),
        ]}
      />
      <ReviewDrawer
        open={drawerOpen}
        record={current}
        mode={drawerMode}
        onClose={() => setDrawerOpen(false)}
        onDone={refresh}
      />
    </Card>
    </div>
  );
};

export default ApplicationPage;
