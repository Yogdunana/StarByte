package repo

import "testing"

func TestNormalizePage(t *testing.T) {
	p, s := normalizePage(0, 0)
	if p != 1 || s != 20 {
		t.Fatalf("%d %d", p, s)
	}
	p, s = normalizePage(3, 200)
	if p != 3 || s != 20 {
		t.Fatalf("%d %d", p, s)
	}
	p, s = normalizePage(2, 50)
	if p != 2 || s != 50 {
		t.Fatalf("%d %d", p, s)
	}
}
