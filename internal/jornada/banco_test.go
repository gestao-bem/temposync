package jornada

import (
	"testing"
	"time"
)

func day(n int) time.Time {
	return time.Date(2024, 10, n, 0, 0, 0, 0, time.UTC)
}

func TestEffectiveDelta_AppliesCLTTolerance(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want time.Duration
	}{
		{10 * time.Minute, 0},
		{-10 * time.Minute, 0},
		{11 * time.Minute, 11 * time.Minute},
		{-11 * time.Minute, -11 * time.Minute},
		{2 * time.Hour, 2 * time.Hour},
	}
	for _, c := range cases {
		if got := EffectiveDelta(c.in); got != c.want {
			t.Errorf("EffectiveDelta(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSaldo_SumsEffectiveDeltas(t *testing.T) {
	days := []DayBalance{
		{Day: day(1), Delta: 2 * time.Hour},
		{Day: day(2), Delta: -5 * time.Minute}, // tolerância
		{Day: day(3), Delta: -time.Hour},
	}
	want := time.Hour
	if got := Saldo(days); got != want {
		t.Errorf("Saldo = %v, want %v", got, want)
	}
}

func TestCreditosDebitos(t *testing.T) {
	days := []DayBalance{
		{Day: day(1), Delta: 2 * time.Hour},
		{Day: day(2), Delta: 90 * time.Minute},
		{Day: day(3), Delta: -3 * time.Hour},
		{Day: day(4), Delta: -4 * time.Minute},
	}
	if got := Creditos(days); got != 3*time.Hour+30*time.Minute {
		t.Errorf("Creditos = %v", got)
	}
	if got := Debitos(days); got != 3*time.Hour {
		t.Errorf("Debitos = %v", got)
	}
}

func TestLancamentos_AreDescendingWithRunningBalance(t *testing.T) {
	days := []DayBalance{
		{Day: day(1), Delta: 2 * time.Hour},
		{Day: day(2), Delta: 5 * time.Minute}, // ignorado
		{Day: day(3), Delta: -time.Hour},
		{Day: day(4), Delta: -8 * time.Hour},
	}
	got := Lancamentos(days)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if !got[0].Day.Equal(day(4)) || !got[2].Day.Equal(day(1)) {
		t.Errorf("ordem errada: %v..%v", got[0].Day, got[2].Day)
	}
	if got[0].SaldoAfter != time.Hour-8*time.Hour {
		t.Errorf("SaldoAfter[0] = %v, want -7h", got[0].SaldoAfter)
	}
	if got[2].SaldoAfter != 2*time.Hour {
		t.Errorf("SaldoAfter[2] = %v, want +2h", got[2].SaldoAfter)
	}
	if got[0].Kind != "compensacao" || got[2].Kind != "credito" || got[1].Kind != "debito" {
		t.Errorf("kinds = %s/%s/%s", got[0].Kind, got[1].Kind, got[2].Kind)
	}
}

func TestMediaDiaria(t *testing.T) {
	days := []DayBalance{
		{Day: day(1), Delta: time.Hour},
		{Day: day(2), Delta: 2 * time.Hour},
		{Day: day(3), Delta: 0}, // dia sem batida não entra
	}
	if got := MediaDiaria(days); got != 90*time.Minute {
		t.Errorf("MediaDiaria = %v, want 90m", got)
	}
}

func TestPct_ClampsTo100(t *testing.T) {
	if got := Pct(15*time.Hour, 30*time.Hour); got != 50 {
		t.Errorf("Pct = %d, want 50", got)
	}
	if got := Pct(40*time.Hour, 30*time.Hour); got != 100 {
		t.Errorf("Pct = %d, want 100", got)
	}
	if got := Pct(-3*time.Hour, 30*time.Hour); got != 0 {
		t.Errorf("Pct negativo = %d, want 0", got)
	}
}
