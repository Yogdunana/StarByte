package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeSponsor   int16 = 1
	TypeEvent     int16 = 2
	TypePurchase  int16 = 3
	TypeOther     int16 = 4
	StatusDraft   int16 = 0
	StatusActive  int16 = 1
	StatusExpired int16 = 2
	StatusEnded   int16 = 3
	TplActive     int16 = 0
)

func ValidType(t int16) bool { return t >= TypeSponsor && t <= TypeOther }

type Template struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(200);not null" json:"name"`
	Code      string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    int16     `gorm:"type:smallint;not null;default:0" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Template) TableName() string { return "contract_templates" }

type Contract struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TemplateID   *uuid.UUID `gorm:"type:uuid" json:"template_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Title        string     `gorm:"type:varchar(200);not null" json:"title"`
	Status       int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	ContractType int16      `gorm:"type:smallint;not null;default:4" json:"contract_type"`
	PartyName    string     `gorm:"type:varchar(200);not null;default:''" json:"party_name"`
	Amount       *float64   `gorm:"type:decimal(12,2)" json:"amount"`
	StartAt      *time.Time `gorm:"type:date" json:"start_at"`
	SignedAt     *time.Time `json:"signed_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	FileID       *uuid.UUID `gorm:"type:uuid" json:"file_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Contract) TableName() string { return "contracts" }

type ContractNamed struct {
	Contract
	OwnerName    string `gorm:"column:owner_name"`
	TemplateName string `gorm:"column:template_name"`
	FileName     string `gorm:"column:file_name"`
}

type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
}
