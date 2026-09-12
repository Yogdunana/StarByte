import { tx, useLocale } from '@/i18n/text';
import { useEffect, useState } from 'react';
import { Alert, Avatar, Button, Empty, Popconfirm, Select, Space, Tag, message } from 'antd';
import { addAttendees, removeAttendee } from '@/api/meeting';
import { getUserList } from '@/api/user';
import type { MeetingAttendee } from '@/types/api';
import styles from './MeetingDetail.module.css';
interface Props {
  meetingId: string;
  organizerId: string;
  items: MeetingAttendee[];
  canManage: boolean;
  onChange: (items: MeetingAttendee[]) => void;
}
const positions: Record<string, string> = {
  get president() {
    return tx('会长');
  },
  get vice_president() {
    return tx('副会长');
  },
  get minister() {
    return tx('部长');
  },
  get vice_minister() {
    return tx('副部长');
  },
  get deputy() {
    return tx('副部长');
  },
  get officer() {
    return tx('干事');
  },
  get member() {
    return tx('会员');
  },
  get center_director() {
    return tx('中心主任');
  },
};
export default function AttendeePanel({
  meetingId,
  organizerId,
  items,
  canManage,
  onChange,
}: Props) {
  useLocale();
  const [options, setOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [query, setQuery] = useState('');
  useEffect(() => {
    if (!canManage) return;
    let active = true;
    const timer = window.setTimeout(() => {
      void getUserList({ page: 1, page_size: 50, keyword: query })
        .then((result) => {
          if (active) {
            setOptions(
              result.list.map((user) => ({
                value: user.id,
                label: user.real_name || user.username,
              })),
            );
            setFailed(false);
          }
        })
        .catch(() => {
          if (active) setFailed(true);
        });
    }, 300);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [canManage, query]);
  const add = async () => {
    setBusy(true);
    try {
      onChange(await addAttendees(meetingId, selected));
      setSelected([]);
      message.success(tx('参会人已添加'));
    } catch {
      /* Keep selections. */
    } finally {
      setBusy(false);
    }
  };
  const remove = async (userId: string) => {
    setBusy(true);
    try {
      await removeAttendee(meetingId, userId);
      onChange(items.filter((item) => item.user_id !== userId));
    } catch {
      /* Backend retains check-in and voting history. */
    } finally {
      setBusy(false);
    }
  };
  return (
    <section>
      <div className={styles.panelHeader}>
        <div>
          <h2>{tx('参会伙伴')}</h2>
          <p>
            {items.length} {tx('人受邀 ·')}
            {items.filter((item) => item.attended).length} {tx('人已签到')}
          </p>
        </div>
      </div>
      {canManage && (
        <div className={styles.invite}>
          {failed && (
            <Alert
              type="warning"
              message={tx('人员列表暂不可用或无读取权限，请联系管理员。')}
              showIcon
            />
          )}
          <Select
            mode="multiple"
            showSearch
            filterOption={false}
            onSearch={setQuery}
            aria-label={tx('选择参会人')}
            placeholder={tx('搜索姓名或账号，选择参会人')}
            value={selected}
            onChange={setSelected}
            options={options.filter(
              (option) => !items.some((item) => item.user_id === option.value),
            )}
            style={{ width: '100%' }}
          />
          <Button
            type="primary"
            disabled={!selected.length}
            loading={busy}
            onClick={() => void add()}
          >
            {tx('确认添加')}
          </Button>
        </div>
      )}
      {items.length ? (
        <div className={styles.people}>
          {items.map((item) => (
            <article key={item.id}>
              <Avatar>{item.name.slice(0, 1)}</Avatar>
              <div className={styles.person}>
                <strong>{item.name}</strong>
                <span>
                  {item.user_id === organizerId
                    ? tx('会议组织者')
                    : positions[item.position_code || ''] || tx('参会人')}
                </span>
              </div>
              <Space>
                <Tag color={item.attended ? 'green' : 'default'}>
                  {item.attended ? tx('已签到') : tx('未签到')}
                </Tag>
                {canManage && !item.attended && item.user_id !== organizerId && (
                  <Popconfirm
                    title={tx('移除这位参会人？')}
                    description={tx('有投票记录的参会人将保留。')}
                    okText={tx('移除')}
                    cancelText={tx('保留')}
                    onConfirm={() => remove(item.user_id)}
                  >
                    <Button type="text" danger size="small" disabled={busy}>
                      {tx('移除')}
                    </Button>
                  </Popconfirm>
                )}
              </Space>
            </article>
          ))}
        </div>
      ) : (
        <Empty description={tx('暂无参会人')} />
      )}
    </section>
  );
}
