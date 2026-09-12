package main

import "gorm.io/gorm"

type seedTemplate struct {
	Code     string
	Name     string
	Title    string
	Body     string
	Category string
	Schema   string
}

var seedTemplatesData = []seedTemplate{
	{
		Code: "member_approved", Name: "入会申请通过", Category: "member",
		Title:  "入会申请已通过",
		Body:   "{{.real_name}}，你的入会申请已通过，欢迎加入。",
		Schema: `{"real_name":"string"}`,
	},
	{
		Code: "interview_invite", Name: "面试邀请", Category: "interview",
		Title:  "面试邀请：{{.round}}",
		Body:   "{{.real_name}}，请于 {{.scheduled_at}} 在 {{.location}} 参加面试。",
		Schema: `{"real_name":"string","round":"string","scheduled_at":"string","location":"string"}`,
	},
	{
		Code: "interview_assigned", Name: "面试官分配", Category: "interview",
		Title:  "你被分配为面试官",
		Body:   "请面试 {{.applicant_name}}，时间 {{.scheduled_at}}，地点 {{.location}}。",
		Schema: `{"applicant_name":"string","scheduled_at":"string","location":"string","round":"string"}`,
	},
	{
		Code: "interview_result", Name: "面试结果", Category: "interview",
		Title:  "面试结果：{{.result}}",
		Body:   "{{.real_name}}，你的面试结果为 {{.result}}。请在申请进度页查看后续安排。",
		Schema: `{"real_name":"string","result":"string"}`,
	},
	{
		Code: "meeting_notice", Name: "会议通知", Category: "meeting",
		Title:  "会议通知：{{.title}}",
		Body:   "会议「{{.title}}」将于 {{.start_time}} 在 {{.location}} 召开。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "meeting_started", Name: "会议开始", Category: "meeting",
		Title:  "会议已开始：{{.title}}",
		Body:   "会议「{{.title}}」已开始，地点 {{.location}}。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "meeting_ended", Name: "会议结束", Category: "meeting",
		Title:  "会议已结束：{{.title}}",
		Body:   "会议「{{.title}}」已结束。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "discipline_notice", Name: "处分通知", Category: "system",
		Title:  "纪律处分通知",
		Body:   "{{.real_name}}，你收到纪律处分：{{.title}}。",
		Schema: `{"real_name":"string","title":"string"}`,
	},
	{
		Code: "discipline_revoked", Name: "处分撤销", Category: "system",
		Title:  "纪律处分已撤销",
		Body:   "{{.real_name}}，你的纪律处分「{{.title}}」已撤销。",
		Schema: `{"real_name":"string","title":"string"}`,
	},
	{
		Code: "task_assigned", Name: "任务分配", Category: "task",
		Title:  "新任务：{{.title}}",
		Body:   "{{.real_name}}，你被分配了任务「{{.title}}」。",
		Schema: `{"real_name":"string","title":"string"}`,
	},
	{
		Code: "task_transferred", Name: "任务转办", Category: "task",
		Title:  "任务转办：{{.title}}",
		Body:   "任务「{{.title}}」已转办给你。{{.message}}",
		Schema: `{"title":"string","message":"string"}`,
	},
	{
		Code: "task_urged", Name: "任务催办", Category: "task",
		Title:  "催办：{{.title}}",
		Body:   "任务「{{.title}}」被催办。{{.message}}",
		Schema: `{"title":"string","message":"string"}`,
	},
	{
		Code: "task_mention", Name: "任务评论提及", Category: "task",
		Title:  "有人在任务中提到你：{{.title}}",
		Body:   "任务「{{.title}}」的评论提到了你：{{.message}}",
		Schema: `{"title":"string","message":"string"}`,
	},
	{
		Code: "task_due_soon", Name: "任务即将到期", Category: "task",
		Title:  "任务即将到期：{{.title}}",
		Body:   "任务「{{.title}}」将于 {{.due_date}} 到期。",
		Schema: `{"title":"string","due_date":"string"}`,
	},
	{
		Code: "task_overdue", Name: "任务已超期", Category: "task",
		Title:  "任务已超期：{{.title}}",
		Body:   "任务「{{.title}}」已超过截止时间 {{.due_date}}。",
		Schema: `{"title":"string","due_date":"string"}`,
	},
	{
		Code: "export_ready", Name: "导出完成", Category: "system",
		Title:  "导出已完成：{{.filename}}",
		Body:   "文件「{{.filename}}」已生成，30 分钟内可下载。",
		Schema: `{"filename":"string","file_id":"string"}`,
	},
	{
		Code: "scheduler_task_failed", Name: "定时任务失败", Category: "system",
		Title:  "定时任务失败：{{.task_name}}",
		Body:   "任务「{{.task_name}}」已进入死信：{{.error}}",
		Schema: `{"task_name":"string","error":"string"}`,
	},
	{
		Code: "contract_expiring", Name: "合同即将到期", Category: "system",
		Title:  "合同即将到期：{{.title}}",
		Body:   "{{.real_name}}，合同「{{.title}}」将于 {{.expired_at}} 到期。",
		Schema: `{"real_name":"string","title":"string","expired_at":"string"}`,
	},
	{
		Code: "activity_registered", Name: "活动报名成功", Category: "activity",
		Title:  "报名成功：{{.title}}",
		Body:   "你已成功报名活动「{{.title}}」，时间 {{.start_time}}，地点 {{.location}}。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "activity_waitlist", Name: "活动候补", Category: "activity",
		Title:  "已进入候补队列：{{.title}}",
		Body:   "活动「{{.title}}」名额已满，你已进入候补队列，如有空位将自动递补。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "activity_approved", Name: "活动报名通过", Category: "activity",
		Title:  "报名已通过：{{.title}}",
		Body:   "你报名的活动「{{.title}}」已通过，时间 {{.start_time}}，地点 {{.location}}。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "activity_rejected", Name: "活动报名被拒", Category: "activity",
		Title:  "报名未通过：{{.title}}",
		Body:   "很遗憾，你报名的活动「{{.title}}」未通过审核。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "activity_starting", Name: "活动即将开始", Category: "activity",
		Title:  "活动即将开始：{{.title}}",
		Body:   "活动「{{.title}}」将于 {{.start_time}} 在 {{.location}} 开始，请准时参加。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "schedule_reminder", Name: "日程提醒", Category: "schedule",
		Title:  "日程提醒：{{.title}}",
		Body:   "「{{.title}}」将于 {{.start_at}} 开始（提前 {{.minutes}} 分钟）。",
		Schema: `{"title":"string","start_at":"string","minutes":"string"}`,
	},
	{
		Code: "announcement_published", Name: "公告发布", Category: "announcement",
		Title:  "新公告：{{.title}}",
		Body:   "「{{.title}}」已发布，请查看公告中心。",
		Schema: `{"title":"string","category":"string","id":"string"}`,
	},
}

func seedTemplates(db *gorm.DB) error {
	channels := `["in_app","websocket","email"]`
	for _, t := range seedTemplatesData {
		if err := db.Exec(`
			INSERT INTO notification_templates
				(id, code, name, title_template, body_template, channels, category, variables_schema, status)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?, ?::jsonb, 0)
			ON CONFLICT (code) DO NOTHING
		`, t.Code, t.Name, t.Title, t.Body, channels, t.Category, t.Schema).Error; err != nil {
			return err
		}
	}
	return nil
}
