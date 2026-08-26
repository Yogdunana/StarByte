package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// identifierChainRegex matches a chain of identifiers connected by dots,
// e.g., "applicant.age" or "user.profile.address.city".
// It does NOT match floating point numbers like "3.14" because the first
// character must be a letter or underscore.
var identifierChainRegex = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)+)`)

// ExpressionEngine evaluates conditional expressions for exclusive gateways.
// It wraps govaluate to provide a consistent evaluation interface and
// translates errors into AppError with workflow error codes.
type ExpressionEngine struct{}

// NewExpressionEngine creates a new ExpressionEngine.
func NewExpressionEngine() *ExpressionEngine {
	return &ExpressionEngine{}
}

// Evaluate evaluates the given expression string against the provided variables.
// Returns the result (typically bool) and an error if evaluation fails.
//
// Expressions support dot notation for nested variable access (e.g.,
// "applicant.age > 18"). This is internally converted to underscore notation
// (applicant_age) before evaluation, and the variables map is flattened
// using the same convention.
func (e *ExpressionEngine) Evaluate(expr string, vars map[string]interface{}) (interface{}, error) {
	if expr == "" {
		return nil, response.NewAppError(response.CodeWorkflowExprError, "expression is empty")
	}

	// Convert dot notation to underscore notation for govaluate compatibility.
	convertedExpr := convertDotNotation(expr)

	// Flatten variables to match the converted expression's parameter names.
	params := flattenParameters(vars)

	parsed, err := govaluate.NewEvaluableExpression(convertedExpr)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowExprError,
			"failed to parse expression '%s': %v", expr, err)
	}

	result, err := parsed.Evaluate(params)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowExprError,
			"failed to evaluate expression '%s': %v", expr, err)
	}

	return result, nil
}

// EvaluateBool evaluates an expression and returns the result as a bool.
// If the result is not a bool, it returns an error.
func (e *ExpressionEngine) EvaluateBool(expr string, vars map[string]interface{}) (bool, error) {
	result, err := e.Evaluate(expr, vars)
	if err != nil {
		return false, err
	}

	b, ok := result.(bool)
	if !ok {
		return false, response.NewAppErrorf(response.CodeWorkflowExprError,
			"expression '%s' returned non-boolean result: %v", expr, result)
	}
	return b, nil
}

// convertDotNotation converts dot-notation variable access (e.g., "applicant.age")
// to underscore notation (e.g., "applicant_age") for govaluate compatibility.
// govaluate parameter names can only contain letters, numbers, and underscores.
// Floating point numbers (e.g., "3.14") are not affected because the regex
// requires the first character to be a letter or underscore.
func convertDotNotation(expr string) string {
	return identifierChainRegex.ReplaceAllStringFunc(expr, func(match string) string {
		return strings.ReplaceAll(match, ".", "_")
	})
}

// flattenParameters converts a nested variables map into a flat map with
// underscore-separated keys. For example, {"applicant": {"age": 30}} becomes
// {"applicant_age": 30}. Top-level non-map values are passed through as-is.
func flattenParameters(vars map[string]interface{}) map[string]interface{} {
	params := make(map[string]interface{})
	for k, v := range vars {
		flattenInto(params, k, v)
	}
	return params
}

// flattenInto recursively flattens nested maps into the params map using
// underscore as the separator.
func flattenInto(params map[string]interface{}, prefix string, val interface{}) {
	if nested, ok := val.(map[string]interface{}); ok && len(nested) > 0 {
		for k, v := range nested {
			flattenInto(params, prefix+"_"+k, v)
		}
		return
	}
	params[prefix] = val
}

// EvaluateBranch evaluates multiple branch expressions and returns the first
// matching branch ID. If none match, the default branch is returned.
func (e *ExpressionEngine) EvaluateBranch(branches []BranchConfig, vars map[string]interface{}) (string, error) {
	var defaultBranch string

	for _, b := range branches {
		if b.IsDefault {
			defaultBranch = b.ID
			continue
		}

		matched, err := e.EvaluateBool(b.Expression, vars)
		if err != nil {
			return "", fmt.Errorf("branch '%s': %w", b.ID, err)
		}
		if matched {
			return b.ID, nil
		}
	}

	if defaultBranch != "" {
		return defaultBranch, nil
	}

	return "", response.NewAppError(response.CodeWorkflowExprError,
		"no branch matched and no default branch configured")
}

// BranchConfig represents a single branch in an exclusive gateway.
type BranchConfig struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Expression string `json:"expression"`
	IsDefault  bool   `json:"is_default"`
}

// ParseBranches extracts branch configs from a node's configuration.
func ParseBranches(config map[string]interface{}) ([]BranchConfig, error) {
	rawBranches, ok := config["branches"]
	if !ok {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"exclusive gateway missing 'branches' in config")
	}

	branchesJSON, err := json.Marshal(rawBranches)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to marshal branches: %v", err)
	}

	var branches []BranchConfig
	if err := json.Unmarshal(branchesJSON, &branches); err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to unmarshal branches: %v", err)
	}

	if len(branches) == 0 {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"exclusive gateway must have at least one branch")
	}

	// Validate that at most one default branch exists.
	defaultCount := 0
	for _, b := range branches {
		if b.Expression == "" && !b.IsDefault {
			return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
				"branch '%s' has no expression and is not default", b.ID)
		}
		if b.IsDefault {
			defaultCount++
		}
	}
	if defaultCount > 1 {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"exclusive gateway can have at most one default branch")
	}

	return branches, nil
}
