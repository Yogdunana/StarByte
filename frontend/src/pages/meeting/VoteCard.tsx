import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Popconfirm, Progress, Radio, Skeleton, Space, Tag, message } from 'antd';
import { CheckCircleOutlined, LockOutlined, ReloadOutlined } from '@ant-design/icons';
import { castVote, closeVote, getVoteResult } from '@/api/meeting';
import type { MeetingVote, VoteResult } from '@/types/api';
import styles from './Voting.module.css';
interface Props { vote: MeetingVote; canManage: boolean; onRefresh: () => Promise<void> }
export default function VoteCard({ vote, canManage, onRefresh }: Props) {
  const [result, setResult] = useState<VoteResult | null>(null);
  const [failed, setFailed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [selected, setSelected] = useState<string>();
  const [submitted, setSubmitted] = useState(false);
  const [now, setNow] = useState(Date.now());
  const sequence = useRef(0);
  const invalidate = useCallback(() => { sequence.current += 1; }, []);
  const load = useCallback(async () => {
    const seq = ++sequence.current;
    try { const data = await getVoteResult(vote.id); if (seq !== sequence.current) return; setResult(data); setFailed(false); }
    catch { if (seq === sequence.current) setFailed(true); }
  }, [vote.id]);
  useEffect(() => { void load(); return invalidate; }, [load, invalidate]);
  const status = result?.status ?? vote.status;
  useEffect(() => {
    if (status !== 1) return;
    const timer = window.setInterval(() => { if (!document.hidden) void load(); }, 10000);
    const clock = window.setInterval(() => setNow(Date.now()), 1000);
    return () => { window.clearInterval(timer); window.clearInterval(clock); };
  }, [status, load]);
  const endTime = result?.end_time ?? vote.end_time;
  const remaining = endTime ? Math.max(0, Math.ceil((new Date(endTime).getTime() - now) / 1000)) : null;
  const open = status === 1 && remaining !== 0;
  const voted = vote.has_voted || submitted;
  const submit = async () => {
    if (!selected || busy || !open) return;
    setBusy(true);
    try { await castVote(vote.id, selected); setSubmitted(true); message.success('你的投票已记录'); await load(); await onRefresh(); }
    catch { await load(); }
    finally { setBusy(false); }
  };
  const close = async () => {
    setBusy(true);
    try { await closeVote(vote.id); message.success('投票已截止'); await load(); await onRefresh(); }
    catch { /* The API layer displays the concrete error. */ }
    finally { setBusy(false); }
  };
  return <article className={styles.voteCard}>
    <div className={styles.cardHeader}><div><Space size={4} wrap><Tag color={open ? 'green' : 'default'}>{open ? '投票中' : status === 0 ? '未开始' : status === 3 ? '已取消' : '已截止'}</Tag><Tag>{vote.vote_type === 1 ? '每人一票' : '职务加权'}</Tag>{vote.is_anonymous && <Tag icon={<LockOutlined />}>匿名</Tag>}</Space><h3>{vote.title}</h3></div>
      {open && <span className={styles.countdown}>{remaining === null ? '手动截止' : `${Math.floor(remaining / 60)}:${String(remaining % 60).padStart(2, '0')}`}</span>}
    </div>
    {vote.description && <p className={styles.description}>{vote.description}</p>}
    {open && !voted && vote.can_vote && <div className={styles.ballot}>
      <Radio.Group aria-label={`${vote.title}的投票选项`} value={selected} onChange={event => setSelected(event.target.value)} disabled={busy} className={styles.choices}>
        {vote.options.map(option => <Radio value={option.key} key={option.key}>{option.label}</Radio>)}
      </Radio.Group>
      <div className={styles.submitRow}><Button type="primary" disabled={!selected} loading={busy} onClick={() => void submit()}>确认投票</Button><span>提交后不可更改</span></div>
      {vote.is_anonymous && <p className={styles.privacy}>参与记录与选票分开保存，不记录你与选项的对应关系。</p>}
    </div>}
    {open && !voted && !vote.can_vote && <p className={styles.privacy}>你不在本轮投票的参与名单中，可查看汇总结果。</p>}
    {!vote.electorate_frozen && <p className={styles.privacy}>历史投票未记录参与名单与权重快照，请人工核对计票依据。</p>}
    {vote.electorate_frozen && <p className={styles.privacy}>本轮参与名单已固定：{vote.eligible_count} 人；开始后的职务与配置变更不影响本轮。</p>}
    {voted && <p className={styles.receipt}><CheckCircleOutlined /> 你已完成投票{vote.is_anonymous ? ' · 匿名选票' : ''}</p>}
    <div className={styles.resultHeader}><h4>票数统计</h4><Button size="small" type="text" icon={<ReloadOutlined />} aria-label={`刷新${vote.title}结果`} onClick={() => void load()} /></div>
    {failed && <Alert type="warning" showIcon message="统计暂时无法更新" description={result ? '下方保留上次读取的数据。' : undefined} />}
    {!result && !failed ? <Skeleton active paragraph={{ rows: 2 }} /> : result && <>
      <p className={styles.resultTotal}>{result.total_voters} 人已参与{vote.vote_type === 2 ? ` · 总权重 ${result.total_weight}` : ''}</p>
      {result.results.map(item => {
        const percent = result.total_weight > 0 ? item.weight_total / result.total_weight * 100 : 0;
        return <div className={styles.resultRow} key={item.option_key}><div><span>{item.option_label}</span><strong>{item.count} 票{vote.vote_type === 2 ? ` · 权重 ${item.weight_total}` : ''}</strong></div><Progress status="normal" percent={Number(percent.toFixed(1))} strokeColor="var(--sb-brand)" /></div>;
      })}
    </>}
    {canManage && open && <div className={styles.manage}><Popconfirm title="现在截止投票？" description="截止后将停止接收新选票。" onConfirm={() => void close()} okText="确认截止" cancelText="继续投票"><Button disabled={busy}>截止投票</Button></Popconfirm></div>}
  </article>;
}
