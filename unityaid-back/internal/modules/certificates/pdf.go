package certificates

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

//go:embed assets/DejaVuSans.ttf
var regularFont []byte

//go:embed assets/DejaVuSans-Bold.ttf
var boldFont []byte

func BuildPDF(item Certificate) []byte {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddUTF8FontFromBytes("DejaVu", "", regularFont)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", boldFont)
	pdf.AddPage()

	drawHeader(pdf, item)
	drawCertificateBody(pdf, item)
	drawVerificationBlock(pdf, item.VerifyCode)
	drawFooter(pdf)

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil
	}
	return buffer.Bytes()
}

func drawHeader(pdf *gofpdf.Fpdf, item Certificate) {
	subtitle := "Сертификат волонтерской активности"
	if item.Type == "hours" {
		subtitle = "Справка о волонтерских часах"
	}
	pdf.SetFillColor(47, 160, 111)
	pdf.Rect(0, 0, 210, 42, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("DejaVu", "B", 22)
	pdf.SetXY(18, 14)
	pdf.CellFormat(0, 9, "Пульс Добра", "", 1, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 10)
	pdf.SetX(18)
	pdf.CellFormat(0, 6, subtitle, "", 1, "L", false, 0, "")
}

func drawCertificateBody(pdf *gofpdf.Fpdf, item Certificate) {
	pdf.SetTextColor(22, 33, 58)
	pdf.SetY(58)
	pdf.SetFont("DejaVu", "B", 24)
	pdf.MultiCell(174, 10, clean(item.Title), "", "L", false)

	if strings.TrimSpace(item.Description) != "" {
		pdf.Ln(4)
		pdf.SetFont("DejaVu", "", 11)
		pdf.SetTextColor(83, 96, 121)
		pdf.MultiCell(174, 7, clean(item.Description), "", "L", false)
	}

	pdf.Ln(8)
	drawInfoCard(pdf, "Получатель", item.UserName)
	if item.OrganizationName != nil && strings.TrimSpace(*item.OrganizationName) != "" {
		drawInfoCard(pdf, "Организация", *item.OrganizationName)
	}
	if item.Type == "hours" {
		drawInfoCard(pdf, "Подтвержденные часы", fmt.Sprintf("%.1f ч.", item.TotalHours))
	} else {
		drawParticipationEvents(pdf, item.Events)
	}
	drawInfoCard(pdf, "Дата выдачи", item.IssuedAt.Format("02.01.2006"))
}

func drawParticipationEvents(pdf *gofpdf.Fpdf, events []ParticipationEvent) {
	pdf.SetFont("DejaVu", "B", 11)
	pdf.SetTextColor(22, 33, 58)
	pdf.CellFormat(174, 7, "Мероприятия участника", "", 1, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 9.5)
	pdf.SetTextColor(83, 96, 121)
	if len(events) == 0 {
		pdf.MultiCell(174, 6, "Подтвержденные мероприятия не найдены.", "", "L", false)
		pdf.Ln(4)
		return
	}
	for index, event := range events {
		if index >= 8 {
			pdf.MultiCell(174, 6, fmt.Sprintf("И еще %d мероприятие(й).", len(events)-index), "", "L", false)
			break
		}
		parts := []string{event.StartsAt.Format("02.01.2006")}
		if event.Location != nil && strings.TrimSpace(*event.Location) != "" {
			parts = append(parts, strings.TrimSpace(*event.Location))
		}
		if event.OrganizationName != nil && strings.TrimSpace(*event.OrganizationName) != "" {
			parts = append(parts, strings.TrimSpace(*event.OrganizationName))
		}
		pdf.SetFont("DejaVu", "B", 9.5)
		pdf.SetTextColor(22, 33, 58)
		pdf.MultiCell(174, 5.5, "• "+clean(event.Title), "", "L", false)
		pdf.SetFont("DejaVu", "", 8.5)
		pdf.SetTextColor(100, 112, 138)
		pdf.MultiCell(174, 5, strings.Join(parts, " · "), "", "L", false)
		pdf.Ln(1)
	}
	pdf.Ln(3)
}

func drawVerificationBlock(pdf *gofpdf.Fpdf, verifyCode string) {
	pdf.Ln(6)
	pdf.SetFillColor(244, 248, 246)
	pdf.SetDrawColor(217, 229, 224)
	pdf.RoundedRect(18, pdf.GetY(), 174, 30, 3, "1234", "FD")
	pdf.SetXY(24, pdf.GetY()+6)
	pdf.SetFont("DejaVu", "", 9)
	pdf.SetTextColor(83, 96, 121)
	pdf.CellFormat(0, 5, "Код проверки", "", 1, "L", false, 0, "")
	pdf.SetX(24)
	pdf.SetFont("DejaVu", "B", 16)
	pdf.SetTextColor(22, 33, 58)
	pdf.CellFormat(0, 8, verifyCode, "", 1, "L", false, 0, "")
}

func drawFooter(pdf *gofpdf.Fpdf) {
	pdf.SetY(255)
	pdf.SetDrawColor(217, 229, 224)
	pdf.Line(18, 252, 192, 252)
	pdf.SetFont("DejaVu", "", 9)
	pdf.SetTextColor(100, 112, 138)
	pdf.MultiCell(174, 5, "Проверить подлинность сертификата можно в системе Пульс по коду проверки. Документ сформирован автоматически.", "", "L", false)
}

func drawInfoCard(pdf *gofpdf.Fpdf, label string, value string) {
	y := pdf.GetY()
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(18, y, 174, 20, 3, "1234", "FD")
	pdf.SetXY(24, y+4)
	pdf.SetFont("DejaVu", "", 8.5)
	pdf.SetTextColor(100, 112, 138)
	pdf.CellFormat(54, 5, label, "", 0, "L", false, 0, "")
	pdf.SetFont("DejaVu", "B", 11)
	pdf.SetTextColor(22, 33, 58)
	pdf.MultiCell(110, 5, clean(value), "", "L", false)
	pdf.SetY(y + 24)
}

func clean(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Не указано"
	}
	return value
}
