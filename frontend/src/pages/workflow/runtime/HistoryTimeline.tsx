import { tx, useLocale } from '@/i18n/text';
import { Empty, Timeline } from 'antd';
import type { WorkflowHistory } from '@/api/workflowRuntime';
import { actionLabels, dateLabel } from './meta';
import styles from './Runtime.module.css';
export default function HistoryTimeline({ items }: { items: WorkflowHistory[] }) {
  useLocale();
  if (!items.length)
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={tx('暂无操作记录')} />;
  return (
    <Timeline
      className={styles.history}
      items={items.map((item) => ({
        color: item.action === 'reject' ? 'red' : 'green',
        children: (
          <>
            <strong>
              {item.node_name || tx('流程')} · {actionLabels[item.action] || item.action}
            </strong>
            <p className={styles.muted}>
              {item.operator_name || tx('系统')} · {dateLabel(item.created_at)}
            </p>
            {item.comment && <p>{item.comment}</p>}
          </>
        ),
      }))}
    />
  );
}
