package jornada

import (
	"fmt"
	"time"
)

// ToleranceDaily é a faixa de variação do Art. 58 CLT: até 10min por dia não
// contam como crédito nem débito de banco de horas.
const ToleranceDaily = 10 * time.Minute

type DayBalance struct {
	Day   time.Time
	Delta time.Duration
}

type Lancamento struct {
	Day        time.Time
	Delta      time.Duration
	SaldoAfter time.Duration
	Kind       string // credito | debito | compensacao
}

// EffectiveDelta zera variações dentro da tolerância, preservando o sinal.
func EffectiveDelta(delta time.Duration) time.Duration {
	if delta < 0 {
		if -delta <= ToleranceDaily {
			return 0
		}
		return delta
	}
	if delta <= ToleranceDaily {
		return 0
	}
	return delta
}

func effective(d time.Duration) time.Duration {
	return EffectiveDelta(d)
}

func Saldo(days []DayBalance) time.Duration {
	var total time.Duration
	for _, d := range days {
		total += effective(d.Delta)
	}
	return total
}

func Creditos(days []DayBalance) time.Duration {
	var total time.Duration
	for _, d := range days {
		if e := effective(d.Delta); e > 0 {
			total += e
		}
	}
	return total
}

func Debitos(days []DayBalance) time.Duration {
	var total time.Duration
	for _, d := range days {
		if e := effective(d.Delta); e < 0 {
			total -= e
		}
	}
	return total
}

// Lancamentos retorna os dias com variação relevante, do mais recente ao mais
// antigo, com o saldo acumulado naquele ponto (calculado em ordem cronológica).
func Lancamentos(days []DayBalance) []Lancamento {
	out := make([]Lancamento, 0, len(days))
	var running time.Duration
	for _, d := range days {
		e := effective(d.Delta)
		if e == 0 {
			continue
		}
		running += e
		kind := "credito"
		if e < 0 {
			kind = "debito"
			if -e >= 4*time.Hour {
				kind = "compensacao"
			}
		}
		out = append(out, Lancamento{Day: d.Day, Delta: e, SaldoAfter: running, Kind: kind})
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func MediaDiaria(days []DayBalance) time.Duration {
	var count int
	var total time.Duration
	for _, d := range days {
		if d.Delta == 0 {
			continue
		}
		count++
		total += effective(d.Delta)
	}
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

func Pct(value, total time.Duration) int {
	if total <= 0 || value <= 0 {
		return 0
	}
	p := int(value * 100 / total)
	if p > 100 {
		return 100
	}
	return p
}

func FmtSignedHM(d time.Duration) string {
	sign := "+"
	if d < 0 {
		sign = "-"
		d = -d
	}
	return fmt.Sprintf("%s%02dh %02dm", sign, int(d.Hours()), int(d.Minutes())%60)
}
