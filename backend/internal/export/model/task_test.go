package model

import "testing"

func TestKeys(t *testing.T) {
	if TaskKey("a") != "export:task:a" {
		t.Fatalf("TaskKey")
	}
	if FileKey("b") != "export:file:b" {
		t.Fatalf("FileKey")
	}
	if BlobKey("c") != "export:blob:c" {
		t.Fatalf("BlobKey")
	}
}
