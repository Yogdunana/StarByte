package service

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func DefaultWeightConfig() model.VoteWeightConfig {
	return model.VoteWeightConfig{Weights: map[string]float64{"president": 5, "minister": 3, "vice_minister": 2, "deputy": 2, "officer": 1}, DefaultWeight: 1}
}
func parseWeightConfig(raw string) (model.VoteWeightConfig, error) {
	var cfg model.VoteWeightConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, response.NewError(response.CodeBadRequest, "投票权重配置损坏，请管理员修复后再发起加权投票")
	}
	return cfg, validateWeightConfig(cfg)
}
func validateWeightConfig(cfg model.VoteWeightConfig) error {
	valid := func(w float64) bool {
		return !math.IsNaN(w) && !math.IsInf(w, 0) && w > 0 && w <= 999999.99 && math.Abs(w*100-math.Round(w*100)) < 0.000001
	}
	if !valid(cfg.DefaultWeight) || cfg.Weights == nil {
		return response.NewError(response.CodeBadRequest, "默认权重须为正数，最多两位小数，且不大于 999999.99")
	}
	for code, w := range cfg.Weights {
		if strings.TrimSpace(code) == "" || !valid(w) {
			return response.NewError(response.CodeBadRequest, "职务权重须为正数，最多两位小数，且不大于 999999.99")
		}
	}
	if canonical, ok := cfg.Weights["vice_minister"]; ok {
		if alias, ok := cfg.Weights["deputy"]; ok && canonical != alias {
			return response.NewError(response.CodeBadRequest, "副部长的 vice_minister 与 deputy 权重必须一致")
		}
	}
	return nil
}
func configuredWeight(cfg model.VoteWeightConfig, code string) (float64, bool) {
	if code == "deputy" || code == "vice_minister" {
		if weight, ok := cfg.Weights["vice_minister"]; ok {
			return weight, true
		}
		weight, ok := cfg.Weights["deputy"]
		return weight, ok
	}
	weight, ok := cfg.Weights[code]
	return weight, ok
}
func ResolveWeight(cfg model.VoteWeightConfig, position string, kind int16) float64 {
	if kind != model.VoteWeighted {
		return 1
	}
	if weight, ok := configuredWeight(cfg, position); ok {
		return weight
	}
	return cfg.DefaultWeight
}
