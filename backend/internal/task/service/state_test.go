package service

import "testing"

func TestCanTransit(t *testing.T) {
	cases := []struct {
		from, to int16
		ok       bool
	}{
		{0, 1, true},
		{0, 3, true},
		{0, 2, false},
		{0, 4, false},
		{1, 2, true},
		{1, 4, true},
		{1, 3, true},
		{1, 0, false},
		{4, 1, true},
		{4, 2, false},
		{4, 3, false},
		{2, 1, false},
		{3, 0, false},
		{1, 1, true},
	}
	for _, c := range cases {
		if got := CanTransit(c.from, c.to); got != c.ok {
			t.Fatalf("CanTransit(%d,%d)=%v want %v", c.from, c.to, got, c.ok)
		}
	}
}
