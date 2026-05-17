package analytics

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="unityaid-%s-report.pdf"`, item.Code))
		c.String(http.StatusOK, renderReportText(item))
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="unityaid-%s-report.xlsx"`, item.Code))
	c.String(http.StatusOK, renderReportCSV(item))
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

func renderReportText(report ManagementReport) string {
	lines := []string{report.Title, ""}
	for _, metric := range report.Metrics {
		lines = append(lines, fmt.Sprintf("%s: %.2f", metric.Label, metric.Value))
	}
	lines = append(lines, "", "Rows:")
	for _, row := range report.Rows {
		lines = append(lines, fmt.Sprintf("- %s / %s: %.2f %s", row.Group, row.Label, row.Value, row.Details))
	}
	lines = append(lines, "", "Risks:")
	for _, row := range report.Risks {
		lines = append(lines, fmt.Sprintf("- %s / %s: %.2f %s", row.Group, row.Label, row.Value, row.Details))
	}
	return strings.Join(lines, "\n")
}
