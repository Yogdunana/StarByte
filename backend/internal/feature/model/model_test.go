package model

import (
	"testing"
)

func TestValidType(t *testing.T) {
	if !ValidType(TypeBoolean) || ValidType("ab_test") {
		t.Fatal("valid type")
	}
}

func TestRulesValueScan(t *testing.T) {
	r := Rules{UserIDs: []string{"u1"}, Percent: 10, Salt: "s"}
	v, err := r.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out Rules
	if err := out.Scan(v); err != nil || out.Percent != 10 || out.Salt != "s" {
		t.Fatalf("%+v %v", out, err)
	}
	if err := out.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if err := out.Scan("{\"percent\":5}"); err != nil || out.Percent != 5 {
		t.Fatalf("string scan: %+v %v", out, err)
	}
	if err := out.Scan([]byte{}); err != nil {
		t.Fatal(err)
	}
	if err := out.Scan(1); err == nil {
		t.Fatal("expected type error")
	}
	_ = Flag{}.TableName()
	_ = Audit{}.TableName()
}
