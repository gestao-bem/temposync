package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/jornada"
)

// writeCSV escreve planilha pt-BR: BOM UTF-8 (Excel) + ";" como separador.
func writeCSV(w http.ResponseWriter, filename string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte("\ufeff"))
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	for _, row := range rows {
		_ = cw.Write(row)
	}
	cw.Flush()
}

// ExportCSV do espelho: uma linha por dia útil apurado no mês.
func (h *EspelhoHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	now := saoPauloNow()
	loc := now.Location()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	if q := r.URL.Query().Get("month"); q != "" {
		if t, err := time.ParseInLocation("2006-01", q, loc); err == nil {
			month = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
	}
	monthEnd := month.AddDate(0, 1, 0)
	lastDay := monthEnd
	if month.Year() == now.Year() && month.Month() == now.Month() {
		lastDay = dayStart(now).Add(24 * time.Hour)
	}

	rows := [][]string{{
		"Data", "Dia", "Entrada 1", "Intervalo Saída", "Intervalo Retorno",
		"Saída 2", "Líquido", "Saldo", "Status",
	}}
	clk := func(t time.Time) string { return t.In(loc).Format("15:04") }
	for d := month; d.Before(monthEnd); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) || !d.Before(lastDay) {
			continue
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dia := ptWeekdaysShort[d.Weekday()]
		if len(punches) == 0 {
			rows = append(rows, []string{d.Format("02/01/2006"), dia, "--:--", "--:--", "--:--", "--:--", "--:--", "--:--", "Sem registro"})
			continue
		}
		cells := []string{d.Format("02/01/2006"), dia}
		for i := 0; i < 4; i++ {
			if i < len(punches) {
				cells = append(cells, clk(punches[i].HappenedAt))
			} else {
				cells = append(cells, "--:--")
			}
		}
		worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
		bal := worked - jornada.Goal
		status := "Jornada regular"
		if bal > 10*time.Minute {
			status = "Hora extra"
		} else if bal < -10*time.Minute {
			status = "Débito"
		}
		if d.Equal(dayStart(now)) && len(punches) < 4 {
			status = "Em andamento"
		}
		rows = append(rows, append(cells,
			jornada.FmtHM(worked), jornada.FmtSigned(bal), status))
	}
	writeCSV(w, "espelho-"+month.Format("2006-01")+".csv", rows)
}

// ExportCSV do banco de horas: lançamentos do ciclo semestral.
func (h *BancoHorasHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	now := saoPauloNow()
	loc := now.Location()
	cyStart, _ := semestre(now)
	days := h.collectDays(uid, cyStart, dayStart(now), now, loc)
	lancs := jornada.Lancamentos(days)

	rows := [][]string{{
		"Data", "Descrição", "Tipo", "Variação", "Saldo Parcial",
	}}
	for _, l := range lancs {
		label, tipo := "Jornada excedente", "Crédito"
		switch l.Kind {
		case "debito":
			label, tipo = "Compensação de jornada reduzida", "Débito"
		case "compensacao":
			label, tipo = "Folga / compensação de jornada", "Compensação"
		}
		rows = append(rows, []string{
			l.Day.Format("02/01/2006"),
			label,
			tipo,
			jornada.FmtSignedHM(l.Delta),
			jornada.FmtSignedHM(l.SaldoAfter),
		})
	}
	if len(lancs) == 0 {
		rows = append(rows, []string{"--", "Nenhum lançamento no ciclo", "--", "--", "--"})
	}
	writeCSV(w, fmt.Sprintf("banco-horas-%d-%s.csv", cyStart.Year(), now.Format("01")), rows)
}
