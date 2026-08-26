package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpressionEngine_Evaluate_SimpleComparison(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"amount":  5000,
		"days":    3,
		"approve": true,
	}

	tests := []struct {
		name   string
		expr   string
		expect bool
	}{
		{"amount > 1000", "amount > 1000", true},
		{"amount > 10000", "amount > 10000", false},
		{"days >= 3", "days >= 3", true},
		{"approve == true", "approve == true", true},
		{"approve == false", "approve == false", false},
		{"amount > 1000 && days >= 3", "amount > 1000 && days >= 3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.EvaluateBool(tt.expr, vars)
			require.NoError(t, err)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestExpressionEngine_Evaluate_NestedVariables(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"applicant": map[string]interface{}{
			"name":    "Alice",
			"age":     30,
			"dept":    "engineering",
			"manager": "Bob",
		},
	}

	result, err := engine.EvaluateBool("applicant.age > 18", vars)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = engine.EvaluateBool("applicant.dept == 'engineering'", vars)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestExpressionEngine_Evaluate_EmptyExpression(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{}

	_, err := engine.Evaluate("", vars)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "表达式不能为空")
}

func TestExpressionEngine_Evaluate_InvalidExpression(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{}

	_, err := engine.Evaluate("invalid expr @#$", vars)
	require.Error(t, err)
}

func TestExpressionEngine_EvaluateBool_NonBoolResult(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"amount": 5000,
	}

	_, err := engine.EvaluateBool("amount", vars)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "非布尔")
}

func TestEvaluateBranch_FirstMatch(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"amount": 5000,
	}

	branches := []BranchConfig{
		{ID: "b1", Expression: "amount > 10000"},
		{ID: "b2", Expression: "amount > 1000"},
		{ID: "b3", IsDefault: true},
	}

	result, err := engine.EvaluateBranch(branches, vars)
	require.NoError(t, err)
	assert.Equal(t, "b2", result)
}

func TestEvaluateBranch_DefaultBranch(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"amount": 100,
	}

	branches := []BranchConfig{
		{ID: "b1", Expression: "amount > 10000"},
		{ID: "b2", Expression: "amount > 1000"},
		{ID: "b3", IsDefault: true},
	}

	result, err := engine.EvaluateBranch(branches, vars)
	require.NoError(t, err)
	assert.Equal(t, "b3", result)
}

func TestEvaluateBranch_NoMatchNoDefault(t *testing.T) {
	engine := NewExpressionEngine()
	vars := map[string]interface{}{
		"amount": 100,
	}

	branches := []BranchConfig{
		{ID: "b1", Expression: "amount > 10000"},
		{ID: "b2", Expression: "amount > 1000"},
	}

	_, err := engine.EvaluateBranch(branches, vars)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "没有匹配的分支")
}

func TestParseBranches_Valid(t *testing.T) {
	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "b1",
				"label":      "High",
				"expression": "amount > 1000",
			},
			map[string]interface{}{
				"id":         "b2",
				"label":      "Default",
				"is_default": true,
			},
		},
	}

	branches, err := ParseBranches(config)
	require.NoError(t, err)
	assert.Len(t, branches, 2)
	assert.Equal(t, "b1", branches[0].ID)
	assert.Equal(t, "High", branches[0].Label)
	assert.False(t, branches[0].IsDefault)
	assert.True(t, branches[1].IsDefault)
}

func TestParseBranches_MissingBranches(t *testing.T) {
	config := map[string]interface{}{}

	_, err := ParseBranches(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少 'branches'")
}

func TestParseBranches_MultipleDefaults(t *testing.T) {
	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "b1",
				"is_default": true,
			},
			map[string]interface{}{
				"id":         "b2",
				"is_default": true,
			},
		},
	}

	_, err := ParseBranches(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "最多只能有一个默认分支")
}

func TestParseBranches_BranchWithoutExpressionOrDefault(t *testing.T) {
	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id": "b1",
			},
		},
	}

	_, err := ParseBranches(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "没有表达式且不是默认分支")
}

func TestConvertDotNotation_SimpleAccess(t *testing.T) {
	result := convertDotNotation("applicant.age > 18")
	assert.Equal(t, "applicant_age > 18", result)
}

func TestConvertDotNotation_ChainedAccess(t *testing.T) {
	result := convertDotNotation("user.profile.city == 'NYC'")
	assert.Equal(t, "user_profile_city == 'NYC'", result)
}

func TestConvertDotNotation_NoDots(t *testing.T) {
	result := convertDotNotation("amount > 1000")
	assert.Equal(t, "amount > 1000", result)
}

func TestConvertDotNotation_FloatNotConverted(t *testing.T) {
	result := convertDotNotation("3.14 > 3")
	assert.Equal(t, "3.14 > 3", result)
}

func TestConvertDotNotation_MixedExpression(t *testing.T) {
	result := convertDotNotation("amount > 1000 && applicant.age > 18")
	assert.Equal(t, "amount > 1000 && applicant_age > 18", result)
}
