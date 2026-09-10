package middleware

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/audit"
	"github.com/gin-gonic/gin"
)

const (
	auditCtxEntityType = "audit_entity_type"
	auditCtxEntityID   = "audit_entity_id"
	auditCtxBeforeJSON = "audit_before_json"
)

var uuidLike = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var resourceSingular = map[string]string{
	"users":       "user",
	"roles":       "role",
	"departments": "department",
	"permissions": "permission",
	"positions":   "position",
	"audit-logs":  "audit_log",
	"members":     "member",
	"interviews":  "interview",
	"meetings":    "meeting",
	"tasks":       "task",
	"internships": "internship",
	"activities":  "activity",
}

// SetAuditSnapshot 由业务 handler 在写操作前写入实体 before 快照。
func SetAuditSnapshot(c *gin.Context, entityType, entityID string, before any) {
	if c == nil {
		return
	}
	if entityType != "" {
		c.Set(auditCtxEntityType, entityType)
	}
	if entityID != "" {
		c.Set(auditCtxEntityID, entityID)
	}
	if before == nil {
		return
	}
	c.Set(auditCtxBeforeJSON, audit.CompactJSON(before, maxRequestBodySize))
}

func snapshotFromContext(c *gin.Context) (entityType, entityID, beforeJSON string) {
	if c == nil {
		return "", "", ""
	}
	return c.GetString(auditCtxEntityType), c.GetString(auditCtxEntityID), c.GetString(auditCtxBeforeJSON)
}

func shouldAudit(method, path string) bool {
	if skipAuditPaths[path] {
		return false
	}
	if writeMethods[method] {
		return true
	}
	if strings.EqualFold(method, "GET") && isExportPath(path) {
		return true
	}
	return false
}

func isExportPath(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "/export") || strings.HasSuffix(p, "/reports")
}

// ParseEntity 从 /api/v1/{module?}/{resource}/{id} 解析实体类型与 ID。
func ParseEntity(path string) (entityType, entityID string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" {
		return "", ""
	}
	rest := parts[2:]
	if len(rest) > 0 && rest[0] == "system" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return "", ""
	}
	entityType = singularResource(rest[0])
	if len(rest) >= 2 && looksLikeEntityID(rest[1]) {
		entityID = rest[1]
	}
	return entityType, entityID
}

func singularResource(name string) string {
	if mapped, ok := resourceSingular[name]; ok {
		return mapped
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		return strings.TrimSuffix(name, "s")
	}
	return name
}

func looksLikeEntityID(s string) bool {
	if uuidLike.MatchString(s) {
		return true
	}
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

// ComplianceFlags 按方法/路径/请求体自动打标：delete / permission / export。
func ComplianceFlags(method, path, reqBody string) string {
	var flags []string
	if strings.EqualFold(method, "DELETE") {
		flags = append(flags, "delete")
	}
	if isExportPath(path) {
		flags = append(flags, "export")
	}
	p := strings.ToLower(path)
	body := strings.ToLower(reqBody)
	if strings.Contains(p, "/permissions") || strings.Contains(p, "/roles") ||
		strings.Contains(body, `"role_ids"`) || strings.Contains(body, `"permission_ids"`) {
		flags = append(flags, "permission")
	}
	return strings.Join(flags, ",")
}

func extractDataJSON(respBody string) string {
	if strings.TrimSpace(respBody) == "" {
		return ""
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &envelope); err != nil || len(envelope.Data) == 0 {
		return ""
	}
	if string(envelope.Data) == "null" {
		return ""
	}
	return audit.DesensitizeJSON(string(envelope.Data))
}

func extractJSONID(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return ""
	}
	id, _ := obj["id"].(string)
	if looksLikeEntityID(id) {
		return id
	}
	return ""
}

func fillTraceFields(c *gin.Context, entry *AuditLogEntry, reqBody, respBody string) {
	typ, id := ParseEntity(entry.Path)
	snapType, snapID, before := snapshotFromContext(c)
	if snapType != "" {
		typ = snapType
	}
	if snapID != "" {
		id = snapID
	}
	if id == "" {
		if pid := c.Param("id"); pid != "" && looksLikeEntityID(pid) {
			id = pid
		}
	}
	reqAfter := sanitizeRequestBody(entry.Path, reqBody)
	respAfter := extractDataJSON(respBody)
	after := resolveAfterJSON(before, reqAfter, respAfter)
	if id == "" {
		id = extractJSONID(respAfter)
	}
	if id == "" {
		id = extractJSONID(after)
	}
	if len(after) > maxRequestBodySize {
		after = after[:maxRequestBodySize] + "...[truncated]"
	}
	var diff string
	if before != "" || after != "" {
		diff = audit.DiffJSON(before, after)
	}
	entry.EntityType = typ
	entry.EntityID = id
	entry.BeforeJSON = before
	entry.AfterJSON = after
	entry.DiffJSON = diff
	entry.ComplianceFlags = ComplianceFlags(entry.Method, entry.Path, reqBody)
	if isExportPath(entry.Path) && strings.EqualFold(entry.Method, "GET") {
		entry.Action = "EXPORT"
	}
}

func resolveAfterJSON(before, reqAfter, respAfter string) string {
	patch := reqAfter
	if strings.TrimSpace(respAfter) != "" {
		patch = respAfter
	}
	if strings.TrimSpace(patch) == "" || patch == "[redacted: sensitive endpoint]" {
		return patch
	}
	if strings.TrimSpace(before) != "" {
		return overlayJSON(before, patch)
	}
	return patch
}

// overlayJSON 用 after 的顶层键覆盖 before，避免部分请求体把快照字段标成删除。
func overlayJSON(before, after string) string {
	if strings.TrimSpace(before) == "" {
		return after
	}
	if strings.TrimSpace(after) == "" {
		return before
	}
	var baseMap, overlayMap map[string]any
	if err := json.Unmarshal([]byte(before), &baseMap); err != nil || baseMap == nil {
		return after
	}
	if err := json.Unmarshal([]byte(after), &overlayMap); err != nil || overlayMap == nil {
		return after
	}
	for k, v := range overlayMap {
		baseMap[k] = v
	}
	out, err := json.Marshal(baseMap)
	if err != nil {
		return after
	}
	return string(out)
}
