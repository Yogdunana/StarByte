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
  expires_in: number;
  refresh_expires_in: number;
  user: UserInfo;
}

export interface RefreshResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
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
  department_id?: string;
  position_id?: string;
  roles: string[];
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
// 0=待审核 1=审核中 2=面试中 3=通过 4=拒绝 5=补充材料

export type MemberApplicantType = 1 | 2;

export interface MemberReviewer {
  id: string;
  name: string;
}

export interface MemberApplication {
  id: string;
  user_id: string;
  username?: string;
  applicant_type: MemberApplicantType;
  real_name: string;
  student_no: string;
  department_id?: string;
  department_name?: string;
  reason: string;
  skills: string[];
  experience: string;
  contact_phone: string;
  contact_email: string;
  status: MemberApplicationStatus;
  current_stage?: string;
  flow_instance_id?: string;
  reviewer?: MemberReviewer;
  review_comment?: string;
  required_fields?: string[];
  reviewed_at?: string;
  submitted_at: string;
  created_at: string;
  updated_at: string;
}

export interface MemberProjectItem {
  name: string;
  role: string;
  period: string;
}

export interface MemberNamedRef {
  id: string;
  name: string;
}

export interface MemberProfile {
  id: string;
  user_id: string;
  username?: string;
  real_name: string;
  student_no: string;
  gender: 0 | 1 | 2;
  grade: string;
  major: string;
  department?: MemberNamedRef;
  position?: MemberNamedRef;
  member_type: 1 | 2 | 3 | 4;
  status: 0 | 1 | 2;
  join_date?: string;
  leave_date?: string;
  skills: string[];
  projects: MemberProjectItem[];
  bio: string;
  contact_phone: string;
  contact_email: string;
  created_at: string;
  updated_at: string;
}

export interface CreateMemberApplicationParams {
  applicant_type: MemberApplicantType;
  real_name: string;
  student_no: string;
  department_id?: string;
  reason: string;
  skills: string[];
  experience?: string;
  contact_phone: string;
  contact_email: string;
}

export interface ResubmitMemberApplicationParams {
  real_name?: string;
  student_no?: string;
  department_id?: string;
  reason?: string;
  skills?: string[];
  experience?: string;
  contact_phone?: string;
  contact_email?: string;
}

export interface ListMemberApplicationParams extends ListParams {
  status?: MemberApplicationStatus;
  applicant_type?: MemberApplicantType;
  department_id?: string;
}

export interface ListMemberParams extends ListParams {
  status?: number;
  member_type?: number;
  department_id?: string;
  ids?: string;
}

export interface MemberApplicationHistory {
  id: string;
  from_status: MemberApplicationStatus;
  to_status: MemberApplicationStatus;
  operator_id?: string;
  comment: string;
  created_at: string;
}

export interface MemberProfileHistory {
  id: string;
  field_name: string;
  old_value: string;
  new_value: string;
  operator_id?: string;
  reason: string;
  created_at: string;
}

export interface MemberStatItem {
  key: string;
  label: string;
  count: number;
}

export interface MemberStatsResponse {
  group_by: string;
  items: MemberStatItem[];
}

export interface MemberDepartmentOption {
  id: string;
  name: string;
}

// ============================================================
// 面试
// ============================================================

export type InterviewSessionStatus = 0 | 1 | 2 | 3;
// 0待开始 1进行中 2已结束 3已取消

export type InterviewRecordStatus = 0 | 1 | 2 | 3 | 4 | 5;
// 0待面试 1已签到 2面试中 3已完成 4缺席 5已取消

export type InterviewResult = 0 | 1 | 2 | 3;
// 0未出 1通过 2不通过 3待定

export interface InterviewPerson {
  id: string;
  name: string;
}

export interface InterviewSession {
  id: string;
  title: string;
  round: number;
  department_id?: string;
  department_name?: string;
  start_time: string;
  end_time: string;
  location: string;
  online_link?: string;
  status: InterviewSessionStatus;
  max_candidates: number;
  description: string;
  candidate_count: number;
  created_at: string;
  updated_at: string;
}

export interface Interview {
  id: string;
  session_id?: string;
  session_title?: string;
  applicant: InterviewPerson;
  student_no?: string;
  application_id?: string;
  status: InterviewRecordStatus;
  scheduled_time?: string;
  actual_start_time?: string;
  actual_end_time?: string;
  result: InterviewResult;
  result_comment?: string;
  location?: string;
  duration?: number;
  score?: number;
  evaluators: InterviewPerson[];
  department_name?: string;
  created_at: string;
  updated_at: string;
}

export interface InterviewQRCode {
  session_id: string;
  token: string;
  checkin_path: string;
  png_base64: string;
}

export interface EvaluationDimension {
  id: string;
  name: string;
  weight: number;
  max_score: number;
  sort_order: number;
}

export interface DimensionScore {
  dimension: string;
  score: number;
  comment?: string;
}

export interface EvaluatorScores {
  evaluator: InterviewPerson;
  scores: DimensionScore[];
  total_score: number;
}

export interface EvaluationSummary {
  interview_id: string;
  applicant: InterviewPerson;
  evaluations: EvaluatorScores[];
  average_score: number;
  weighted_score: number;
}

export interface InterviewStats {
  total: number;
  pass_count: number;
  fail_count: number;
  pending_count: number;
  pass_rate: number;
  score_buckets: Array<{ range: string; count: number }>;
  by_department: Array<{ department: string; count: number; pass_count: number }>;
}

export interface CreateSessionParams {
  title: string;
  round: number;
  department_id?: string;
  start_time: string;
  end_time: string;
  location?: string;
  online_link?: string;
  max_candidates: number;
  description?: string;
}

export interface CreateInterviewParams {
  session_id: string;
  applicant_id?: string;
  application_id?: string;
  scheduled_time?: string;
  location?: string;
  duration?: number;
}

export interface ListSessionParams extends ListParams {
  status?: InterviewSessionStatus;
  department_id?: string;
  round?: number;
}

export interface ListInterviewParams extends ListParams {
  session_id?: string;
  status?: InterviewRecordStatus;
  result?: InterviewResult;
}

// ============================================================
// 会议 & 投票
// ============================================================

export type MeetingStatus = 0 | 1 | 2 | 3;
// 0=待开始 1=进行中 2=已结束 3=已取消

export interface MeetingOrganizer {
  id: string;
  name: string;
}

export interface Meeting {
  id: string;
  title: string;
  description?: string;
  status: MeetingStatus;
  meeting_type: number; // 1=例会 2=临时 3=线上
  start_time: string;
  end_time: string;
  location?: string;
  online_link?: string;
  organizer: MeetingOrganizer;
  minutes?: string;
  cancel_reason?: string;
  attendee_count: number;
  checked_in_count: number;
  created_at: string;
  updated_at: string;
}

export interface MeetingAgenda {
  id: string;
  meeting_id: string;
  title: string;
  content: string;
  duration: number;
  sort_order: number;
  presenter: string;
}

export interface MeetingAttendee {
  id: string;
  user_id: string;
  name: string;
  position_code?: string;
  attended: boolean;
  checked_in_at?: string;
}

export type VoteStatus = 0 | 1 | 2 | 3;

export interface MeetingVote {
  id: string;
  meeting_id: string;
  title: string;
  description?: string;
  vote_type: 1 | 2;
  is_anonymous: boolean;
  options: Array<{ key: string; label: string }>;
  status: VoteStatus;
  start_time?: string;
  end_time?: string;
  has_voted: boolean;
  created_at: string;
}

export interface VoteResultItem {
  option_key: string;
  option_label: string;
  count: number;
  weight_total: number;
}

export interface VoteResult {
  id: string;
  title: string;
  vote_type: 1 | 2;
  is_anonymous: boolean;
  status: VoteStatus;
  results: VoteResultItem[];
  total_voters: number;
  total_weight: number;
  start_time?: string;
  end_time?: string;
}

export interface MeetingQRCode {
  meeting_id: string;
  token: string;
  checkin_path: string;
  png_base64: string;
}

export interface VoteWeightConfig {
  weights: Record<string, number>;
  default_weight: number;
}

export interface CreateMeetingParams {
  title: string;
  description?: string;
  meeting_type: number;
  start_time: string;
  end_time: string;
  location?: string;
  online_link?: string;
  user_ids?: string[];
}

export interface UpdateMeetingParams {
  title?: string;
  description?: string;
  meeting_type?: number;
  start_time?: string;
  end_time?: string;
  location?: string;
  online_link?: string;
}

export interface ListMeetingParams extends ListParams {
  status?: MeetingStatus;
  start_date?: string;
  end_date?: string;
}

export interface CreateVoteParams {
  title: string;
  description?: string;
  vote_type: 1 | 2;
  is_anonymous: boolean;
  options: Array<{ key: string; label: string }>;
  duration: number;
}

// ============================================================
// 任务
// ============================================================

export type TaskStatus = 0 | 1 | 2 | 3 | 4;
// 0=待处理 1=进行中 2=已完成 3=已取消 4=已挂起

export type TaskPriority = 0 | 1 | 2 | 3;
// 0=低 1=中 2=高 3=紧急

export interface TaskPerson {
  id: string;
  name: string;
  avatar?: string;
}

export interface TaskBrief {
  id: string;
  title: string;
  status: TaskStatus;
}

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee?: TaskPerson;
  creator: TaskPerson;
  department?: TaskPerson;
  parent?: { id: string; title: string };
  children: TaskBrief[];
  due_date?: string;
  completed_at?: string;
  tags: string[];
  comment_count: number;
  attachment_count: number;
  progress: number;
  created_at: string;
  updated_at: string;
}

export interface TaskComment {
  id: string;
  task_id: string;
  user_id: string;
  user: TaskPerson;
  content: string;
  mentions: string[];
  created_at: string;
  updated_at: string;
}

export interface TaskLog {
  id: string;
  action_type: 'status_change' | 'assign' | 'transfer' | 'comment' | 'urge' | 'create' | string;
  old_value: string;
  new_value: string;
  operator: TaskPerson;
  comment: string;
  created_at: string;
}

export interface TaskAttachment {
  id: string;
  task_id: string;
  file_id: string;
  file_name: string;
  file_path: string;
  file_size: number;
  file_type: string;
  uploaded_by: string;
  created_at: string;
}

export interface TaskStats {
  total: number;
  overdue: number;
  by_status: Record<string, number>;
  by_priority: Record<string, number>;
}

export interface CreateTaskParams {
  title: string;
  description?: string;
  priority?: TaskPriority;
  assignee_id?: string;
  department_id?: string;
  due_date?: string;
  tags?: string[];
  parent_id?: string;
}

export interface UpdateTaskParams {
  title?: string;
  description?: string;
  priority?: TaskPriority;
  due_date?: string;
  tags?: string[];
}

export interface ListTaskParams extends ListParams {
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string;
  department_id?: string;
  tags?: string;
  sort_by?: string;
  sort_order?: string;
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

/** 通知优先级 */
export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';

/** 通知分类 */
export type NotificationCategory =
  | 'system'
  | 'task'
  | 'meeting'
  | 'approval'
  | 'interview'
  | 'other';

/** 通知发送者信息 */
export interface NotificationSender {
  id: string;
  name: string;
}

/** 通知实体 */
export interface Notification {
  id: string;
  title: string;
  content: string;
  category: NotificationCategory;
  priority: NotificationPriority;
  is_read: boolean;
  action_url: string;
  sender: NotificationSender;
  created_at: string;
}

/** 通知模板实体 */
export interface NotificationTemplate {
  id: string;
  code: string;
  name: string;
  title_template: string;
  body_template: string;
  channels: string[];
  category: string;
  variables_schema: Record<string, string>;
  status: number; // 0=禁用 1=启用
  created_at: string;
  updated_at: string;
}

/** 通知列表查询参数 */
export interface ListNotificationParams extends ListParams {
  category?: NotificationCategory;
  unread_only?: boolean;
}

/** 创建通知模板参数 */
export interface CreateNotificationTemplateParams {
  code: string;
  name: string;
  title_template: string;
  body_template: string;
  channels: string[];
  category?: string;
  variables_schema?: Record<string, string>;
}

/** 更新通知模板参数 */
export interface UpdateNotificationTemplateParams {
  name?: string;
  title_template?: string;
  body_template?: string;
  channels?: string[];
  category?: string;
  variables_schema?: Record<string, string>;
  status?: number;
}

/** 测试模板参数 */
export interface TestTemplateParams {
  variables: Record<string, unknown>;
}

/** 测试模板结果 */
export interface TestTemplateResult {
  title: string;
  content: string;
}

/** 管理员发送通知参数 */
export interface SendNotificationParams {
  user_ids: string[];
  template_code: string;
  variables?: Record<string, unknown>;
  channels?: string[];
}

/** 广播通知参数 */
export interface BroadcastNotificationParams {
  title: string;
  content: string;
  category?: string;
  priority?: NotificationPriority;
  channels?: string[];
}

/** 模板列表查询参数 */
export interface ListTemplateParams extends ListParams {
  keyword?: string;
}

/** WebSocket 实时通知消息 */
export interface WSNotificationMessage {
  type: 'notification' | 'auth_result' | 'pong' | 'connected';
  data: {
    title?: string;
    content?: string;
    category?: string;
    priority?: string;
    action_url?: string;
    sender?: { id: string; name: string };
    success?: boolean;
  };
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

export interface FileUploader {
  id: string;
  name: string;
}

export interface FileInfo {
  id: string;
  filename: string;
  name: string;
  original_name: string;
  file_size: number;
  size: number;
  mime_type: string;
  category: string;
  storage_type?: string;
  bucket?: string;
  path?: string;
  url: string;
  thumbnail_url?: string;
  is_public?: boolean;
  uploader?: FileUploader;
  uploader_id: string;
  uploader_name: string;
  created_at: string;
  uploaded_at?: string;
}

export interface UploadResult {
  id: string;
  filename: string;
  name: string;
  original_name: string;
  file_size: number;
  size: number;
  mime_type: string;
  category: string;
  url: string;
  thumbnail_url?: string;
  uploaded_at: string;
}

export interface ListFileParams extends ListParams {
  category?: string;
  keyword?: string;
  uploader_id?: string;
  mime_type?: string;
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
