package jornada

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseHHMM aceita apenas "HH:MM" com dois dígitos (ex: "08:00", "17:48").
func ParseHHMM(s string) (time.Duration, bool) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, true
}

// LiquidDay calcula a jornada líquida entre entrada, intervalo e saída.
func LiquidDay(in, brk, out time.Duration) time.Duration {
	total := out - in - brk
	if total < 0 {
		return 0
	}
	return total
}

func FmtHHMM(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	return fmt.Sprintf("%02d:%02d", int(d.Hours()), int(d.Minutes())%60)
}

func FmtHMShort(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	return fmt.Sprintf("%02dh %02dm", int(d.Hours()), int(d.Minutes())%60)
}
