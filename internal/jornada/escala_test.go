package jornada

import (
	"testing"
	"time"
)

func TestParseHHMM(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"08:00", 8 * time.Hour, true},
		{"17:48", 17*time.Hour + 48*time.Minute, true},
		{"1:5", 0, false},
		{"8:00", 0, false},
		{"25:00", 0, false},
		{"08:60", 0, false},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := ParseHHMM(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseHHMM(%q) = %v,%v want %v,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestLiquidDay(t *testing.T) {
	in := 8 * time.Hour
	brk := time.Hour
	out := 17*time.Hour + 48*time.Minute
	if got := LiquidDay(in, brk, out); got != 8*time.Hour+48*time.Minute {
		t.Errorf("LiquidDay = %v, want 8h48m", got)
	}
	// saída antes da entrada → 0
	if got := LiquidDay(out, brk, in); got != 0 {
		t.Errorf("LiquidDay invertido = %v, want 0", got)
	}
}

func TestFmtHMShort(t *testing.T) {
	if got := FmtHMShort(8*time.Hour + 48*time.Minute); got != "08h 48m" {
		t.Errorf("got %q", got)
	}
	if got := FmtHMShort(0); got != "00h 00m" {
		t.Errorf("got %q", got)
	}
}
