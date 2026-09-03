package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification 通知模型
type Notification struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Title      string         `gorm:"type:varchar(200);not null" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	Category   string         `gorm:"type:varchar(50);index" json:"category"`          // interview, meeting, task, member, system
	Priority   string         `gorm:"type:varchar(20);default:normal" json:"priority"` // low, normal, high, urgent
	IsRead     bool           `gorm:"default:false;index" json:"is_read"`
	ActionURL  string         `gorm:"type:varchar(500)" json:"action_url"`
	SenderID   *uuid.UUID     `gorm:"type:uuid" json:"sender_id"`
	SenderName string         `gorm:"type:varchar(100)" json:"sender_name"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationTemplate 通知模板模型
type NotificationTemplate struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Code            string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Name            string         `gorm:"type:varchar(200);not null" json:"name"`
	TitleTemplate   string         `gorm:"type:varchar(500);not null" json:"title_template"`
	BodyTemplate    string         `gorm:"type:text;not null" json:"body_template"`
	Channels        string         `gorm:"type:varchar(200);not null" json:"channels"` // JSON array: ["in_app","websocket","email"]
	Category        string         `gorm:"type:varchar(50)" json:"category"`
	VariablesSchema string         `gorm:"type:jsonb" json:"variables_schema"`          // JSON: {"var":"string"}
	Status          int            `gorm:"type:smallint;default:0;index" json:"status"` // 0=启用 1=禁用
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// GetChannels 解析 channels JSON 为字符串切片
func (t *NotificationTemplate) GetChannels() []string {
	var channels []string
	if t.Channels == "" {
		return channels
	}
	_ = json.Unmarshal([]byte(t.Channels), &channels)
	return channels
}

// SetChannels 将字符串切片序列化为 channels JSON
func (t *NotificationTemplate) SetChannels(channels []string) {
	data, _ := json.Marshal(channels)
	t.Channels = string(data)
}

// GetVariablesSchema 解析 variables_schema JSON 为 map
func (t *NotificationTemplate) GetVariablesSchema() map[string]string {
	var schema map[string]string
	if t.VariablesSchema == "" {
		return schema
	}
	_ = json.Unmarshal([]byte(t.VariablesSchema), &schema)
	return schema
}

// SetVariablesSchema 将 map 序列化为 variables_schema JSON
func (t *NotificationTemplate) SetVariablesSchema(schema map[string]string) {
	data, _ := json.Marshal(schema)
	t.VariablesSchema = string(data)
}
