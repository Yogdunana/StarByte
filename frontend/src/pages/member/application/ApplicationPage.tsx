import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Card, Grid, Input, Pagination, Select, Space, Table, Tabs } from 'antd';
import { getApplicationList, getMyApplications } from '@/api/member';
import type { MemberApplication, MemberApplicationStatus } from '@/types/api';
import { usePermission } from '@/hooks/usePermission';
import ApplicationForm from './ApplicationForm';
import ApplicationCards from './ApplicationCards';
import ReviewDrawer from './ReviewDrawer';
import { buildApplicationColumns } from './applicationColumns';
import MemberStats from '../stats/MemberStats';
import PageIntro from '@/components/PageIntro/PageIntro';

const ApplicationPage: React.FC = () => {
  useLocale();
  const mobile = !Grid.useBreakpoint().md;
  const mineRequest = useRef(0);
  const listRequest = useRef(0);
  const [mineError, setMineError] = useState(false);
  const [listError, setListError] = useState(false);
  const [mineLoading, setMineLoading] = useState(false);
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
    const request = ++mineRequest.current;
    setMineLoading(true);
    setMineError(false);
    try {
      const rows = await getMyApplications();
      if (request === mineRequest.current) setMyList(rows);
    } catch {
      if (request === mineRequest.current) setMineError(true);
    } finally {
      if (request === mineRequest.current) setMineLoading(false);
    }
  }, []);

  const loadList = useCallback(async () => {
    if (!canRead) return;
    const request = ++listRequest.current;
    setLoading(true);
    setListError(false);
    try {
      const res = await getApplicationList({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
        status,
      });
      if (request !== listRequest.current) return;
      setList(res.list);
      setTotal(res.total);
    } catch {
      if (request === listRequest.current) setListError(true);
    } finally {
      if (request === listRequest.current) setLoading(false);
    }
  }, [canRead, page, pageSize, keyword, status]);

  useEffect(() => {
    void loadMine();
    const sequence = mineRequest;
    return () => {
      sequence.current++;
    };
  }, [loadMine]);

  useEffect(() => {
    void loadList();
    const sequence = listRequest;
    return () => {
      sequence.current++;
    };
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
      <PageIntro
        eyebrow={tx('JOIN STARBYTE / 成员与招新')}
        title={tx('找到你的团队，一起开始')}
        description={tx(
          '普通会员提交资料后审核；干事参加面试并经过正式签字审批。提交后，可在这里跟踪每一步。',
        )}
      />
      <Card title={tx('入会申请')} styles={{ body: { paddingTop: 12 } }}>
        <Tabs
          items={[
            {
              key: 'form',
              label: tx('提交申请'),
              children: <ApplicationForm onSubmitted={refresh} />,
            },
            {
              key: 'mine',
              label: tx('我的申请'),
              children: mineError ? (
                <Alert
                  type="warning"
                  showIcon
                  message={tx('申请加载失败')}
                  action={<Button onClick={() => void loadMine()}>{tx('重试')}</Button>}
                />
              ) : mobile ? (
                <ApplicationCards rows={myList} loading={mineLoading} mine onOpen={open} />
              ) : (
                <Table
                  loading={mineLoading}
                  scroll={{ x: 840 }}
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
                    label: tx('申请列表'),
                    children: (
                      <>
                        <Space style={{ marginBottom: 16 }} wrap>
                          <Input.Search
                            placeholder={tx('姓名/学号')}
                            aria-label={tx('按姓名或学号查找申请')}
                            allowClear
                            onSearch={(v) => {
                              setKeyword(v);
                              setPage(1);
                            }}
                            style={{ width: 220 }}
                          />
                          <Select
                            allowClear
                            placeholder={tx('状态')}
                            aria-label={tx('申请状态')}
                            style={{ width: 140 }}
                            value={status}
                            onChange={(v) => {
                              setStatus(v);
                              setPage(1);
                            }}
                            options={[
                              { value: 0, label: tx('待审核') },
                              { value: 1, label: tx('审核中') },
                              { value: 2, label: tx('面试中') },
                              { value: 3, label: tx('通过') },
                              { value: 4, label: tx('拒绝') },
                              { value: 5, label: tx('补充材料') },
                            ]}
                          />
                        </Space>
                        {listError ? (
                          <Alert
                            type="warning"
                            showIcon
                            message={tx('申请列表加载失败')}
                            action={<Button onClick={() => void loadList()}>{tx('重试')}</Button>}
                          />
                        ) : mobile ? (
                          <>
                            <ApplicationCards
                              rows={list}
                              loading={loading}
                              review={canApprove}
                              onOpen={open}
                            />
                            <Pagination
                              simple
                              current={page}
                              pageSize={pageSize}
                              total={total}
                              onChange={setPage}
                              style={{ marginTop: 20 }}
                            />
                          </>
                        ) : (
                          <Table
                            scroll={{ x: 840 }}
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
                        )}
                      </>
                    ),
                  },
                  { key: 'stats', label: tx('统计'), children: <MemberStats /> },
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
