package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Attachment struct {
	FileID string `json:"file_id"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

type AttachmentList []Attachment

func (l AttachmentList) Value() (driver.Value, error) {
	if l == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *AttachmentList) Scan(value interface{}) error {
	if l == nil {
		return fmt.Errorf("nil dest")
	}
	if value == nil {
		*l = AttachmentList{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported json type %T", value)
	}
	if len(raw) == 0 {
		*l = AttachmentList{}
		return nil
	}
	return json.Unmarshal(raw, l)
}
