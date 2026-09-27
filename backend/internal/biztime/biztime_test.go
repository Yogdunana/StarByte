package biztime

import (
	"testing"
	"time"
)

// 数据库里的时间戳：pgx 读出来挂在 UTC 上，代表北京时间 2026-09-27 14:41:07。
func dbTime() time.Time {
	return time.Date(2026, 9, 27, 6, 41, 7, 0, time.UTC)
}

func TestFormat_UsesShanghaiNotUTC(t *testing.T) {
	// 这就是运维同学踩到的那个坑：直接 Format 打出来是 06:41，实际是 14:41。
	if got := dbTime().Format("2006-01-02 15:04:05"); got != "2026-09-27 06:41:07" {
		t.Fatalf("前提不成立：原生 Format 应给出 UTC 墙钟，实际 %q", got)
	}
	if got := Format(dbTime(), "2006-01-02 15:04:05"); got != "2026-09-27 14:41:07" {
		t.Errorf("Format = %q, want %q", got, "2026-09-27 14:41:07")
	}
}

// 跨日边界：UTC 16:30 已经是北京时间第二天凌晨，日期必须跟着翻。
func TestFormat_DateRollsOverAtBeijingMidnight(t *testing.T) {
	utc := time.Date(2026, 9, 27, 16, 30, 0, 0, time.UTC)
	if got := Format(utc, "2006-01-02"); got != "2026-09-28" {
		t.Errorf("Format = %q, want 2026-09-28", got)
	}
	if got := Format(utc, "15:04"); got != "00:30" {
		t.Errorf("Format = %q, want 00:30", got)
	}
}

func TestFormatPtr(t *testing.T) {
	v := dbTime()
	if got := FormatPtr(&v, "2006-01-02 15:04", "-"); got != "2026-09-27 14:41" {
		t.Errorf("FormatPtr = %q, want 2026-09-27 14:41", got)
	}
	if got := FormatPtr(nil, "2006-01-02 15:04", "待定"); got != "待定" {
		t.Errorf("FormatPtr(nil) = %q, want 待定", got)
	}
}

func TestOffsetIsPlus8(t *testing.T) {
	_, off := In(dbTime()).Zone()
	if off != 8*3600 {
		t.Errorf("Zone offset = %d, want %d", off, 8*3600)
	}
}

func TestNowIsShanghai(t *testing.T) {
	if name, _ := Now().Zone(); name != "CST" && name != "+0800" {
		// LoadLocation 成功时 Zone 名是 CST，退化到 FixedZone 时也是 CST。
		t.Errorf("Now().Zone() = %q, want CST", name)
	}
	if _, off := Now().Zone(); off != 8*3600 {
		t.Errorf("Now offset = %d, want %d", off, 8*3600)
	}
}

func TestLoc(t *testing.T) {
	if got := Loc().String(); got != Timezone && got != "CST" {
		t.Errorf("Loc() = %q", got)
	}
}
