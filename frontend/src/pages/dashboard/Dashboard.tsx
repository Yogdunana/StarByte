import { Alert, Button, Card, Empty, Skeleton, Tag } from 'antd';
import { ArrowRightOutlined, CalendarOutlined, CheckOutlined, FileTextOutlined, FolderOutlined, NotificationOutlined, ReloadOutlined, ScheduleOutlined, TeamOutlined } from '@ant-design/icons';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import GlassOrb from '@/components/GlassOrb/GlassOrb';
import { usePermission } from '@/hooks/usePermission';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { selectUnreadCount } from '@/store/slices/notificationSlice';
import { formatDateTime } from '@/utils/format';
import { useWorkspace } from './useWorkspace';
import styles from './Dashboard.module.css';

export default function Dashboard() {
  const user = useSelector(selectCurrentUser);
  const unread = useSelector(selectUnreadCount);
  const canReadStats = usePermission('stats:read');
  const canReadFiles = usePermission('file:read');
  const canReadTasks = usePermission('task:read');
  const { tasks, taskTotal, approvals, approvalTotal, interviews, applications, overview, loading, failed, reload } = useWorkspace(canReadStats);
  const date = new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' }).format(new Date());
  const name = user?.real_name || user?.username || '同学';
  const applicationStates: Record<number, string> = { 0: '待初审', 1: '审核中', 2: '面试中', 3: '已通过', 4: '未通过', 5: '待补材料' };
  const metrics = [
    { label: '等我处理的审批', value: approvalTotal, caption: '查看材料，记录正式决定', path: '/workflow/todo' },
    { label: '我的待办任务', value: taskTotal, caption: '让每一件事都有着落', path: '/task/my' },
    { label: '待参加的面试', value: failed.includes('面试安排') ? null : interviews.length, caption: '候选人与面试官安排', path: '/interview/my' },
    { label: '未读消息', value: unread, caption: '及时了解与你有关的事', path: '/notification/list' },
  ];
  return <div className={styles.page}>
    <section className={styles.hero} aria-labelledby="workspace-greeting">
      <div className={styles.heroCopy}>
        <span className={styles.eyebrow}>YOUR WORKSPACE / 我的协作空间</span>
        <h1 id="workspace-greeting">{name}，欢迎回来。</h1>
        <p>从今天的待办开始，把想法变成大家一起完成的事。</p>
        <div className={styles.heroActions}><Link to="/task/my" className={styles.heroLink}>查看我的任务 <ArrowRightOutlined /></Link>
          <Button aria-label="刷新工作台" icon={<ReloadOutlined />} loading={loading} onClick={() => void reload()} />
          <span className={styles.date}>{date}</span></div>
      </div><GlassOrb />
    </section>
    {failed.length > 0 && <Alert className={styles.alert} showIcon type="warning" message={`${failed.join('、')}暂时加载失败，请重试。`} />}
    <div className={styles.metrics}>
      {metrics.map((item, index) => <Link key={item.label} to={item.path} className={styles.metric}>
        <div className={styles.metricTop}><span>{item.label}</span><span className={styles.metricNumber}>0{index + 1}</span></div>
        {loading ? <Skeleton.Input active size="small" /> : <strong>{item.value ?? '—'}<span> 项</span></strong>}
        <div className={styles.metricBottom}><span>{item.caption}</span><ArrowRightOutlined /></div>
      </Link>)}
    </div>
    <div className={styles.workspace}>
      <section className={styles.queue} aria-labelledby="queue-title">
        <div className={styles.sectionHeader}><div><span className={styles.eyebrow}>REVIEW</span><h2>等我处理的审批</h2></div><Link to="/workflow/todo">全部审批 <ArrowRightOutlined /></Link></div>
        <Card className={styles.approvalCard}>
          {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('审批待办') ? <Empty description="审批暂时无法加载" /> : approvals.length === 0 ? <div className={styles.quiet}><CheckOutlined /><div><strong>没有待处理的审批</strong><p>需要你核对与签字的事项会显示在这里。</p></div></div> : approvals.map(item => <Link key={item.id} className={styles.taskRow} to={`/workflow/todo?task_id=${encodeURIComponent(item.id)}`}>
            <span className={styles.calendarIcon}><FileTextOutlined /></span><div className={styles.rowContent}><strong>{item.node_name}</strong><p>{item.definition_name}{item.due_date ? ` · 截止 ${formatDateTime(item.due_date, 'MM-DD HH:mm')}` : ''}</p></div><ArrowRightOutlined />
          </Link>)}
        </Card>
        <div className={styles.sectionHeader}><div><span className={styles.eyebrow}>FOCUS</span><h2 id="queue-title">接下来，要做的事</h2></div><Link to="/task/my">全部任务 <ArrowRightOutlined /></Link></div>
        <Card className={styles.taskCard}>
          {loading ? <Skeleton active paragraph={{ rows: 4 }} /> : failed.includes('我的任务') ? <Empty description="任务暂时无法加载" /> : tasks.length === 0 ?
            <div className={styles.caughtUp}><span className={styles.check}><CheckOutlined /></span><h3>目前没有待办任务</h3><p>新的任务分配后，会出现在这里。</p><Link to="/task/my">查看我的任务记录 <ArrowRightOutlined /></Link></div> :
            <div>{tasks.map((task, index) => <div key={task.id} className={styles.taskRow}>
              <span className={styles.rowNumber}>{String(index + 1).padStart(2, '0')}</span>
              <div className={styles.rowContent}><strong>{task.title}</strong><p>{task.due_date ? `截止 ${formatDateTime(task.due_date, 'MM-DD HH:mm')}` : '未设置截止日期'}</p></div>
              <Tag color={task.status === 1 ? 'processing' : 'default'}>{task.status === 1 ? '进行中' : task.status === 4 ? '已挂起' : '待处理'}</Tag>
            </div>)}</div>}
        </Card>
        <div className={styles.sectionHeader}><div><span className={styles.eyebrow}>UP NEXT</span><h2>面试安排</h2></div><Link to="/interview/my">查看安排 <ArrowRightOutlined /></Link></div>
        <Card className={styles.scheduleCard}>
          {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('面试安排') ? <Empty description="面试安排暂时无法加载" /> : interviews.length === 0 ?
            <div className={styles.quiet}><CalendarOutlined /><div><strong>暂无待参加的面试</strong><p>时间、地点与后续安排将在这里同步。</p></div></div> :
            interviews.slice(0, 3).map(item => <div className={styles.taskRow} key={item.id}><span className={styles.calendarIcon}><CalendarOutlined /></span><div className={styles.rowContent}><strong>{item.session_title || '面试安排'}</strong><p>{item.scheduled_time ? formatDateTime(item.scheduled_time, 'MM-DD HH:mm') : '时间待定'} · {item.location || '地点待定'}</p></div><Tag>{item.status === 2 ? '进行中' : item.status === 1 ? '已签到' : '待面试'}</Tag></div>)}
        </Card>
      </section>
      <aside className={styles.aside}>
        <section className={styles.quickAccess}><span className={styles.eyebrow}>SHORTCUTS</span><h2>常用入口</h2>
          <div className={styles.shortcuts}>
            <Link to="/member/application"><FileTextOutlined /><span>入会申请</span><ArrowRightOutlined /></Link>
            <Link to="/interview/my"><ScheduleOutlined /><span>我的面试</span><ArrowRightOutlined /></Link>
            <Link to="/internship/my"><TeamOutlined /><span>我的实习</span><ArrowRightOutlined /></Link>
            <Link to="/announcement/list"><NotificationOutlined /><span>公告中心</span><ArrowRightOutlined /></Link>
            {canReadTasks && <Link to="/task/board"><CheckOutlined /><span>任务看板</span><ArrowRightOutlined /></Link>}
            {canReadFiles && <Link to="/files"><FolderOutlined /><span>文件资料</span><ArrowRightOutlined /></Link>}
          </div>
        </section>
        <section className={styles.application}><span className={styles.eyebrow}>MY JOURNEY</span><h2>我的申请进度</h2>
          {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('申请进度') ? <p>进度暂时无法加载，请刷新重试。</p> : applications.length ? applications.slice(0, 2).map(item =>
            <div key={item.id} className={styles.applicationRow}><span>{item.applicant_type === 2 ? '干事申请' : '会员申请'}</span><Tag>{item.current_stage || applicationStates[item.status] || '待核验'}</Tag></div>) : <p>加入一个团队，从一次申请开始。提交后可以随时查看进度。</p>}
          <Link to="/member/application">{applications.length ? '查看申请详情' : '开始入会申请'} <ArrowRightOutlined /></Link>
        </section>
      </aside>
    </div>
    {canReadStats && overview && <section className={styles.overview}>
      <div><span className={styles.eyebrow}>ASSOCIATION</span><h2>协会概览</h2></div>
      <div><strong>{overview.total_members}</strong><span>正式成员</span></div>
      <div><strong>{overview.total_tasks_in_progress}</strong><span>进行中任务</span></div>
      <div><strong>{overview.total_meetings_this_month}</strong><span>本月会议</span></div>
      <Link to="/stats/overview">查看统计 <ArrowRightOutlined /></Link>
    </section>}
  </div>;
}
