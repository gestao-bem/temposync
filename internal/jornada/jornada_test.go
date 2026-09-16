package jornada

import (
	"testing"
	"time"
)

func marks(day time.Time, hms ...int) []Punch {
	var out []Punch
	for i := 0; i < len(hms); i += 3 {
		out = append(out, Punch{At: day.Add(time.Duration(hms[i])*time.Hour +
			time.Duration(hms[i+1])*time.Minute +
			time.Duration(hms[i+2])*time.Second)})
	}
	return out
}

func TestWorked_FourPunches(t *testing.T) {
	day := time.Date(2024, 10, 24, 0, 0, 0, 0, time.UTC)
	p := marks(day, 8, 32, 10, 12, 5, 4, 13, 10, 18)
	// (12:05:04-08:32:10) + (18:00-13:10:18)
	want := 8*time.Hour + 22*time.Minute + 36*time.Second
	if got := Worked(p, day.Add(18*time.Hour)); got != want {
		t.Errorf("worked = %v, want %v", got, want)
	}
}

func TestForecast_ThreePunches(t *testing.T) {
	day := time.Date(2024, 10, 24, 0, 0, 0, 0, time.UTC)
	p := marks(day, 8, 32, 0, 12, 15, 0, 13, 20, 0)
	got, ok := ForecastExit(p, 8*time.Hour)
	if !ok {
		t.Fatal("no forecast")
	}
	want := day.Add(17*time.Hour + 37*time.Minute)
	if !got.Equal(want) {
		t.Errorf("forecast = %v, want %v", got, want)
	}
}

func TestNextLabel_Cycles(t *testing.T) {
	want := []string{"Registrar Entrada", "Iniciar Intervalo", "Retornar do Intervalo", "Registrar Saída", "Registrar Saída"}
	for i, w := range want {
		if got := NextLabel(i); got != w {
			t.Errorf("label(%d) = %q, want %q", i, got, w)
		}
	}
}

func TestFmtHM(t *testing.T) {
	if got := FmtHM(5*time.Hour + 28*time.Minute); got != "05h 28m" {
		t.Errorf("got %q", got)
	}
	if got := FmtSigned(4*time.Hour + 45*time.Minute); got != "+04h 45m" {
		t.Errorf("got %q", got)
	}
	if got := FmtSigned(-30 * time.Minute); got != "-00h 30m" {
		t.Errorf("got %q", got)
	}
}
