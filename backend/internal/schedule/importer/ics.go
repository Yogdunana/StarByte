package importer

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

// ICSImporter 解析 RFC 5545 风格的 .ics（VEVENT）。不展开复杂 RRULE，仅导入 DTSTART/DTEND。
type ICSImporter struct{}

func (ICSImporter) Kind() string { return "ics" }

func (ICSImporter) Parse(_ context.Context, raw []byte, opts Options) ([]DraftEvent, Meta, error) {
	text := unfoldICS(string(raw))
	if !strings.Contains(text, "BEGIN:VEVENT") {
		return nil, Meta{}, fmt.Errorf("not an ics calendar")
	}
	blocks := splitICSBlocks(text, "VEVENT")
	loc := locationOrShanghai(opts.Timezone)
	var events []DraftEvent
	calName := icsValue(text, "X-WR-CALNAME")
	if calName == "" {
		calName = strings.TrimSuffix(opts.Filename, ".ics")
	}
	if strings.TrimSpace(calName) == "" {
		calName = "导入日历"
	}
	for _, block := range blocks {
		title := icsUnescape(icsValue(block, "SUMMARY"))
		if title == "" {
			title = "(无标题)"
		}
		startVal, startParams := icsParamValue(block, "DTSTART")
		start, allDay, ok := parseICSTime(startVal, startParams, loc)
		if !ok {
			continue
		}
		endVal, endParams := icsParamValue(block, "DTEND")
		end, _, endOK := parseICSTime(endVal, endParams, loc)
		if !endOK {
			if allDay {
				end = start.Add(24 * time.Hour)
			} else {
				end = start.Add(time.Hour)
			}
		}
		uid := icsValue(block, "UID")
		if uid == "" {
			uid = fmt.Sprintf("%x", sha1.Sum([]byte(title+start.String())))
		}
		events = append(events, DraftEvent{
			Title:       title,
			Description: icsUnescape(icsValue(block, "DESCRIPTION")),
			Location:    icsUnescape(icsValue(block, "LOCATION")),
			StartAt:     start,
			EndAt:       end,
			AllDay:      allDay,
			ExternalUID: "ics:" + uid,
		})
	}
	key := fmt.Sprintf("import:ics:%x", sha1.Sum([]byte(opts.Filename+"|"+calName)))
	return events, Meta{
		Title:        calName,
		CalendarName: calName,
		SourceKey:    key,
		Timezone:     loc.String(),
	}, nil
}

func unfoldICS(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n ", "")
	s = strings.ReplaceAll(s, "\n\t", "")
	return s
}

func splitICSBlocks(text, name string) []string {
	begin := "BEGIN:" + name
	end := "END:" + name
	var out []string
	for {
		i := strings.Index(text, begin)
		if i < 0 {
			break
		}
		text = text[i+len(begin):]
		j := strings.Index(text, end)
		if j < 0 {
			out = append(out, text)
			break
		}
		out = append(out, text[:j])
		text = text[j+len(end):]
	}
	return out
}

func icsValue(block, key string) string {
	v, _ := icsParamValue(block, key)
	return v
}

func icsParamValue(block, key string) (string, string) {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		name, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		base, params, _ := strings.Cut(name, ";")
		if !strings.EqualFold(base, key) {
			continue
		}
		return rest, params
	}
	return "", ""
}

func parseICSTime(value, params string, loc *time.Location) (time.Time, bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, false
	}
	if strings.Contains(strings.ToUpper(params), "VALUE=DATE") || (len(value) == 8 && !strings.Contains(value, "T")) {
		t, err := time.ParseInLocation("20060102", value, loc)
		return t, true, err == nil
	}
	if strings.HasSuffix(value, "Z") {
		t, err := time.Parse("20060102T150405Z", value)
		return t, false, err == nil
	}
	if tzid := icsTZID(params); tzid != "" {
		if z, err := time.LoadLocation(tzid); err == nil {
			loc = z
		}
	}
	t, err := time.ParseInLocation("20060102T150405", value, loc)
	if err != nil {
		t, err = time.ParseInLocation("20060102T1504", value, loc)
	}
	return t, false, err == nil
}

func icsTZID(params string) string {
	for _, part := range strings.Split(params, ";") {
		k, v, ok := strings.Cut(part, "=")
		if ok && strings.EqualFold(strings.TrimSpace(k), "TZID") {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func icsUnescape(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
