package utils

import "testing"

func TestPluralS(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}

	for _, tc := range tests {
		got := PluralS(tc.n)
		if got != tc.want {
			t.Errorf("PluralS(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

func TestFormatMinutes(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{0, "0 minutes"},
		{1, "1 minute"},
		{2, "2 minutes"},
		{59, "59 minutes"},
		{60, "1 hour"},
		{61, "1 hour 1 minute"},
		{90, "1 hour 30 minutes"},
		{120, "2 hours"},
		{121, "2 hours 1 minute"},
		{125, "2 hours 5 minutes"},
	}

	for _, tc := range tests {
		got := FormatMinutes(tc.minutes)
		if got != tc.want {
			t.Errorf("FormatMinutes(%d) = %q, want %q", tc.minutes, got, tc.want)
		}
	}
}
