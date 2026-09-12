package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Attachment 公告附件快照（文件已由 /files 上传）
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
	return scanJSON(value, l, func() AttachmentList { return AttachmentList{} })
}

type IDList []string

func (l IDList) Value() (driver.Value, error) {
	if l == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *IDList) Scan(value interface{}) error {
	return scanJSON(value, l, func() IDList { return IDList{} })
}

func scanJSON[T any](value interface{}, dest *T, empty func() T) error {
	if dest == nil {
		return fmt.Errorf("nil dest")
	}
	if value == nil {
		*dest = empty()
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
		*dest = empty()
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return err
	}
	if dest == nil {
		*dest = empty()
	}
	return nil
}
