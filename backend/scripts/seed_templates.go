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
		Code: "meeting_notice", Name: "会议通知", Category: "meeting",
		Title:  "会议通知：{{.title}}",
		Body:   "会议「{{.title}}」将于 {{.start_time}} 在 {{.location}} 召开。",
		Schema: `{"title":"string","start_time":"string","location":"string"}`,
	},
	{
		Code: "discipline_notice", Name: "处分通知", Category: "system",
		Title:  "纪律处分通知",
		Body:   "{{.real_name}}，你收到纪律处分：{{.title}}。",
		Schema: `{"real_name":"string","title":"string"}`,
	},
	{
		Code: "task_assigned", Name: "任务分配", Category: "task",
		Title:  "新任务：{{.title}}",
		Body:   "{{.real_name}}，你被分配了任务「{{.title}}」。",
		Schema: `{"real_name":"string","title":"string"}`,
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
