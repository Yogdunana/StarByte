import { Button, Empty, Skeleton, Space, Tag } from 'antd';
import dayjs from 'dayjs';
import type { MemberApplication } from '@/types/api';
import { ApplicationStatusMap } from '../meta';
import styles from './ApplicationCards.module.css';

interface Props {
  rows: MemberApplication[];
  loading?: boolean;
  review?: boolean;
  mine?: boolean;
  onOpen: (row: MemberApplication, mode: 'view' | 'review' | 'resubmit') => void;
}
export default function ApplicationCards({ rows, loading, review, mine, onOpen }: Props) {
  if (loading) return <Skeleton active paragraph={{ rows: 4 }} />;
  if (!rows.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={mine ? '提交申请后，在这里查看每一步进度' : '没有符合条件的申请'} />;
  return <div className={styles.cards}>{rows.map(row => <article key={row.id} className={styles.card}>
    <div className={styles.heading}><h3>{row.real_name}</h3><Tag>{row.applicant_type === 2 ? '干事申请' : '会员申请'}</Tag></div>
    <p>{row.department_name || '未选择部门'} · {row.student_no}</p>
    <strong className={styles.stage}>{row.historical_review_required ? '历史待核验' : row.current_stage || ApplicationStatusMap[row.status]?.text}</strong>
    <p className={styles.date}>提交于 {dayjs(row.submitted_at).format('YYYY-MM-DD HH:mm')}</p>
    <Space wrap>
      <Button onClick={() => onOpen(row, 'view')}>查看进度</Button>
      {review && row.status !== 4 && (row.status !== 3 || row.admission_stage === 'probation') && <Button type="primary" onClick={() => onOpen(row, 'review')}>处理申请</Button>}
      {mine && row.status === 5 && !row.historical_review_required && <Button type="primary" onClick={() => onOpen(row, 'resubmit')}>补充材料</Button>}
    </Space>
  </article>)}</div>;
}
