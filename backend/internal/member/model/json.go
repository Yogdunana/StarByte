package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JSONStrings 技能等字符串数组，落库 jsonb。
type JSONStrings []string

func jsonbArrayValue(payload interface{}) (driver.Value, error) {
	if payload == nil {
		return "[]", nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || string(b) == "null" {
		return "[]", nil
	}
	return string(b), nil
}

func jsonbArrayExpr(payload interface{}) clause.Expr {
	v, err := jsonbArrayValue(payload)
	if err != nil {
		return clause.Expr{SQL: "'[]'::jsonb"}
	}
	return clause.Expr{SQL: "?::jsonb", Vars: []interface{}{v}}
}

func (j JSONStrings) Value() (driver.Value, error) {
	return jsonbArrayValue(j)
}

// GormValue forces GORM to persist [] instead of NULL for a zero slice.
func (j JSONStrings) GormValue(context.Context, *gorm.DB) clause.Expr {
	return jsonbArrayExpr(j)
}

func (j *JSONStrings) Scan(value interface{}) error {
	bytes, err := asBytes(value)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		*j = JSONStrings{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ProjectItem 档案项目经历。
type ProjectItem struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Period string `json:"period"`
}

// JSONProjects 项目经历数组。
type JSONProjects []ProjectItem

func (j JSONProjects) Value() (driver.Value, error) {
	return jsonbArrayValue(j)
}

// GormValue forces GORM to persist [] instead of NULL for a zero slice.
func (j JSONProjects) GormValue(context.Context, *gorm.DB) clause.Expr {
	return jsonbArrayExpr(j)
}

func (j *JSONProjects) Scan(value interface{}) error {
	bytes, err := asBytes(value)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		*j = JSONProjects{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func asBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported jsonb type %T", value)
	}
}
