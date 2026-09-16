package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/jornada"
)

type espelhoReportRow struct {
	Date    string
	Weekday string
	In1     string
	BrkOut  string
	BrkBack string
	Out2    string
	Liquid  string
	Balance string
	Status  string
	Worked  time.Duration
	Delta   time.Duration
	HasData bool
}

// monthReport monta as linhas do espelho (CSV e PDF compartilham).
func (h *EspelhoHandler) monthReport(uid int64, month, now time.Time, loc *time.Location) []espelhoReportRow {
	monthEnd := month.AddDate(0, 1, 0)
	lastDay := monthEnd
	if month.Year() == now.Year() && month.Month() == now.Month() {
		lastDay = dayStart(now).Add(24 * time.Hour)
	}
	clk := func(t time.Time) string { return t.In(loc).Format("15:04") }
	var rows []espelhoReportRow
	for d := month; d.Before(monthEnd); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) || !d.Before(lastDay) {
			continue
		}
		row := espelhoReportRow{
			Date: d.Format("02/01/2006"), Weekday: ptWeekdaysShort[d.Weekday()],
			In1: "--:--", BrkOut: "--:--", BrkBack: "--:--", Out2: "--:--",
			Liquid: "--:--", Balance: "--:--", Status: "Sem registro",
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil {
			continue
		}
		folgaMin := folgaMinutes(punches)
		if len(punches) == 0 {
			rows = append(rows, row)
			continue
		}
		row.HasData = true
		main, _ := splitPunches(punches)
		for i, p := range main {
			switch i {
			case 0:
				row.In1 = clk(p.HappenedAt)
			case 1:
				row.BrkOut = clk(p.HappenedAt)
			case 2:
				row.BrkBack = clk(p.HappenedAt)
			case 3:
				row.Out2 = clk(p.HappenedAt)
			}
		}
		worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
		bal := worked - jornada.Goal - time.Duration(folgaMin)*time.Minute
		row.Worked, row.Delta = worked, bal
		row.Liquid = jornada.FmtHM(worked)
		row.Balance = jornada.FmtSigned(bal)
		row.Status = "Jornada regular"
		switch {
		case folgaMin > 0 && worked == 0:
			row.Status = fmt.Sprintf("Folga aprovada (%s)", jornada.FmtHM(time.Duration(folgaMin)*time.Minute))
			row.Liquid = "00h 00m"
		case bal > 10*time.Minute:
			row.Status = "Hora extra"
		case bal < -10*time.Minute:
			row.Status = "Débito"
		}
		if d.Equal(dayStart(now)) && len(main) < 4 {
			row.Status = "Em andamento"
		}
		rows = append(rows, row)
	}
	return rows
}

// contentHash gera SHA-256 determinístico do conteúdo apurado do mês.
func (h *EspelhoHandler) contentHash(uid int64, month, now time.Time) string {
	loc := now.Location()
	var sb strings.Builder
	for _, r := range h.monthReport(uid, month, now, loc) {
		fmt.Fprintf(&sb, "%s|%s|%s|%s|%s|%s|%s|%s\n",
			r.Date, r.In1, r.BrkOut, r.BrkBack, r.Out2, r.Liquid, r.Balance, r.Status)
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

// ExportPDF emite o espelho de ponto em PDF (Portaria MTE 671/2021).
func (h *EspelhoHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
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
	rows := h.monthReport(uid, month, now, loc)
	hash := h.contentHash(uid, month, now)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Espelho de Ponto - TempoSync", false)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 9, "TempoSync - Espelho de Ponto Eletronico", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, "Documento emitido nos termos da Portaria MTE 671/2021 (REP-P)", "", 1, "L", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Colaborador: %s", h.userEmail(uid)), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Mes de referencia: %s de %d", ptMonths[month.Month()-1], month.Year()), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Emitido em: %s", now.Format("02/01/2006 15:04")), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	headers := []string{"Data", "Dia", "Entrada", "Int.Saida", "Int.Ret.", "Saida", "Liquido", "Saldo", "Status"}
	widths := []float64{20, 10, 17, 17, 17, 17, 17, 20, 45}
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetFillColor(230, 237, 255)
	for i, hd := range headers {
		pdf.CellFormat(widths[i], 6, hd, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	var totalWorked time.Duration
	var totalDelta time.Duration
	pdf.SetFont("Helvetica", "", 8)
	for _, row := range rows {
		pdf.CellFormat(widths[0], 5.5, row.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 5.5, row.Weekday, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[2], 5.5, row.In1, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[3], 5.5, row.BrkOut, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[4], 5.5, row.BrkBack, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[5], 5.5, row.Out2, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[6], 5.5, row.Liquid, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[7], 5.5, row.Balance, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[8], 5.5, row.Status, "1", 1, "L", false, 0, "")
		totalWorked += row.Worked
		totalDelta += row.Delta
	}
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total trabalhado no mes: %s   |   Saldo do mes: %s",
		jornada.FmtHM(totalWorked), jornada.FmtSigned(totalDelta)), "", 1, "L", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 8)
	pdf.MultiCell(0, 4.5,
		fmt.Sprintf("Integridade: este documento foi gerado eletronicamente e o conteudo acima corresponde ao hash SHA-256 %s. "+
			"Alteracoes posteriores invalidam o hash. TempoSync - controle de jornada.", hash),
		"", "L", false)

	w.Header().Set("X-Document-Hash", hash)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"espelho-%s.pdf\"", month.Format("2006-01")))
	if err := pdf.Output(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
