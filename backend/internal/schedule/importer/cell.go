package importer

import (
	"strings"
)

type sessionDraft struct {
	Title    string
	Teacher  string
	Weeks    []int
	Clock    clockRange
	HasClock bool
	Room     string
	Raw      string
}

// ParseCellSessions 解析课表格子，例如：
//
//	马克思主义基本原理 01
//	1-18周 王兰 08:00-10:30 【3-227】
//
// 同一格子可有多段（不同周次/老师），以换行分隔。
func ParseCellSessions(cell string) []sessionDraft {
	lines := splitCellLines(cell)
	if len(lines) == 0 {
		return nil
	}
	var out []sessionDraft
	title := ""
	for _, line := range lines {
		if looksLikeWeekLine(line) {
			sess := parseSessionLine(line)
			if sess.Title == "" {
				sess.Title = title
			}
			if sess.Title == "" {
				sess.Title = strings.TrimSpace(lines[0])
			}
			out = append(out, sess)
			continue
		}
		title = line
	}
	return out
}

func splitCellLines(cell string) []string {
	cell = strings.ReplaceAll(cell, "\r\n", "\n")
	cell = strings.ReplaceAll(cell, "\r", "\n")
	raw := strings.Split(cell, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func parseSessionLine(line string) sessionDraft {
	sess := sessionDraft{Raw: line, Weeks: ParseWeeks(line), Room: parseRoom(line)}
	if c, ok := parseClockRange(line); ok {
		sess.Clock = c
		sess.HasClock = true
	}
	teacher := stripDecorations(line)
	sess.Teacher = teacher
	return sess
}

func parseTitleRow(title string) Meta {
	meta := Meta{Title: strings.TrimSpace(title)}
	// 2026-2027学年 第一学期 段茗尧[1120240002] 课表
	year, rest, ok := strings.Cut(title, "学年")
	if ok {
		meta.AcademicYear = strings.TrimSpace(year)
		title = strings.TrimSpace(rest)
	}
	if i := strings.Index(title, "学期"); i >= 0 {
		// 含「第…学期」
		start := 0
		if j := strings.LastIndex(title[:i+len("学期")], "第"); j >= 0 {
			start = j
		}
		meta.Term = strings.TrimSpace(title[start : i+len("学期")])
		title = strings.TrimSpace(title[i+len("学期"):])
	}
	name, no, ok := cutBracketID(title)
	if ok {
		meta.StudentName = name
		meta.StudentNo = no
	}
	key := "timetable:smbu"
	if meta.StudentNo != "" {
		key += ":" + meta.StudentNo
	}
	if meta.AcademicYear != "" {
		key += ":" + meta.AcademicYear
	}
	if meta.Term != "" {
		key += ":" + meta.Term
	}
	meta.SourceKey = key
	if meta.StudentName != "" {
		meta.CalendarName = meta.StudentName + "的课表"
	} else {
		meta.CalendarName = "学校课表"
	}
	return meta
}

func cutBracketID(s string) (name, id string, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "课表")
	s = strings.TrimSpace(s)
	lb := strings.IndexAny(s, "[［")
	rb := strings.LastIndexAny(s, "]］")
	if lb < 0 || rb <= lb {
		return "", "", false
	}
	name = strings.TrimSpace(s[:lb])
	id = strings.TrimSpace(s[lb+1 : rb])
	return name, id, name != "" && id != ""
}
