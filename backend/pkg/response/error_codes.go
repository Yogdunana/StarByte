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

const (
	// ===== Success =====
	CodeSuccess = 0

	// ===== General errors (1000-1999) =====
	CodeBadRequest    = 1001 // 参数错误 / 校验失败
	CodeUnauthorized  = 1002 // 未授权 / 未登录
	CodeForbidden     = 1003 // 禁止访问 / 权限不足
	CodeNotFound      = 1004 // 通用资源不存在
	CodeConflict      = 1005 // 通用冲突（资源已存在、状态冲突）
	CodeTooManyReq    = 1006 // 请求过于频繁
	CodeInternalError = 5000 // 内部服务器错误

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
	CodeAuditNotFound    = 5001 // 审计日志不存在
	CodeAuditExportErr   = 5002 // 导出格式不支持
	CodeAuditExportLimit = 5003 // 导出数量超限
	CodeAuditArchiveErr  = 5004 // 归档失败

	// ===== Member module (6000-6999) =====
	CodeMemberAppNotFound = 6001 // 申请不存在
	CodeMemberAppInvalid  = 6002 // 状态不允许操作

	// ===== Interview module (7000-7999) =====
	CodeInterviewNotFound = 7001 // 面试场次不存在
	CodeInterviewConflict = 7002 // 面试时间冲突

	// ===== Meeting module (8000-8999) =====
	CodeMeetingNotFound = 8001 // 会议不存在
	CodeVoteClosed      = 8002 // 投票已结束

	// ===== Task module (9000-9999) =====
	CodeTaskNotFound = 9001 // 任务不存在
	CodeTaskNoAccess = 9002 // 无权操作任务

	// ===== Internship module (10000-10999) =====
	CodeInternshipNotFound = 10001 // 实习记录不存在
	CodeInternshipNoAccess = 10002 // 无权操作实习记录

	// ===== Statistics module (11000-11999) =====
	CodeStatsProviderNotFound = 11001 // 数据提供者不存在
	CodeStatsInvalidParam     = 11002 // 统计参数无效

	// ===== Notification module (12000-12999) =====
	CodeNotificationNotFound    = 12001 // 通知不存在
	CodeNotificationTplExists   = 12002 // 通知模板已存在
	CodeNotificationTplNotFound = 12003 // 通知模板不存在
	CodeNotificationRenderFail  = 12004 // 模板渲染失败（变量缺失）
	CodeNotificationWSAuthFail  = 12005 // WebSocket 认证失败
	CodeNotificationEmailFail   = 12006 // 邮件发送失败
	CodeNotificationBadChannel  = 12007 // 不支持的通知渠道
	CodeNotificationNoAccess    = 12008 // 无权操作该通知
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
}
