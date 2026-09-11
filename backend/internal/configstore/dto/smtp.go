package dto

type SMTPSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	SSLMode  string `json:"ssl_mode"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
	Username string `json:"username"`
}

type SMTPSettingsResponse struct {
	SMTPSettings
	PasswordConfigured bool   `json:"password_configured"`
	PasswordSource     string `json:"password_source,omitempty"`
}

type UpdateSMTPRequest struct {
	Host     string `json:"host" binding:"required,max=255"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	SSLMode  string `json:"ssl_mode" binding:"required"`
	From     string `json:"from" binding:"required,email,max=200"`
	FromName string `json:"from_name" binding:"required,max=100"`
	Username string `json:"username" binding:"omitempty,max=200"`
}

type TestSMTPRequest struct {
	To string `json:"to" binding:"required,email"`
}

type TestSMTPResponse struct {
	Sent bool   `json:"sent"`
	To   string `json:"to"`
}
