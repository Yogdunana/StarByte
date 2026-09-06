package repo

import "testing"

func TestTruncExpr(t *testing.T) {
	if got := truncExpr("created_at", "day"); got != "to_char(date_trunc('day', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("day: %s", got)
	}
	if got := truncExpr("m.start_time", "week"); got != "to_char(date_trunc('week', m.start_time), 'YYYY-MM-DD')" {
		t.Fatalf("week: %s", got)
	}
	if got := truncExpr("created_at", "MONTH"); got != "to_char(date_trunc('month', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("month: %s", got)
	}
	if got := truncExpr("created_at", ""); got != "to_char(date_trunc('month', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("default: %s", got)
	}
}
