package importer

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	weekTokenRe = regexp.MustCompile(`(\d+)(?:\s*[-–~]\s*(\d+))?\s*周(?:\s*[\(（]([单双])[\)）])?`)
	weekHintRe  = regexp.MustCompile(`\d+\s*(?:[-–~]\s*\d+)?\s*周`)
)

func looksLikeWeekLine(s string) bool {
	return weekHintRe.MatchString(s)
}

// ParseWeeks 解析「1-18周」「1-17周(单)」「2-18周(双)」「1-7周,9-18周」「10周」。
func ParseWeeks(raw string) []int {
	seen := map[int]struct{}{}
	var out []int
	for _, m := range weekTokenRe.FindAllStringSubmatch(raw, -1) {
		from, _ := strconv.Atoi(m[1])
		to := from
		if m[2] != "" {
			to, _ = strconv.Atoi(m[2])
		}
		if from <= 0 || to <= 0 {
			continue
		}
		if to < from {
			from, to = to, from
		}
		parity := strings.TrimSpace(m[3])
		for w := from; w <= to; w++ {
			if parity == "单" && w%2 == 0 {
				continue
			}
			if parity == "双" && w%2 == 1 {
				continue
			}
			if _, ok := seen[w]; ok {
				continue
			}
			seen[w] = struct{}{}
			out = append(out, w)
		}
	}
	return out
}
