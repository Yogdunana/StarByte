package service

import (
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func DefaultWeightConfig() model.VoteWeightConfig {
	return model.VoteWeightConfig{
		Weights: map[string]float64{
			"president":      5,
			"vice_president": 4,
			"minister":       3,
			"deputy":         2,
			"vice_minister":  2,
			"officer":        1,
		},
		DefaultWeight: 1,
	}
}

func parseWeightConfig(raw string) model.VoteWeightConfig {
	cfg := DefaultWeightConfig()
	if raw == "" {
		return cfg
	}
	var parsed model.VoteWeightConfig
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return cfg
	}
	if parsed.Weights == nil {
		parsed.Weights = cfg.Weights
	}
	if parsed.DefaultWeight <= 0 {
		parsed.DefaultWeight = cfg.DefaultWeight
	}
	return parsed
}

func ResolveWeight(cfg model.VoteWeightConfig, positionCode string, voteType int16) float64 {
	if voteType != model.VoteWeighted {
		return 1
	}
	if w, ok := cfg.Weights[positionCode]; ok {
		return w
	}
	if positionCode == "vice_minister" {
		if w, ok := cfg.Weights["deputy"]; ok {
			return w
		}
	}
	if positionCode == "deputy" {
		if w, ok := cfg.Weights["vice_minister"]; ok {
			return w
		}
	}
	if cfg.DefaultWeight > 0 {
		return cfg.DefaultWeight
	}
	return 1
}
