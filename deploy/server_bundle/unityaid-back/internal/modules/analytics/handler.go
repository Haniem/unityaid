package analytics

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

//go:embed assets/DejaVuSans.ttf
var analyticsRegularFont []byte

//go:embed assets/DejaVuSans-Bold.ttf
var analyticsBoldFont []byte

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(c *gin.Context) {
	item, err := h.service.Overview(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load overview report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Volunteers(c *gin.Context) {
	item, err := h.service.Volunteers(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load volunteers report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Events(c *gin.Context) {
	item, err := h.service.Events(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load events report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Tasks(c *gin.Context) {
	item, err := h.service.Tasks(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load tasks report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Gamification(c *gin.Context) {
	item, err := h.service.Gamification(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load gamification report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Audit(c *gin.Context) {
	item, err := h.service.Audit(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load audit report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Management(c *gin.Context) {
	item, err := h.service.Management(c.Request.Context(), c.Param("kind"), filters(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Report not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) ExportManagement(c *gin.Context) {
	item, err := h.service.Management(c.Request.Context(), c.Param("kind"), filters(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Report not found"})
		return
	}
	format := c.Param("format")
	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="puls-%s-report.pdf"`, item.Code))
		c.Data(http.StatusOK, "application/pdf", renderReportPDF(item))
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="puls-%s-report.xlsx"`, item.Code))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", renderReportXLSX(item))
}

func filters(c *gin.Context) Filters {
	return Filters{From: c.Query("from"), To: c.Query("to")}
}

func renderReportCSV(report ManagementReport) string {
	var buffer bytes.Buffer
	_, _ = buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"section", "label", "group", "value", "details"})
	for _, metric := range report.Metrics {
		_ = writer.Write([]string{"metric", metric.Label, metric.Code, fmt.Sprintf("%.2f", metric.Value), ""})
	}
	for _, row := range report.Rows {
		_ = writer.Write([]string{"row", row.Label, row.Group, fmt.Sprintf("%.2f", row.Value), row.Details})
	}
	for _, row := range report.Risks {
		_ = writer.Write([]string{"risk", row.Label, row.Group, fmt.Sprintf("%.2f", row.Value), row.Details})
	}
	writer.Flush()
	return buffer.String()
}

func renderReportPDF(report ManagementReport) []byte {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddUTF8FontFromBytes("DejaVu", "", analyticsRegularFont)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", analyticsBoldFont)
	pdf.AddPage()

	pdf.SetFillColor(47, 160, 111)
	pdf.Rect(0, 0, 210, 34, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("DejaVu", "B", 18)
	pdf.SetXY(16, 11)
	pdf.CellFormat(178, 8, report.Title, "", 1, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 9)
	pdf.SetX(16)
	pdf.CellFormat(178, 5, "Управленческий отчет системы Пульс", "", 1, "L", false, 0, "")

	pdf.SetY(46)
	pdf.SetTextColor(22, 33, 58)
	pdf.SetFont("DejaVu", "B", 13)
	pdf.CellFormat(0, 7, "Ключевые показатели", "", 1, "L", false, 0, "")
	for _, metric := range report.Metrics {
		drawReportRow(pdf, metric.Label, metric.Code, metric.Value, "")
	}
	drawReportSection(pdf, "Строки отчета", report.Rows)
	drawReportSection(pdf, "Риски", report.Risks)

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil
	}
	return buffer.Bytes()
}

func drawReportSection(pdf *gofpdf.Fpdf, title string, rows []ReportRow) {
	if len(rows) == 0 {
		return
	}
	pdf.Ln(5)
	pdf.SetFont("DejaVu", "B", 13)
	pdf.SetTextColor(22, 33, 58)
	pdf.CellFormat(0, 7, title, "", 1, "L", false, 0, "")
	for _, row := range rows {
		drawReportRow(pdf, row.Label, row.Group, row.Value, row.Details)
	}
}

func drawReportRow(pdf *gofpdf.Fpdf, label string, group string, value float64, details string) {
	y := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.RoundedRect(16, y, 178, 17, 2, "1234", "FD")
	pdf.SetXY(20, y+3)
	pdf.SetFont("DejaVu", "B", 9)
	pdf.SetTextColor(22, 33, 58)
	pdf.CellFormat(82, 5, label, "", 0, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 8)
	pdf.SetTextColor(100, 112, 138)
	pdf.CellFormat(50, 5, group, "", 0, "L", false, 0, "")
	pdf.SetFont("DejaVu", "B", 10)
	pdf.SetTextColor(47, 160, 111)
	pdf.CellFormat(36, 5, fmt.Sprintf("%.2f", value), "", 1, "R", false, 0, "")
	if strings.TrimSpace(details) != "" {
		pdf.SetX(20)
		pdf.SetFont("DejaVu", "", 7.5)
		pdf.SetTextColor(100, 112, 138)
		pdf.CellFormat(164, 4, details, "", 1, "L", false, 0, "")
	}
	pdf.SetY(y + 20)
}

func renderReportXLSX(report ManagementReport) []byte {
	rows := [][]string{{"Раздел", "Показатель", "Группа", "Значение", "Описание"}}
	for _, metric := range report.Metrics {
		rows = append(rows, []string{"Показатель", metric.Label, metric.Code, fmt.Sprintf("%.2f", metric.Value), ""})
	}
	for _, row := range report.Rows {
		rows = append(rows, []string{"Строка", row.Label, row.Group, fmt.Sprintf("%.2f", row.Value), row.Details})
	}
	for _, row := range report.Risks {
		rows = append(rows, []string{"Риск", row.Label, row.Group, fmt.Sprintf("%.2f", row.Value), row.Details})
	}
	var buffer bytes.Buffer
	zw := zip.NewWriter(&buffer)
	writeZip(zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`)
	writeZip(zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`)
	writeZip(zw, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Отчет" sheetId="1" r:id="rId1"/></sheets></workbook>`)
	writeZip(zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`)
	writeZip(zw, "xl/worksheets/sheet1.xml", worksheetXML(rows))
	_ = zw.Close()
	return buffer.Bytes()
}

func worksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for r, row := range rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, r+1))
		for c, value := range row {
			cell := fmt.Sprintf("%c%d", rune('A'+c), r+1)
			builder.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, cell, xmlEscape(value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func xmlEscape(value string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}

func writeZip(zw *zip.Writer, name string, content string) {
	file, _ := zw.Create(name)
	_, _ = file.Write([]byte(content))
}
