package response

// Error code ranges for each module, as defined in TEAM_DEV_GUIDE.md §3.4.
//
// Ranges:
//
//	0          Success
//	1000-1999  General errors (validation, auth, not found, conflict, etc.)
//	2000-2999  User module
//	3000-3999  RBAC (role / permission / department / position)
//	4000-4999  Workflow engine
//	5000-5999  Audit log
//	6000-6999  Member module
//	7000-7999  Interview module
//	8000-8999  Meeting module
//	9000-9999  Task module
//	10000-10999 Internship module
//	11000-11999 Statistics module
//	12000-12999 Notification module
//	13000-13999 File module (reserved; #18 type/size checks use 1001)
//	14000-14999 Runtime configstore (#47)
//	15000-15999 Data dictionary (#48)
//	16000-16999 Session management (#50)
//	17000-17999 Export / print engine (#71)
//	18000-18999 Cache management (#72)
//	19000-19999 Scheduler (#73)
//	20000-20999 Unified search (#74)
//	21000-21999 API rate limit / circuit breaker (#75)
//	22000-22999 Dynamic form engine (#28; issue listed 16001-16099, taken by #50 sessions)
//
//	Note: issue #71 asked for 9000-9499, but that range is already owned by
//	the task module (9000-9999). Export therefore uses 17000-17999 (after
//	session 16000). Issue #72 asked for 9500-9699, also inside the task
//	range; cache management uses 18000-18999. Issue #73 asked for 9700-9899,
//	also inside the task range; the scheduler uses 19000-19999. Issue #74
//	asked for 9900-9999; unified search uses 20000-20999. #75 uses 21000-21999.

const (
	// ===== Success =====
	CodeSuccess = 0

	// ===== General errors (1000-1999) =====
	CodeBadRequest     = 1001 // 参数错误 / 校验失败
	CodeUnauthorized   = 1002 // 未授权 / 未登录
	CodeForbidden      = 1003 // 禁止访问 / 权限不足
	CodeNotFound       = 1004 // 通用资源不存在
	CodeConflict       = 1005 // 通用冲突（资源已存在、状态冲突）
	CodeTooManyReq     = 1006 // 请求过于频繁
	CodeInternalError  = 1500 // 内部服务器错误（通用段，不得占用审计 5000-5999）
	CodeNotImplemented = 1501 // 功能未实现 / 接口预留

	// ===== User module (2000-2999) =====
	CodeUserNotFound        = 2001 // 用户不存在
	CodeUserExists          = 2002 // 用户名或邮箱已存在
	CodeInvalidCredentials  = 2003 // 用户名或密码错误
	CodeUserDisabled        = 2004 // 用户已禁用
	CodeUserLocked          = 2005 // 用户已被锁定
	CodeTokenInvalid        = 2006 // Token 无效
	CodeTokenExpired        = 2007 // Token 已过期
	CodeTokenBlacklisted    = 2008 // Token 已失效
	CodeRefreshTokenInvalid = 2009 // Refresh Token 无效
	CodeRefreshTokenExpired = 2010 // Refresh Token 已过期
	CodeRefreshTokenReused  = 2011 // Refresh Token 已被使用（旋转检测）
	CodePasswordTooWeak     = 2012 // 密码强度不足
	CodeOldPasswordWrong    = 2013 // 原密码错误
	CodeAccountLocked       = 2014 // 登录失败次数过多，账号已被锁定

	// ===== RBAC module (3000-3999) =====
	// (defined in internal/rbac/errors.go, range 3001-3020)

	// ===== Workflow engine (4000-4999) =====
	CodeWorkflowNotFound     = 4001 // 流程定义不存在
	CodeWorkflowInstanceEnd  = 4002 // 流程已结束
	CodeWorkflowTaskNotFnd   = 4003 // 流程任务不存在
	CodeWorkflowInvalidNode  = 4004 // 无效的节点配置
	CodeWorkflowKeyExists    = 4005 // 流程定义 key 已存在
	CodeWorkflowDefPublished = 4006 // 流程定义已发布，不可修改
	CodeWorkflowVerNotFound  = 4007 // 流程版本不存在
	CodeWorkflowInstNotFound = 4008 // 流程实例不存在
	CodeWorkflowInstStatus   = 4009 // 流程实例状态不允许操作
	CodeWorkflowTaskStatus   = 4010 // 流程任务状态不允许操作
	CodeWorkflowTaskNoAccess = 4011 // 无权操作流程任务
	CodeWorkflowNodeNotFound = 4012 // 流程节点不存在
	CodeWorkflowExprError    = 4013 // 表达式解析错误
	CodeWorkflowNodeType     = 4014 // 节点类型不支持
	CodeWorkflowDefNotPub    = 4015 // 流程定义未发布

	// ===== Audit log (5000-5999) =====
	CodeAuditNotFound        = 5001 // 审计日志不存在
	CodeAuditExportErr       = 5002 // 导出格式不支持
	CodeAuditExportLimit     = 5003 // 导出数量超限
	CodeAuditArchiveErr      = 5004 // 归档失败
	CodeAuditArchiveNotFound = 5005 // 归档记录不存在
	CodeAuditArchiveFetch    = 5006 // 归档对象拉取失败
	CodeAuditReportErr       = 5007 // 合规报告生成失败
	CodeAuditEntityInvalid   = 5008 // 实体类型或 ID 无效

	// ===== Member module (6000-6999) =====
	CodeMemberAppNotFound   = 6001 // 申请不存在
	CodeMemberAppInvalid    = 6002 // 状态不允许操作
	CodeMemberAppDuplicate  = 6003 // 重复申请
	CodeMemberProfileGone   = 6004 // 档案不存在
	CodeMemberProfileDenied = 6005 // 无权操作该档案
	CodeMemberStudentExists = 6006 // 学号已存在
	CodeMemberExportFail    = 6007 // 导出失败
	CodeMemberFieldRequired = 6008 // 必填字段缺失

	// ===== Interview module (7000-7999) =====
	CodeInterviewNotFound     = 7001 // 面试场次不存在
	CodeInterviewRecordGone   = 7002 // 面试记录不存在
	CodeInterviewConflict     = 7003 // 面试时间冲突
	CodeInterviewNoEvaluator  = 7004 // 面试官未分配
	CodeInterviewDupEval      = 7005 // 重复评分
	CodeInterviewInvalidState = 7006 // 状态不允许操作
	CodeInterviewScoreRange   = 7007 // 评分超出范围
	CodeInterviewDimGone      = 7008 // 维度不存在
	CodeInterviewSessionFull  = 7009 // 超过最大候选人数

	// ===== Meeting module (8000-8999) =====
	CodeMeetingNotFound     = 8001 // 会议不存在
	CodeMeetingInvalidState = 8002 // 会议状态不允许该操作
	CodeMeetingNotAttendee  = 8003 // 非参会人，无权签到
	CodeMeetingDupCheckin   = 8004 // 重复签到
	CodeVoteNotFound        = 8005 // 投票不存在
	CodeVoteNotOpen         = 8006 // 投票未开始或已结束
	CodeVoteDuplicate       = 8007 // 重复投票
	CodeVoteNoAccess        = 8008 // 无权投票（非参会人）
	CodeVoteOptionGone      = 8009 // 投票选项不存在
	CodeVoteAnonymousHidden = 8010 // 匿名投票无法查看个人记录

	// ===== Task module (9000-9999) =====
	CodeTaskNotFound     = 9001 // 任务不存在
	CodeTaskInvalidState = 9002 // 任务状态不允许该操作
	CodeTaskNoAccess     = 9003 // 无权操作该任务
	CodeTaskCommentGone  = 9004 // 评论不存在
	CodeTaskAttachGone   = 9005 // 附件不存在
	CodeTaskUploadFail   = 9006 // 文件上传失败
	CodeTaskClosed       = 9007 // 任务已关闭，无法操作
	CodeTaskTargetGone   = 9008 // 转办目标用户不存在

	// ===== Internship module (10000-10999) =====
	CodeInternshipNotFound     = 10001 // 实习记录不存在
	CodeInternshipNoAccess     = 10002 // 无权操作该实习记录
	CodeInternshipInvalidState = 10003 // 实习状态不允许该操作
	CodeInternshipClosed       = 10004 // 实习已结束，无法修改
	CodeInternshipRankHidden   = 10005 // 排行榜不可见（未开启）
	CodeInternshipDupComplete  = 10006 // 重复完成操作

	// ===== Statistics module (11000-11999) =====
	CodeStatsProviderNotFound = 11001 // 数据提供者不存在
	CodeStatsInvalidParam     = 11002 // 统计参数无效
	CodeStatsExportFormat     = 11003 // 导出格式不支持
	CodeStatsTooLarge         = 11004 // 数据量过大

	// ===== Notification module (12000-12999) =====
	CodeNotificationNotFound     = 12001 // 通知不存在
	CodeNotificationTplExists    = 12002 // 通知模板已存在
	CodeNotificationTplNotFound  = 12003 // 通知模板不存在
	CodeNotificationRenderFail   = 12004 // 模板渲染失败（变量缺失）
	CodeNotificationWSAuthFail   = 12005 // WebSocket 认证失败
	CodeNotificationEmailFail    = 12006 // 邮件发送失败
	CodeNotificationBadChannel   = 12007 // 不支持的通知渠道
	CodeNotificationNoAccess     = 12008 // 无权操作该通知
	CodeNotificationEmailInvalid = 12009 // 邮件参数无效
	CodeNotificationEmailRate    = 12010 // 邮件批量超出限流
	CodeNotificationEmailAttach  = 12011 // 邮件附件不存在

	// ===== Runtime configstore (#47, 14000-14999) =====
	CodeConfigNotFound     = 14001 // 配置不存在
	CodeConfigKeyExists    = 14002 // 配置键已存在
	CodeConfigInvalidType  = 14003 // 不支持的类型或分组
	CodeConfigInvalidValue = 14004 // 配置值与类型不匹配
	CodeConfigProtected    = 14005 // 业务占用配置不可删除
	CodeConfigInvalidKey   = 14006 // 配置键格式不合法

	// ===== Data dictionary (15000-15999) =====
	CodeDictTypeNotFound = 15001 // 字典类型不存在
	CodeDictTypeExists   = 15002 // 字典类型编码已存在
	CodeDictItemNotFound = 15003 // 字典项不存在
	CodeDictItemExists   = 15004 // 字典项值已存在
	CodeDictSystemLocked = 15005 // 系统字典类型不可删除

	// ===== Session management (#50, 16000-16999) =====
	CodeSessionNotFound    = 16001 // 会话不存在或已失效
	CodeSessionUserOffline = 16002 // 该用户当前没有在线会话

	// ===== Export / print engine (#71, 17000-17999) =====
	// Issue #71 listed 9000-9499; that range is the task module. Use 17000+.
	CodeExportInvalidFormat = 17001 // 不支持的导出格式
	CodeExportTplNotFound   = 17002 // 导出模板不存在
	CodeExportTaskNotFound  = 17003 // 导出任务不存在
	CodeExportFileExpired   = 17004 // 导出文件已过期
	CodeExportTooManyRows   = 17005 // 数据量过大且无法异步导出
	CodeExportEmptyData     = 17006 // 导出数据为空

	// ===== Cache management (#72, 18000-18999) =====
	// Issue #72 listed 9500-9699; that range is the task module. Use 18000+.
	CodeCacheKeyNotFound    = 18001 // 缓存键不存在
	CodeCacheInvalidPattern = 18002 // 清除模式无效或过宽
	CodeCacheRedisDown      = 18003 // Redis 不可用
	CodeCacheWarmupFail     = 18004 // 缓存预热失败
	CodeCacheInvalidKey     = 18005 // 缓存键不合法

	// ===== Scheduler (#73, 19000-19999) =====
	// Issue #73 listed 9700-9899; that range is the task module. Use 19000+.
	CodeSchedulerNotFound    = 19001 // 定时任务不存在
	CodeSchedulerInvalidCron = 19002 // Cron 表达式无效
	CodeSchedulerCodeExists  = 19003 // 任务编码已存在
	CodeSchedulerBadHandler  = 19004 // 未知处理器
	CodeSchedulerPaused      = 19005 // 任务状态不允许该操作
	CodeSchedulerBusy        = 19006 // 调度引擎忙或未启动

	// ===== Unified search (#74, 20000-20999) =====
	// Issue #74 listed 9900-9999; that range is the task module. Use 20000+.
	CodeSearchUnknownResource = 20001 // 未知检索资源
	CodeSearchUnknownField    = 20002 // 未知字段
	CodeSearchInvalidOp       = 20003 // 不支持的运算符
	CodeSearchInvalidQuery    = 20004 // 查询不合法
	CodeSearchInvalidCursor   = 20005 // 游标无效
	CodeSearchInvalidAgg      = 20006 // 聚合不合法
	CodeSearchDeepPage        = 20007 // 深分页请改用游标

	// ===== Traffic protection (#75, 21000-21999) =====
	CodeRateLimited = 21001 // 令牌桶限流触发
	CodeCircuitOpen = 21002 // 熔断器打开
	CodeBlacklisted = 21003 // IP/用户在黑名单
	CodeDegraded    = 21004 // 已降级返回

	// ===== Dynamic form engine (#28, 22000-22999) =====
	// Issue #28 listed 16001-16099; that range is session management. Use 22000+.
	CodeFormNotFound      = 22001 // 表单不存在
	CodeFormNotPublished  = 22002 // 表单未发布
	CodeFormFieldRequired = 22003 // 必填字段缺失
	CodeFormFieldInvalid  = 22004 // 字段校验失败
	CodeFormInvalidSchema = 22005 // 表单字段定义不合法
	CodeFormInvalidStatus = 22006 // 表单状态不允许该操作
	CodeFormNameExists    = 22007 // 表单名称已存在
)

// ModuleRanges maps each module name to its error-code range [min, max].
// Used for documentation and validation purposes.
var ModuleRanges = map[string][2]int{
	"general":      {1000, 1999},
	"user":         {2000, 2999},
	"rbac":         {3000, 3999},
	"workflow":     {4000, 4999},
	"audit":        {5000, 5999},
	"member":       {6000, 6999},
	"interview":    {7000, 7999},
	"meeting":      {8000, 8999},
	"task":         {9000, 9999},
	"internship":   {10000, 10999},
	"statistics":   {11000, 11999},
	"notification": {12000, 12999},
	"file":         {13000, 13999},
	"configstore":  {14000, 14999},
	"dict":         {15000, 15999},
	"session":      {16000, 16999},
	"export":       {17000, 17999},
	"cache":        {18000, 18999},
	"scheduler":    {19000, 19999},
	"search":       {20000, 20999},
	"traffic":      {21000, 21999},
	"form":         {22000, 22999},
}
