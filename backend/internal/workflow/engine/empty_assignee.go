package engine

import (
	"errors"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// DefaultFallbackRoleCode is used when an approval node has nobody that can be
// assigned and the node policy allows escalation.
const DefaultFallbackRoleCode = "president"

// EmptyPolicy decides what an approval node does when its assignee strategy
// resolves to nobody.
//
// The default is derived from the strategy: only role-based nodes can be
// genuinely unstaffed ("the association has nobody in this office right now"),
// so they escalate first. Every other strategy keeps the historical behaviour
// of failing loudly, because an empty static list or a missing business
// resolver is a configuration error rather than a vacancy.
type EmptyPolicy string

const (
	// EmptySkip passes through the node without creating a task. A history row
	// records the skip so the step never disappears without a trace.
	EmptySkip EmptyPolicy = "skip"
	// EmptyEscalate retries with FallbackRoleCode before falling back to EmptySkip.
	EmptyEscalate EmptyPolicy = "escalate"
	// EmptyFail rethrows the error and blocks the flow.
	EmptyFail EmptyPolicy = "fail"
)

// EmptyAssigneeError marks "this node structurally has nobody who could act".
//
// The empty-assignee policy must not be decided by matching Chinese message
// text: two producers ("no role holder", "no department scope") expressed the
// same situation and only one of them was recognised, so the other tore down
// the whole approval instead of following the node policy.
type EmptyAssigneeError struct {
	Err    *response.AppError
	NodeID string
	Role   string
	Reason string
}

// NewEmptyAssignee wraps message in a recognisable error while keeping the
// existing AppError contract for HTTP handlers.
func NewEmptyAssignee(nodeID, role, reason, message string) *EmptyAssigneeError {
	return &EmptyAssigneeError{
		Err:    response.NewAppError(response.CodeWorkflowInvalidNode, message),
		NodeID: nodeID,
		Role:   role,
		Reason: reason,
	}
}

func (e *EmptyAssigneeError) Error() string { return e.Err.Error() }
func (e *EmptyAssigneeError) Unwrap() error { return e.Err }

// IsEmptyAssignee reports whether err is an empty-assignee situation.
func IsEmptyAssignee(err error) bool {
	var target *EmptyAssigneeError
	return errors.As(err, &target)
}

// EmptyAssigneeOf unwraps the detail, or nil when err has another cause.
func EmptyAssigneeOf(err error) *EmptyAssigneeError {
	var target *EmptyAssigneeError
	if errors.As(err, &target) {
		return target
	}
	return nil
}

// EmptyPolicyOf reads the onEmpty decision for a node.
//
// Explicit `onEmpty` always wins. Otherwise nodes that already opted into
// `skipIfEmpty` keep skipping, and the remaining role-based nodes escalate to
// their fallback role instead of blocking the whole business transaction.
func EmptyPolicyOf(node *FlowNode) EmptyPolicy {
	if node != nil && node.Config != nil {
		if _, ok := node.Config["onEmpty"]; ok {
			if raw, ok := node.Config["onEmpty"].(string); ok {
				switch strings.ToLower(strings.TrimSpace(raw)) {
				case string(EmptySkip):
					return EmptySkip
				case string(EmptyEscalate):
					return EmptyEscalate
				case string(EmptyFail):
					return EmptyFail
				}
			}
		}
		if v, ok := node.Config["skipIfEmpty"]; ok {
			switch typed := v.(type) {
			case bool:
				if typed {
					return EmptySkip
				}
				return EmptyFail
			case string:
				if typed == "true" || typed == "1" {
					return EmptySkip
				}
				return EmptyFail
			}
		}
	}
	if strategyOf(node) != "role" && strategyOf(node) != "" {
		return EmptyFail
	}
	if skipIfEmptyNode(node) {
		return EmptySkip
	}
	return EmptyEscalate
}

func strategyOf(node *FlowNode) string {
	if node == nil || node.Config == nil {
		return ""
	}
	raw, _ := node.Config["assigneeStrategy"].(string)
	return strings.TrimSpace(raw)
}

// ExplicitFallbackRoleCode reports a fallback configured on the node itself,
// as opposed to the engine-wide DefaultFallbackRoleCode.
func ExplicitFallbackRoleCode(node *FlowNode) (string, bool) {
	if node != nil && node.Config != nil {
		if raw, ok := node.Config["fallbackRoleCode"].(string); ok {
			if code := strings.TrimSpace(raw); code != "" {
				return code, true
			}
		}
	}
	return "", false
}

// FallbackRoleCode returns the role that takes over an unstaffed node.
func FallbackRoleCode(node *FlowNode) string {
	if code, ok := ExplicitFallbackRoleCode(node); ok {
		return code
	}
	return DefaultFallbackRoleCode
}
