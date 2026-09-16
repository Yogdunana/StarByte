package phone

import "testing"

func TestNormalizeCN(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"13800138000", "13800138000", false},
		{"+86 138-0013-8000", "13800138000", false},
		{"8613800138000", "13800138000", false},
		{"008613800138000", "13800138000", false},
		{"12800138000", "", true},
		{"1380013800", "", true},
		{"", "", true},
		{"abc", "", true},
	}
	for _, tc := range cases {
		got, err := NormalizeCN(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%q: expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeCNOptional(t *testing.T) {
	t.Parallel()
	got, err := NormalizeCNOptional("  ")
	if err != nil || got != "" {
		t.Fatalf("empty should pass, got %q %v", got, err)
	}
}
