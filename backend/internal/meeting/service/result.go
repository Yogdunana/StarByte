package service

import "github.com/Yogdunana/StarByte/backend/internal/meeting/model"

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
		cur.Weight += r.Weight
		agg[r.OptionKey] = cur
		totalWeight += r.Weight
	}
	return agg, len(records), totalWeight
}
