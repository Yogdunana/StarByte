import { tx, useLocale } from '@/i18n/text';
import { useState } from 'react';
import { Button, Drawer, Empty, Form, Input, InputNumber, Popconfirm, Space, message } from 'antd';
import { ArrowDownOutlined, ArrowUpOutlined, PlusOutlined } from '@ant-design/icons';
import { addAgenda, deleteAgenda, sortAgendas, updateAgenda } from '@/api/meeting';
import type { MeetingAgenda } from '@/types/api';
import styles from './MeetingDetail.module.css';
interface Props {
  meetingId: string;
  items: MeetingAgenda[];
  canManage: boolean;
  onChange: (items: MeetingAgenda[]) => void;
}
interface Fields {
  title: string;
  content: string;
  presenter: string;
  duration: number;
}
export default function AgendaPanel({ meetingId, items, canManage, onChange }: Props) {
  useLocale();
  const [editing, setEditing] = useState<MeetingAgenda | 'new' | null>(null);
  const [busy, setBusy] = useState(false);
  const [form] = Form.useForm<Fields>();
  const edit = (item: MeetingAgenda | 'new') => {
    form.resetFields();
    if (item !== 'new') form.setFieldsValue(item);
    setEditing(item);
  };
  const save = async (values: Fields) => {
    if (!editing) return;
    setBusy(true);
    try {
      const data = { ...values, title: values.title.trim() };
      const item =
        editing === 'new'
          ? await addAgenda(meetingId, { ...data, sort_order: items.length + 1 })
          : await updateAgenda(meetingId, editing.id, data);
      onChange(
        editing === 'new'
          ? [...items, item]
          : items.map((row) => (row.id === item.id ? item : row)),
      );
      setEditing(null);
      message.success(tx('议程已保存'));
    } catch {
      /* Keep draft for correction or retry. */
    } finally {
      setBusy(false);
    }
  };
  const move = async (index: number, offset: number) => {
    const next = [...items];
    const target = index + offset;
    if (target < 0 || target >= next.length || busy) return;
    [next[index], next[target]] = [next[target], next[index]];
    setBusy(true);
    try {
      onChange(
        await sortAgendas(
          meetingId,
          next.map((item) => item.id),
        ),
      );
    } catch {
      /* Preserve original order. */
    } finally {
      setBusy(false);
    }
  };
  const remove = async (id: string) => {
    setBusy(true);
    try {
      await deleteAgenda(meetingId, id);
      onChange(items.filter((item) => item.id !== id));
    } catch {
      /* API displays failure. */
    } finally {
      setBusy(false);
    }
  };
  return (
    <section>
      <div className={styles.panelHeader}>
        <div>
          <h2>{tx('按议程，有序推进')}</h2>
          <p>
            {items.length} {tx('项议程 · 预计')}
            {items.reduce((sum, item) => sum + (item.duration || 0), 0)} {tx('分钟')}
          </p>
        </div>
        {canManage && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => edit('new')}>
            {tx('添加议程')}
          </Button>
        )}
      </div>
      {items.length ? (
        <ol className={styles.agendaList}>
          {items.map((item, index) => (
            <li key={item.id}>
              <span className={styles.order}>{String(index + 1).padStart(2, '0')}</span>
              <div className={styles.agendaBody}>
                <h3>{item.title}</h3>
                <p>
                  {item.presenter || tx('汇报人待确认')}
                  {item.duration > 0 ? tx(' · {{value0}} 分钟', { value0: item.duration }) : ''}
                </p>
                {item.content && <div className={styles.text}>{item.content}</div>}
                {canManage && (
                  <Space wrap className={styles.agendaActions}>
                    <Button
                      size="small"
                      disabled={busy || index === 0}
                      aria-label={tx('上移{{value0}}', { value0: item.title })}
                      icon={<ArrowUpOutlined />}
                      onClick={() => void move(index, -1)}
                    />
                    <Button
                      size="small"
                      disabled={busy || index === items.length - 1}
                      aria-label={tx('下移{{value0}}', { value0: item.title })}
                      icon={<ArrowDownOutlined />}
                      onClick={() => void move(index, 1)}
                    />
                    <Button size="small" disabled={busy} onClick={() => edit(item)}>
                      {tx('编辑')}
                    </Button>
                    <Popconfirm
                      title={tx('删除这项议程？')}
                      okText={tx('删除')}
                      cancelText={tx('保留')}
                      onConfirm={() => remove(item.id)}
                    >
                      <Button size="small" danger disabled={busy}>
                        {tx('删除')}
                      </Button>
                    </Popconfirm>
                  </Space>
                )}
              </div>
            </li>
          ))}
        </ol>
      ) : (
        <Empty description={tx('议程尚未安排')} />
      )}
      <Drawer
        title={editing === 'new' ? tx('添加议程') : tx('编辑议程')}
        open={editing !== null}
        onClose={() => {
          if (!busy) setEditing(null);
        }}
        width="min(520px,100vw)"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => void save(values)}
          initialValues={{ duration: 0 }}
        >
          <Form.Item
            name="title"
            label={tx('议程标题')}
            rules={[
              { required: true, whitespace: true, message: tx('请填写议程标题') },
              { max: 200 },
            ]}
          >
            <Input maxLength={200} />
          </Form.Item>
          <Form.Item name="presenter" label={tx('汇报人')}>
            <Input maxLength={100} />
          </Form.Item>
          <Form.Item name="duration" label={tx('预计时长（分钟）')}>
            <InputNumber min={0} precision={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="content" label={tx('议程说明')}>
            <Input.TextArea rows={5} />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={busy}>
            {tx('保存议程')}
          </Button>
        </Form>
      </Drawer>
    </section>
  );
}
