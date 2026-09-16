package jornada

import (
	"fmt"
	"time"
)

const (
	Goal         = 8 * time.Hour
	DefaultBreak = time.Hour
)

type Punch struct {
	At time.Time
}

// Worked soma os pares (entrada→saída); com contagem ímpar, fecha o último
// par em now (expediente em andamento).
func Worked(punches []Punch, now time.Time) time.Duration {
	var total time.Duration
	for i := 0; i+1 < len(punches); i += 2 {
		total += punches[i+1].At.Sub(punches[i].At)
	}
	if len(punches)%2 == 1 {
		total += now.Sub(punches[len(punches)-1].At)
	}
	return total
}

// BreakTime retorna p3−p2 quando há ao menos 3 batidas.
func BreakTime(punches []Punch) time.Duration {
	if len(punches) < 3 {
		return 0
	}
	return punches[2].At.Sub(punches[1].At)
}

// ForecastExit projeta a saída: entrada + meta + intervalo (real ou padrão).
func ForecastExit(punches []Punch, goal time.Duration) (time.Time, bool) {
	if len(punches) == 0 {
		return time.Time{}, false
	}
	brk := BreakTime(punches)
	if brk == 0 {
		brk = DefaultBreak
	}
	return punches[0].At.Add(goal + brk), true
}

var nextLabels = []string{"Registrar Entrada", "Iniciar Intervalo", "Retornar do Intervalo", "Registrar Saída"}
var nextKinds = []string{"entrada", "saida_almoco", "retorno_almoco", "saida"}

func NextLabel(n int) string {
	if n >= len(nextLabels) {
		n = len(nextLabels) - 1
	}
	return nextLabels[n]
}

func NextKind(n int) string {
	if n >= len(nextKinds) {
		n = len(nextKinds) - 1
	}
	return nextKinds[n]
}

func Remaining(worked, goal time.Duration) time.Duration {
	if worked >= goal {
		return 0
	}
	return goal - worked
}

func Percent(worked, goal time.Duration) int {
	if goal <= 0 {
		return 0
	}
	p := int(worked * 100 / goal)
	if p > 100 {
		p = 100
	}
	return p
}

func FmtHM(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	return fmt.Sprintf("%02dh %02dm", int(d.Hours()), int(d.Minutes())%60)
}

func FmtSigned(d time.Duration) string {
	sign := "+"
	if d < 0 {
		sign = "-"
	}
	return sign + FmtHM(d)
}

func FmtClock(t time.Time) string {
	return t.Format("15:04")
}

func FmtClockS(t time.Time) string {
	return t.Format("15:04:05")
}
