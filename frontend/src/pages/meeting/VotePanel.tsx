import { tx, useLocale } from '@/i18n/text';
import { useState } from 'react';
import { Button, Drawer, Empty, Pagination } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { MeetingVote } from '@/types/api';
import VoteCard from './VoteCard';
import VoteCreateForm from './VoteCreateForm';
import styles from './Voting.module.css';
interface Props {
  canManage: boolean;
  meetingId: string;
  votes: MeetingVote[];
  onRefresh: () => Promise<void>;
}
export default function VotePanel({ meetingId, votes, onRefresh, canManage }: Props) {
  useLocale();
  const [creating, setCreating] = useState(false);
  const [page, setPage] = useState(1);
  const current = Math.min(page, Math.max(1, Math.ceil(votes.length / 4)));
  return (
    <section>
      <div className={styles.panelHeader}>
        <div>
          <h2>{tx('一起做决定')}</h2>
          <p>{tx('核对方案，提交选票，查看大家的选择。')}</p>
        </div>
        {canManage && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreating(true)}>
            {tx('发起投票')}
          </Button>
        )}
      </div>
      {votes.length ? (
        <div className={styles.voteList}>
          {votes.slice((current - 1) * 4, current * 4).map((vote) => (
            <VoteCard key={vote.id} vote={vote} canManage={canManage} onRefresh={onRefresh} />
          ))}
        </div>
      ) : (
        <Empty description={tx('还没有投票，发起一个需要大家共同决定的议题。')} />
      )}
      {votes.length > 4 && (
        <Pagination
          current={current}
          pageSize={4}
          total={votes.length}
          showSizeChanger={false}
          onChange={setPage}
          className={styles.pagination}
        />
      )}
      <Drawer
        open={creating}
        onClose={() => setCreating(false)}
        title={tx('发起新投票')}
        width="min(540px,100vw)"
        destroyOnClose
      >
        <VoteCreateForm
          meetingId={meetingId}
          onCreated={async () => {
            await onRefresh();
            setPage(1);
            setCreating(false);
          }}
        />
      </Drawer>
    </section>
  );
}
