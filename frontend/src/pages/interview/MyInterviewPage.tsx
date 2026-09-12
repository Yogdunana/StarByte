import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Card, Empty, Grid, Table, Tabs } from 'antd';
import { CalendarOutlined, EnvironmentOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import PageIntro from '@/components/PageIntro/PageIntro';
import StatusTag from '@/components/StatusTag/StatusTag';
import { getMyInterviews } from '@/api/interview';
import type { Interview } from '@/types/api';
import { InterviewStatusMap, ResultMap } from './meta';
import styles from './MyInterviewPage.module.css';

const groups = [
  {
    key: 'pending',
    get label() {
      return tx('待参加');
    },
    statuses: [0, 1],
  },
  {
    key: 'ongoing',
    get label() {
      return tx('进行中');
    },
    statuses: [2],
  },
  {
    key: 'done',
    get label() {
      return tx('已完成');
    },
    statuses: [3],
  },
  {
    key: 'all',
    get label() {
      return tx('全部');
    },
    statuses: [],
  },
];
const formatTime = (value?: string) =>
  value ? dayjs(value).format(tx('M月D日 HH:mm')) : tx('时间待安排');

const MyInterviewPage: React.FC = () => {
  useLocale();
  const [tab, setTab] = useState('pending');
  const [list, setList] = useState<Interview[]>([]);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState(false);
  const sequence = useRef(0);
  const screens = Grid.useBreakpoint();
  const load = useCallback(async () => {
    const request = ++sequence.current;
    setLoading(true);
    setFailed(false);
    try {
      const items = await getMyInterviews();
      if (request === sequence.current) setList(items);
    } catch {
      if (request === sequence.current) {
        setFailed(true);
        setList([]);
      }
    } finally {
      if (request === sequence.current) setLoading(false);
    }
  }, []);
  useEffect(() => {
    const counter = sequence;
    void load();
    return () => {
      counter.current++;
    };
  }, [load]);
  const group = groups.find((item) => item.key === tab)!;
  const visible = list.filter(
    (item) => !group.statuses.length || group.statuses.includes(item.status),
  );
  return (
    <>
      <PageIntro
        eyebrow={tx('MY INTERVIEWS / 我的面试')}
        title={tx('为下一次见面做好准备')}
        description={tx('查看候选人与面试官安排，及时了解时间、地点和后续进度。')}
        actions={
          <Button icon={<ReloadOutlined />} loading={loading} onClick={() => void load()}>
            {tx('刷新安排')}
          </Button>
        }
      />
      {failed && (
        <Alert
          type="error"
          showIcon
          message={tx('面试安排暂时加载失败')}
          description={tx('请重试，已有安排不会受到影响。')}
          action={<Button onClick={() => void load()}>{tx('重试')}</Button>}
        />
      )}
      <Card className={styles.panel}>
        <Tabs
          activeKey={tab}
          onChange={setTab}
          items={groups.map((item) => ({ key: item.key, label: item.label }))}
        />
        {screens.md ? (
          <Table<Interview>
            rowKey="id"
            loading={loading}
            dataSource={visible}
            scroll={{ x: 780 }}
            locale={{
              emptyText: failed ? (
                tx('加载失败，请重试')
              ) : (
                <Empty description={tx('这里还没有面试安排')} />
              ),
            }}
            columns={[
              {
                title: tx('面试安排'),
                render: (_, item) => (
                  <>
                    <strong>{item.session_title || tx('面试安排')}</strong>
                    <div className={styles.secondary}>
                      {tx('候选人 ·')}
                      {item.applicant.name}
                    </div>
                  </>
                ),
              },
              { title: tx('时间'), dataIndex: 'scheduled_time', render: formatTime },
              {
                title: tx('地点'),
                dataIndex: 'location',
                render: (value) => value || tx('地点待安排'),
              },
              {
                title: tx('面试官'),
                render: (_, item) =>
                  item.evaluators.map((person) => person.name).join('、') || tx('待分配'),
              },
              {
                title: tx('状态'),
                dataIndex: 'status',
                render: (value) => <StatusTag status={value} mapping={InterviewStatusMap} />,
              },
              {
                title: tx('面试结果'),
                dataIndex: 'result',
                render: (value) => <StatusTag status={value} mapping={ResultMap} />,
              },
            ]}
          />
        ) : (
          <div className={styles.cards} aria-busy={loading}>
            {loading ? (
              <p>{tx('正在加载安排…')}</p>
            ) : (
              visible.map((item) => (
                <article key={item.id} className={styles.interview}>
                  <div className={styles.heading}>
                    <h2>{item.session_title || tx('面试安排')}</h2>
                    <StatusTag status={item.status} mapping={InterviewStatusMap} />
                  </div>
                  <p>
                    {tx('候选人 ·')}
                    {item.applicant.name}
                  </p>
                  <p>
                    <CalendarOutlined /> {formatTime(item.scheduled_time)}
                  </p>
                  <p>
                    <EnvironmentOutlined /> {item.location || tx('地点待安排')}
                  </p>
                  <div className={styles.result}>
                    {tx('面试结果')}
                    <StatusTag status={item.result} mapping={ResultMap} />
                  </div>
                </article>
              ))
            )}
            {!loading && !visible.length && (
              <Empty description={failed ? tx('加载失败，请重试') : tx('这里还没有面试安排')} />
            )}
          </div>
        )}
      </Card>
      <p className={styles.note}>
        {tx('面试结果与正式录用审批分别记录。内部评分和评语不向候选人展示。')}
      </p>
    </>
  );
};
export default MyInterviewPage;
