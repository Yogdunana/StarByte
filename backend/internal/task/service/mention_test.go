package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseMentionNames(t *testing.T) {
	got := parseMentionNames("请看 @alice 和 @bob，以及重复 @alice")
	want := []string{"alice", "bob"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestEncodeDecodeJSONList(t *testing.T) {
	raw := encodeJSONList([]string{"a", "b"})
	got := decodeJSONList(raw)
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("roundtrip %v", got)
	}
	if len(decodeJSONList("")) != 0 {
		t.Fatal("empty")
	}
	if !reflect.DeepEqual(decodeJSONList("x,y"), []string{"x", "y"}) {
		t.Fatal("csv fallback")
	}
}

func TestEncodeTagsTruncates(t *testing.T) {
	tags := make([]string, 80)
	for i := range tags {
		tags[i] = "tag-name-xxxxxxxx"
	}
	s := encodeTags(tags)
	if len(s) > 500 {
		t.Fatalf("len=%d", len(s))
	}
	if !json.Valid([]byte(s)) {
		t.Fatalf("invalid json %s", s)
	}
}
