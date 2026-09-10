package service

import (
	"math"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

type optionAgg struct {
	Count  int
	Weight float64
}

func CalculateVoteResult(records []model.VoteRecord) (map[string]optionAgg, int, float64) {
	agg := map[string]optionAgg{}
	var totalWeight float64
	for _, r := range records {
		cur := agg[r.OptionKey]
		cur.Count++
		cur.Weight = math.Round((cur.Weight+r.Weight)*100) / 100
		agg[r.OptionKey] = cur
		totalWeight = math.Round((totalWeight+r.Weight)*100) / 100
	}
	return agg, len(records), totalWeight
}
