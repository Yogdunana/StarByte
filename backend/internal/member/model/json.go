package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONStrings 技能等字符串数组，落库 jsonb。
type JSONStrings []string

func (j JSONStrings) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
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
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
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
