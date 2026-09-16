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
	rows := [][]string{{
		"Data", "Dia", "Entrada 1", "Intervalo Saída", "Intervalo Retorno",
		"Saída 2", "Líquido", "Saldo", "Status",
	}}
	for _, rep := range h.monthReport(uid, month, now, loc) {
		rows = append(rows, []string{
			rep.Date, rep.Weekday, rep.In1, rep.BrkOut, rep.BrkBack, rep.Out2,
			rep.Liquid, rep.Balance, rep.Status,
		})
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
