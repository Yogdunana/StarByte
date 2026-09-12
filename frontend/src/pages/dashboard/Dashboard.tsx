import { Alert, Button, Card, Skeleton, Tag } from 'antd';
import { ArrowRightOutlined, CalendarOutlined, CheckOutlined, FileTextOutlined, FolderOutlined, NotificationOutlined, ReloadOutlined, ScheduleOutlined, TeamOutlined } from '@ant-design/icons';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import EmptyState from '@/components/EmptyState/EmptyState';
import { usePermission } from '@/hooks/usePermission';
import { selectCurrentUser } from '@/store/slices/userSlice';
import { selectUnreadCount } from '@/store/slices/notificationSlice';
import { motion } from 'motion/react';
import { fadeUp, staggerEnter } from '@/motion/tokens';
import { formatDateTime } from '@/utils/format';
import { useFeature } from '@/hooks/useFeature';
import { useWorkspace } from './useWorkspace';
import styles from './Dashboard.module.css';

const FAIL_KEYS = ['approvals', 'tasks', 'interviews', 'applications', 'overview'] as const;

export default function Dashboard() {
  const { t, i18n } = useTranslation();
  const user = useSelector(selectCurrentUser);
  const unread = useSelector(selectUnreadCount);
  const canReadStats = usePermission('stats:read');
  const canReadFiles = usePermission('file:read');
  const canReadTasks = usePermission('task:read');
  const announcementFeed = useFeature('announcement.feed');
  const { tasks, taskTotal, approvals, approvalTotal, interviews, applications, announcements, overview, loading, failed, reload } = useWorkspace(canReadStats, announcementFeed.enabled);
  const locale = i18n.language === 'en-US' ? 'en-US' : 'zh-CN';
  const date = new Intl.DateTimeFormat(locale, { month: 'long', day: 'numeric', weekday: 'long' }).format(new Date());
  const name = user?.real_name || user?.username || t('dashboard.classmate');
  const applicationStates: Record<number, string> = {
    0: t('dashboard.appState.0'),
    1: t('dashboard.appState.1'),
    2: t('dashboard.appState.2'),
    3: t('dashboard.appState.3'),
    4: t('dashboard.appState.4'),
    5: t('dashboard.appState.5'),
  };
  const metrics = [
    { label: t('dashboard.metricApprovals'), value: approvalTotal, caption: t('dashboard.metricApprovalsHint'), path: '/workflow/todo' },
    { label: t('dashboard.metricTasks'), value: taskTotal, caption: t('dashboard.metricTasksHint'), path: '/task/my' },
    { label: t('dashboard.metricInterviews'), value: failed.includes('interviews') ? null : interviews.length, caption: t('dashboard.metricInterviewsHint'), path: '/interview/my' },
    { label: t('dashboard.metricUnread'), value: unread, caption: t('dashboard.metricUnreadHint'), path: '/notification/list' },
  ];
  const failedLabels = FAIL_KEYS.filter((key) => failed.includes(key)).map((key) => t(`dashboard.failed.${key}`));
  return (
    <motion.div className={styles.page} variants={staggerEnter} initial="hidden" animate="show">
      <motion.section className={styles.hero} aria-labelledby="workspace-greeting" variants={fadeUp}>
        <div className={styles.heroCopy}>
          <span className={styles.eyebrow}>{t('dashboard.eyebrow')}</span>
          <h1 id="workspace-greeting">{t('dashboard.greeting', { name })}</h1>
          <p>{t('dashboard.subtitle')}</p>
          <div className={styles.heroActions}>
            <Link to="/task/my" className={styles.heroLink}>{t('dashboard.viewTasks')} <ArrowRightOutlined /></Link>
            <Button aria-label={t('dashboard.refresh')} icon={<ReloadOutlined />} loading={loading} onClick={() => void reload()} />
            <span className={styles.date}>{date}</span>
          </div>
        </div>
      </motion.section>
      {failedLabels.length > 0 && (
        <motion.div variants={fadeUp}>
          <Alert className={styles.alert} showIcon type="warning" message={t('dashboard.loadPartial', { parts: failedLabels.join(t('dashboard.partSep')) })} />
        </motion.div>
      )}
      <motion.div className={styles.metrics} variants={staggerEnter}>
        {metrics.map((item, index) => (
          <motion.div key={item.label} className={styles.metricCell} variants={fadeUp}>
          <Link to={item.path} className={styles.metric}>
            <div className={styles.metricTop}>
              <span>{item.label}</span>
              <span className={styles.metricNumber}>0{index + 1}</span>
            </div>
            {loading ? <Skeleton.Input active size="small" /> : <strong>{item.value ?? '—'}<span> {t('dashboard.unit')}</span></strong>}
            <div className={styles.metricBottom}><span>{item.caption}</span><ArrowRightOutlined /></div>
          </Link>
          </motion.div>
        ))}
      </motion.div>
      <motion.div className={styles.workspace} variants={staggerEnter}>
        <motion.section className={styles.queue} aria-labelledby="queue-title" variants={fadeUp}>
          <div className={styles.sectionHeader}>
            <div><span className={styles.eyebrow}>{t('dashboard.reviewEyebrow')}</span><h2>{t('dashboard.reviewTitle')}</h2></div>
            <Link to="/workflow/todo">{t('dashboard.allApprovals')} <ArrowRightOutlined /></Link>
          </div>
          <Card className={styles.approvalCard}>
            {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('approvals') ? (
              <EmptyState title={t('dashboard.emptyApprovalsFail')} description={t('empty.retryHint')} />
            ) : approvals.length === 0 ? (
              <EmptyState icon={<CheckOutlined />} title={t('dashboard.emptyApprovals')} description={t('dashboard.emptyApprovalsHint')} />
            ) : approvals.map((item) => (
              <Link key={item.id} className={styles.taskRow} to={`/workflow/todo?task_id=${encodeURIComponent(item.id)}`}>
                <span className={styles.calendarIcon}><FileTextOutlined /></span>
                <div className={styles.rowContent}>
                  <strong>{item.node_name}</strong>
                  <p>{item.definition_name}{item.due_date ? ` · ${t('dashboard.due')} ${formatDateTime(item.due_date, 'MM-DD HH:mm')}` : ''}</p>
                </div>
                <ArrowRightOutlined />
              </Link>
            ))}
          </Card>
          <div className={styles.sectionHeader}>
            <div><span className={styles.eyebrow}>{t('dashboard.focusEyebrow')}</span><h2 id="queue-title">{t('dashboard.focusTitle')}</h2></div>
            <Link to="/task/my">{t('dashboard.allTasks')} <ArrowRightOutlined /></Link>
          </div>
          <Card className={styles.taskCard}>
            {loading ? <Skeleton active paragraph={{ rows: 4 }} /> : failed.includes('tasks') ? (
              <EmptyState title={t('dashboard.emptyTasksFail')} description={t('empty.retryHint')} />
            ) : tasks.length === 0 ? (
              <EmptyState
                icon={<CheckOutlined />}
                title={t('dashboard.emptyTasks')}
                description={t('dashboard.emptyTasksHint')}
                action={<Link to="/task/my">{t('dashboard.viewTaskHistory')} <ArrowRightOutlined /></Link>}
              />
            ) : (
              <div>
                {tasks.map((task, index) => (
                  <div key={task.id} className={styles.taskRow}>
                    <span className={styles.rowNumber}>{String(index + 1).padStart(2, '0')}</span>
                    <div className={styles.rowContent}>
                      <strong>{task.title}</strong>
                      <p>{task.due_date ? `${t('dashboard.due')} ${formatDateTime(task.due_date, 'MM-DD HH:mm')}` : t('dashboard.noDue')}</p>
                    </div>
                    <Tag color={task.status === 1 ? 'processing' : 'default'}>
                      {task.status === 1 ? t('dashboard.taskDoing') : task.status === 4 ? t('dashboard.taskHeld') : t('dashboard.taskTodo')}
                    </Tag>
                  </div>
                ))}
              </div>
            )}
          </Card>
          <div className={styles.sectionHeader}>
            <div><span className={styles.eyebrow}>{t('dashboard.nextEyebrow')}</span><h2>{t('dashboard.nextTitle')}</h2></div>
            <Link to="/interview/my">{t('dashboard.viewSchedule')} <ArrowRightOutlined /></Link>
          </div>
          <Card className={styles.scheduleCard}>
            {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('interviews') ? (
              <EmptyState title={t('dashboard.emptyInterviewsFail')} description={t('empty.retryHint')} />
            ) : interviews.length === 0 ? (
              <EmptyState icon={<CalendarOutlined />} title={t('dashboard.emptyInterviews')} description={t('dashboard.emptyInterviewsHint')} />
            ) : interviews.slice(0, 3).map((item) => (
              <div className={styles.taskRow} key={item.id}>
                <span className={styles.calendarIcon}><CalendarOutlined /></span>
                <div className={styles.rowContent}>
                  <strong>{item.session_title || t('dashboard.interviewFallback')}</strong>
                  <p>{item.scheduled_time ? formatDateTime(item.scheduled_time, 'MM-DD HH:mm') : t('dashboard.timeTbd')} · {item.location || t('dashboard.placeTbd')}</p>
                </div>
                <Tag>{item.status === 2 ? t('dashboard.ivDoing') : item.status === 1 ? t('dashboard.ivChecked') : t('dashboard.ivPending')}</Tag>
              </div>
            ))}
          </Card>
        </motion.section>
        <motion.aside className={styles.aside} variants={fadeUp}>
          <section className={styles.feed} aria-labelledby="announcement-feed-title">
            <span className={styles.eyebrow}>{t('dashboard.feedEyebrow')}</span>
            <h2 id="announcement-feed-title">{t('dashboard.feedTitle')}</h2>
            {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : announcements.length === 0 ? (
              <p>{t('dashboard.feedEmpty')}</p>
            ) : announcements.map((item) => (
              <Link key={item.id} className={styles.feedRow} to={`/announcement/${item.id}`}>
                <span>{item.pinned ? `${t('dashboard.feedPinned')} · ` : ''}{item.title}</span>
                <ArrowRightOutlined />
              </Link>
            ))}
            <Link to="/announcement/list">{t('dashboard.feedAll')} <ArrowRightOutlined /></Link>
          </section>
          <section className={styles.quickAccess}>
            <span className={styles.eyebrow}>{t('dashboard.shortcutsEyebrow')}</span>
            <h2>{t('dashboard.shortcutsTitle')}</h2>
            <div className={styles.shortcuts}>
              <Link to="/member/application"><FileTextOutlined /><span>{t('dashboard.shortcutApply')}</span><ArrowRightOutlined /></Link>
              <Link to="/interview/my"><ScheduleOutlined /><span>{t('dashboard.shortcutInterview')}</span><ArrowRightOutlined /></Link>
              <Link to="/internship/my"><TeamOutlined /><span>{t('dashboard.shortcutInternship')}</span><ArrowRightOutlined /></Link>
              <Link to="/announcement/list"><NotificationOutlined /><span>{t('dashboard.shortcutAnnounce')}</span><ArrowRightOutlined /></Link>
              {canReadTasks && <Link to="/task/board"><CheckOutlined /><span>{t('dashboard.shortcutBoard')}</span><ArrowRightOutlined /></Link>}
              {canReadFiles && <Link to="/files"><FolderOutlined /><span>{t('dashboard.shortcutFiles')}</span><ArrowRightOutlined /></Link>}
            </div>
          </section>
          <section className={styles.application}>
            <span className={styles.eyebrow}>{t('dashboard.journeyEyebrow')}</span>
            <h2>{t('dashboard.journeyTitle')}</h2>
            {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : failed.includes('applications') ? (
              <p>{t('dashboard.journeyFail')}</p>
            ) : applications.length ? applications.slice(0, 2).map((item) => (
              <div key={item.id} className={styles.applicationRow}>
                <span>{item.applicant_type === 2 ? t('dashboard.officerApply') : t('dashboard.memberApply')}</span>
                <Tag>{item.current_stage || applicationStates[item.status] || t('dashboard.appPending')}</Tag>
              </div>
            )) : <p>{t('dashboard.journeyEmpty')}</p>}
            <Link to="/member/application">{applications.length ? t('dashboard.viewApplication') : t('dashboard.startApplication')} <ArrowRightOutlined /></Link>
          </section>
        </motion.aside>
      </motion.div>
      {canReadStats && overview && (
        <motion.section className={styles.overview} variants={fadeUp}>
          <div><span className={styles.eyebrow}>{t('dashboard.assocEyebrow')}</span><h2>{t('dashboard.assocTitle')}</h2></div>
          <div><strong>{overview.total_members}</strong><span>{t('dashboard.members')}</span></div>
          <div><strong>{overview.total_tasks_in_progress}</strong><span>{t('dashboard.inProgress')}</span></div>
          <div><strong>{overview.total_meetings_this_month}</strong><span>{t('dashboard.meetings')}</span></div>
          <Link to="/stats/overview">{t('dashboard.viewStats')} <ArrowRightOutlined /></Link>
        </motion.section>
      )}
    </motion.div>
  );
}
