package model

import (
	"encoding/json"
	"testing"
)

func TestAttachmentListJSONRoundTrip(t *testing.T) {
	src := AttachmentList{{FileID: "f1", Name: "a.pdf", Size: 3}}
	val, err := src.Value()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(src)
	if string(val.([]byte)) != string(raw) {
		t.Fatalf("value = %s", val)
	}
	var dest AttachmentList
	if err := dest.Scan(raw); err != nil || len(dest) != 1 || dest[0].Name != "a.pdf" {
		t.Fatalf("scan = %+v err=%v", dest, err)
	}
}

func TestIDListScanEmpty(t *testing.T) {
	var dest IDList
	if err := dest.Scan(nil); err != nil || dest == nil || len(dest) != 0 {
		t.Fatalf("nil scan = %+v err=%v", dest, err)
	}
}

func TestNormalizeAudience(t *testing.T) {
	if NormalizeAudience("") != AudienceAll || !ValidAudience("role") || ValidAudience("nope") {
		t.Fatal("audience helpers")
	}
}
