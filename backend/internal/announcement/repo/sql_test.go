package repo

import (
	"strings"
	"testing"
)

func TestAudienceVisibleSQL_UsesJSONBExists(t *testing.T) {
	sql := audienceVisibleSQL()
	if strings.Contains(sql, "audience_ids ?") {
		t.Fatal("must not use JSONB ? operator; GORM treats it as a placeholder")
	}
	if !strings.Contains(sql, "jsonb_exists") {
		t.Fatal("expected jsonb_exists")
	}
	if n := strings.Count(sql, "?"); n != 3 {
		t.Fatalf("audienceVisibleSQL placeholders = %d, want 3", n)
	}
}
