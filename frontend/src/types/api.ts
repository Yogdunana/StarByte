// ============================================================
// 通用响应/请求类型
// ============================================================

/** 统一 API 响应 */
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
  request_id: string;
  timestamp: number;
}

/** 分页响应 */
export interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

/** 列表查询参数（支持模块扩展） */
export interface ListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  [key: string]: unknown;
}

/** 通用选项（下拉选择） */
export interface Option {
  label: string;
  value: string | number;
  color?: string;
}

/** 操作结果 */
export interface OperationResult<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
}

// ============================================================
// 用户 & 认证
// ============================================================

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  real_name?: string;
  email?: string;
  phone?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  access_token_expires: number;
  refresh_token_expires: number;
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  username: string;
  real_name: string;
  avatar_url: string;
  email: string;
  phone: string;
  gender: number;
  status: number;
  department_id: string;
  position_id: string;
  roles: RoleInfo[];
  permissions: string[];
  created_at: string;
}

export interface UserListItem {
  id: string;
  username: string;
  real_name: string;
  avatar_url: string;
  email: string;
  phone: string;
  gender: number;
  status: number;
  department_id: string;
  department_name: string;
  position_id: string;
  position_name: string;
  last_login_at: string;
  created_at: string;
}

export interface CreateUserParams {
  username: string;
  password: string;
  real_name?: string;
  email?: string;
  phone?: string;
  gender?: number;
  department_id?: string;
  position_id?: string;
  role_ids?: string[];
}

export interface UpdateUserParams {
  real_name?: string;
  email?: string;
  phone?: string;
  gender?: number;
  status?: number;
  department_id?: string;
  position_id?: string;
  role_ids?: string[];
}

export interface ListUserParams extends ListParams {
  status?: number;
  department_id?: string;
}

// ============================================================
// 角色 & 权限
// ============================================================

export interface RoleInfo {
  id: string;
  name: string;
  code: string;
}

export interface Role {
  id: string;
  name: string;
  code: string;
  sort: number;
  status: number;
  description?: string;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  name: string;
  code: string;
  type: number; // 1=菜单 2=按钮 3=接口
  parent_id?: string;
  sort: number;
  icon?: string;
  path?: string;
  created_at: string;
}

export interface CreateRoleParams {
  name: string;
  code: string;
  sort?: number;
  status?: number;
  description?: string;
  permission_ids?: string[];
}

export interface UpdateRoleParams {
  name?: string;
  code?: string;
  sort?: number;
  status?: number;
  description?: string;
  permission_ids?: string[];
}

export interface ListRoleParams extends ListParams {
  status?: number;
}

// ============================================================
// 部门
// ============================================================

export interface Department {
  id: string;
  name: string;
  code: string;
  parent_id?: string;
  sort: number;
  status: number;
  leader_id?: string;
  description?: string;
  children?: Department[];
  created_at: string;
}

export interface CreateDepartmentParams {
  name: string;
  code: string;
  parent_id?: string;
  sort?: number;
  status?: number;
  leader_id?: string;
  description?: string;
}

export interface UpdateDepartmentParams {
  name?: string;
  code?: string;
  parent_id?: string;
  sort?: number;
  status?: number;
  leader_id?: string;
  description?: string;
}

// ============================================================
// 会员
// ============================================================

export type MemberApplicationStatus = 0 | 1 | 2 | 3 | 4 | 5;
// 0=待审核 1=一面中 2=二面中 3=已通过 4=已拒绝 5=已取消

export interface MemberApplication {
  id: string;
  user_id: string;
  username: string;
  real_name: string;
  type: number; // 1=会员 2=干事
  department_id?: string;
  department_name?: string;
  reason: string;
  status: MemberApplicationStatus;
  current_stage?: string;
  submitted_at: string;
  reviewed_at?: string;
  reviewer_id?: string;
  reviewer_name?: string;
}

export interface MemberProfile {
  id: string;
  user_id: string;
  username: string;
  real_name: string;
  avatar_url: string;
  member_type: number; // 1=会员 2=干事
  department_id?: string;
  department_name?: string;
  position_id?: string;
  position_name?: string;
  join_date: string;
  status: number;
  points: number;
  created_at: string;
}

export interface CreateMemberApplicationParams {
  type: number;
  department_id?: string;
  reason: string;
}

export interface ListMemberApplicationParams extends ListParams {
  status?: MemberApplicationStatus;
  type?: number;
  department_id?: string;
}

export interface ListMemberParams extends ListParams {
  status?: number;
  member_type?: number;
  department_id?: string;
}

// ============================================================
// 面试
// ============================================================

export type InterviewStatus = 0 | 1 | 2 | 3;
// 0=待安排 1=已安排 2=进行中 3=已完成

export interface Interview {
  id: string;
  application_id: string;
  applicant_name: string;
  round: number; // 第几轮
  type: number; // 1=技术面 2=综合面 3=HR面
  status: InterviewStatus;
  interviewer_ids: string[];
  interviewer_names: string[];
  scheduled_at?: string;
  location?: string;
  duration?: number; // 分钟
  score?: number;
  result?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface InterviewEvaluation {
  id: string;
  interview_id: string;
  interviewer_id: string;
  interviewer_name: string;
  score: number;
  comment: string;
  recommendation: number; // 1=强烈推荐 2=推荐 3=待定 4=不推荐
  created_at: string;
}

export interface CreateInterviewParams {
  application_id: string;
  round: number;
  type: number;
  interviewer_ids: string[];
  scheduled_at?: string;
  location?: string;
  duration?: number;
}

export interface UpdateInterviewParams {
  status?: InterviewStatus;
  interviewer_ids?: string[];
  scheduled_at?: string;
  location?: string;
  duration?: number;
  result?: string;
  notes?: string;
}

export interface ListInterviewParams extends ListParams {
  status?: InterviewStatus;
  round?: number;
  type?: number;
  application_id?: string;
}

// ============================================================
// 会议 & 投票
// ============================================================

export type MeetingStatus = 0 | 1 | 2 | 3;
// 0=草稿 1=已发布 2=进行中 3=已结束

export interface Meeting {
  id: string;
  title: string;
  description?: string;
  status: MeetingStatus;
  meeting_type: number; // 1=例会 2=临时会议 3=线上
  start_time: string;
  end_time: string;
  location?: string;
  organizer_id: string;
  organizer_name: string;
  participant_ids: string[];
  participant_count: number;
  created_at: string;
  updated_at: string;
}

export interface MeetingAgenda {
  id: string;
  meeting_id: string;
  title: string;
  description?: string;
  sort: number;
  duration?: number;
  speaker_id?: string;
  speaker_name?: string;
}

export type VoteStatus = 0 | 1 | 2;
// 0=未开始 1=进行中 2=已结束

export interface MeetingVote {
  id: string;
  meeting_id: string;
  title: string;
  description?: string;
  vote_type: number; // 1=等权投票 2=加权投票
  status: VoteStatus;
  is_anonymous: boolean;
  options: VoteOption[];
  total_votes: number;
  start_time?: string;
  end_time?: string;
  created_at: string;
}

export interface VoteOption {
  id: string;
  text: string;
  count: number;
  weight?: number;
}

export interface VoteRecord {
  id: string;
  vote_id: string;
  voter_id: string;
  voter_name: string;
  option_id: string;
  voted_at: string;
}

export interface CreateMeetingParams {
  title: string;
  description?: string;
  meeting_type: number;
  start_time: string;
  end_time: string;
  location?: string;
  participant_ids: string[];
  agendas?: Array<{ title: string; description?: string; sort: number }>;
}

export interface UpdateMeetingParams {
  title?: string;
  description?: string;
  status?: MeetingStatus;
  meeting_type?: number;
  start_time?: string;
  end_time?: string;
  location?: string;
  participant_ids?: string[];
}

export interface ListMeetingParams extends ListParams {
  status?: MeetingStatus;
  meeting_type?: number;
}

export interface CreateVoteParams {
  meeting_id: string;
  title: string;
  description?: string;
  vote_type: number;
  is_anonymous: boolean;
  options: string[];
  start_time?: string;
  end_time?: string;
}

// ============================================================
// 任务
// ============================================================

export type TaskStatus = 0 | 1 | 2 | 3 | 4;
// 0=待处理 1=进行中 2=待审核 3=已完成 4=已取消

export type TaskPriority = 0 | 1 | 2 | 3;
// 0=低 1=中 2=高 3=紧急

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  creator_id: string;
  creator_name: string;
  assignee_id?: string;
  assignee_name?: string;
  department_id?: string;
  department_name?: string;
  due_date?: string;
  progress: number; // 0-100
  tags?: string[];
  related_type?: string;
  related_id?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface TaskComment {
  id: string;
  task_id: string;
  author_id: string;
  author_name: string;
  content: string;
  created_at: string;
}

export interface CreateTaskParams {
  title: string;
  description?: string;
  priority?: TaskPriority;
  assignee_id?: string;
  department_id?: string;
  due_date?: string;
  tags?: string[];
  related_type?: string;
  related_id?: string;
}

export interface UpdateTaskParams {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string;
  department_id?: string;
  due_date?: string;
  progress?: number;
  tags?: string[];
}

export interface ListTaskParams extends ListParams {
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string;
  creator_id?: string;
  department_id?: string;
}

// ============================================================
// 实习
// ============================================================

export interface InternshipRecord {
  id: string;
  user_id: string;
  username: string;
  real_name: string;
  department_id?: string;
  department_name?: string;
  date: string;
  start_time: string;
  end_time: string;
  duration: number; // 小时
  task_description: string;
  status: number; // 0=待审核 1=已通过 2=已拒绝
  reviewer_id?: string;
  reviewer_name?: string;
  reviewed_at?: string;
  created_at: string;
}

export interface InternshipStats {
  user_id: string;
  username: string;
  real_name: string;
  total_hours: number;
  total_days: number;
  rank: number;
}

export interface CreateInternshipParams {
  date: string;
  start_time: string;
  end_time: string;
  task_description: string;
  department_id?: string;
}

export interface UpdateInternshipParams {
  status?: number;
  review_comment?: string;
}

export interface ListInternshipParams extends ListParams {
  user_id?: string;
  status?: number;
  department_id?: string;
  start_date?: string;
  end_date?: string;
}

// ============================================================
// 通知
// ============================================================

export type NotificationType = 1 | 2 | 3 | 4;
// 1=系统通知 2=任务通知 3=会议通知 4=审核通知

export interface Notification {
  id: string;
  title: string;
  content: string;
  type: NotificationType;
  is_read: boolean;
  sender_id?: string;
  sender_name?: string;
  related_type?: string;
  related_id?: string;
  created_at: string;
}

export interface NotificationTemplate {
  id: string;
  name: string;
  code: string;
  title_template: string;
  content_template: string;
  type: number;
  variables: string[];
  status: number;
  created_at: string;
  updated_at: string;
}

export interface ListNotificationParams extends ListParams {
  type?: NotificationType;
  is_read?: boolean;
}

export interface CreateNotificationTemplateParams {
  name: string;
  code: string;
  title_template: string;
  content_template: string;
  type: number;
  variables?: string[];
}

export interface UpdateNotificationTemplateParams {
  name?: string;
  title_template?: string;
  content_template?: string;
  status?: number;
  variables?: string[];
}

// ============================================================
// 流程引擎
// ============================================================

export type FlowDefinitionStatus = 0 | 1 | 2;
// 0=草稿 1=已发布 2=已停用

export type FlowInstanceStatus = 0 | 1 | 2 | 3;
// 0=进行中 1=已完成 2=已驳回 3=已取消

export interface FlowDefinition {
  id: string;
  name: string;
  code: string;
  description?: string;
  status: FlowDefinitionStatus;
  version: number;
  category?: string;
  icon?: string;
  form_schema?: string;
  process_definition?: string;
  created_by: string;
  created_by_name: string;
  created_at: string;
  updated_at: string;
}

export interface FlowInstance {
  id: string;
  definition_id: string;
  definition_name: string;
  business_key?: string;
  status: FlowInstanceStatus;
  initiator_id: string;
  initiator_name: string;
  current_task_id?: string;
  current_task_name?: string;
  started_at: string;
  ended_at?: string;
  form_data?: Record<string, unknown>;
}

export interface FlowTask {
  id: string;
  instance_id: string;
  definition_id: string;
  node_id: string;
  node_name: string;
  task_type: number; // 1=用户任务 2=审批任务
  status: number; // 0=待处理 1=已处理 2=已跳过
  assignee_id?: string;
  assignee_name?: string;
  assignee_type?: number; // 1=用户 2=角色 3=部门
  created_at: string;
  completed_at?: string;
}

export interface FlowTaskHistory {
  id: string;
  task_id: string;
  instance_id: string;
  node_id: string;
  node_name: string;
  action: string; // approve/reject/transfer/comment
  operator_id: string;
  operator_name: string;
  comment?: string;
  created_at: string;
}

export interface ListFlowDefinitionParams extends ListParams {
  status?: FlowDefinitionStatus;
  category?: string;
}

export interface ListFlowInstanceParams extends ListParams {
  definition_id?: string;
  status?: FlowInstanceStatus;
  initiator_id?: string;
}

export interface ListFlowTaskParams extends ListParams {
  status?: number;
  assignee_id?: string;
  definition_id?: string;
}

export interface StartFlowParams {
  definition_id: string;
  business_key?: string;
  form_data?: Record<string, unknown>;
}

export interface ApproveTaskParams {
  task_id: string;
  action: 'approve' | 'reject' | 'transfer';
  comment?: string;
  target_user_id?: string; // 转交时
}

// ============================================================
// 文件
// ============================================================

export interface FileInfo {
  id: string;
  name: string;
  original_name: string;
  size: number;
  mime_type: string;
  storage_type: number; // 1=本地 2=MinIO
  bucket?: string;
  path: string;
  url: string;
  uploader_id: string;
  uploader_name: string;
  created_at: string;
}

export interface UploadResult {
  id: string;
  name: string;
  url: string;
  size: number;
}

export interface ListFileParams extends ListParams {
  uploader_id?: string;
  mime_type?: string;
  start_date?: string;
  end_date?: string;
}

// ============================================================
// 统计
// ============================================================

export interface DashboardStats {
  total_users: number;
  total_members: number;
  total_tasks: number;
  pending_tasks: number;
  total_applications: number;
  pending_applications: number;
  total_interviews: number;
  today_interviews: number;
  total_meetings: number;
  week_meetings: number;
}

export interface ChartDataPoint {
  label: string;
  value: number;
}

export interface PieChartData {
  name: string;
  value: number;
  color?: string;
}
