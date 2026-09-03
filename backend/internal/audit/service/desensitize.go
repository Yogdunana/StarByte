package service

import "regexp"

// compiledPatterns 缓存已编译的正则表达式，避免重复编译
type patternCache struct {
	patterns map[string]*regexp.Regexp
}

var compiledPatterns = &patternCache{
	patterns: make(map[string]*regexp.Regexp),
}

func (pc *patternCache) get(pattern string) *regexp.Regexp {
	if re, ok := pc.patterns[pattern]; ok {
		return re
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	pc.patterns[pattern] = re
	return re
}
