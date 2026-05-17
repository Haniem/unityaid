package volunteerimports

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *pgxpool.Pool
}

type ImportRow struct {
	Row       int      `json:"row"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	City      string   `json:"city"`
	Phone     string   `json:"phone"`
	Interests string   `json:"interests"`
	Status    string   `json:"status"`
	Errors    []string `json:"errors"`
}

type PreviewResponse struct {
	Items      []ImportRow `json:"items"`
	ValidCount int         `json:"validCount"`
	ErrorCount int         `json:"errorCount"`
}

type CommitRequest struct {
	Items []ImportRow `json:"items"`
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Preview(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "File is required"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not read file"})
		return
	}

	records, err := parseTable(content, header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	response := buildPreview(records)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Commit(c *gin.Context) {
	var request CommitRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	imported, err := h.importRows(c.Request.Context(), request.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not import volunteers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": imported})
}

func (h *Handler) importRows(ctx context.Context, rows []ImportRow) (int, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	imported := 0
	for _, row := range rows {
		row = normalizeRow(row)
		row.Errors = validateRow(row)
		if len(row.Errors) > 0 {
			continue
		}
		tx, err := h.db.Begin(ctx)
		if err != nil {
			return imported, err
		}
		var userID string
		err = tx.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, first_name, last_name, is_email_verified, is_active)
			VALUES ($1, $2, $3, $4, true, true)
			ON CONFLICT (email)
			DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, updated_at = now()
			RETURNING id::text
		`, row.Email, string(passwordHash), row.FirstName, row.LastName).Scan(&userID)
		if err != nil {
			_ = tx.Rollback(ctx)
			return imported, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO volunteer_profiles (user_id, city, phone, bio, status, interests)
			VALUES ($1, $2, $3, '', $4::volunteer_status, $5)
			ON CONFLICT (user_id)
			DO UPDATE SET city = EXCLUDED.city, phone = EXCLUDED.phone, status = EXCLUDED.status, interests = EXCLUDED.interests, updated_at = now()
		`, userID, optional(row.City), optional(row.Phone), row.Status, row.Interests)
		if err != nil {
			_ = tx.Rollback(ctx)
			return imported, err
		}
		if err := tx.Commit(ctx); err != nil {
			return imported, err
		}
		imported++
	}
	return imported, nil
}

func buildPreview(records [][]string) PreviewResponse {
	response := PreviewResponse{Items: []ImportRow{}}
	if len(records) == 0 {
		return response
	}
	header := map[string]int{}
	for index, value := range records[0] {
		header[normalizeHeader(value)] = index
	}
	for index, record := range records[1:] {
		row := normalizeRow(ImportRow{
			Row:       index + 2,
			Email:     cell(record, header, "email"),
			FirstName: cell(record, header, "first_name"),
			LastName:  cell(record, header, "last_name"),
			City:      cell(record, header, "city"),
			Phone:     cell(record, header, "phone"),
			Interests: cell(record, header, "interests"),
			Status:    cell(record, header, "status"),
		})
		row.Errors = validateRow(row)
		if len(row.Errors) > 0 {
			response.ErrorCount++
		} else {
			response.ValidCount++
		}
		response.Items = append(response.Items, row)
	}
	return response
}

func parseTable(content []byte, filename string) ([][]string, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".csv":
		reader := csv.NewReader(bytes.NewReader(content))
		reader.FieldsPerRecord = -1
		return reader.ReadAll()
	case ".tsv":
		reader := csv.NewReader(bytes.NewReader(content))
		reader.Comma = '\t'
		reader.FieldsPerRecord = -1
		return reader.ReadAll()
	case ".xlsx":
		return parseXLSX(content)
	default:
		return nil, errors.New("Supported formats: CSV, TSV, XLSX")
	}
}

type sharedStringsXML struct {
	Items []struct {
		Text string `xml:"t"`
	} `xml:"si"`
}

type worksheetXML struct {
	Rows []struct {
		Cells []struct {
			Ref   string `xml:"r,attr"`
			Type  string `xml:"t,attr"`
			Value string `xml:"v"`
		} `xml:"c"`
	} `xml:"sheetData>row"`
}

func parseXLSX(content []byte) ([][]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, errors.New("Could not open XLSX file")
	}
	shared := []string{}
	var sheet []byte
	for _, file := range reader.File {
		switch file.Name {
		case "xl/sharedStrings.xml":
			data, _ := readZipFile(file)
			var parsed sharedStringsXML
			_ = xml.Unmarshal(data, &parsed)
			for _, item := range parsed.Items {
				shared = append(shared, item.Text)
			}
		case "xl/worksheets/sheet1.xml":
			sheet, _ = readZipFile(file)
		}
	}
	if len(sheet) == 0 {
		return nil, errors.New("Could not find first worksheet")
	}
	var parsed worksheetXML
	if err := xml.Unmarshal(sheet, &parsed); err != nil {
		return nil, errors.New("Could not parse worksheet")
	}
	table := [][]string{}
	for _, row := range parsed.Rows {
		values := []string{}
		for _, cell := range row.Cells {
			column := columnIndex(cell.Ref)
			for len(values) < column-1 {
				values = append(values, "")
			}
			value := cell.Value
			if cell.Type == "s" {
				index, _ := strconv.Atoi(value)
				if index >= 0 && index < len(shared) {
					value = shared[index]
				}
			}
			values = append(values, value)
		}
		table = append(table, values)
	}
	return table, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	handle, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	return io.ReadAll(handle)
}

func columnIndex(ref string) int {
	result := 0
	for _, char := range ref {
		if char < 'A' || char > 'Z' {
			break
		}
		result = result*26 + int(char-'A'+1)
	}
	if result == 0 {
		return 1
	}
	return result
}

func cell(record []string, header map[string]int, key string) string {
	index, ok := header[key]
	if !ok || index >= len(record) {
		return ""
	}
	return record[index]
}

func normalizeHeader(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func normalizeRow(row ImportRow) ImportRow {
	row.Email = strings.ToLower(strings.TrimSpace(row.Email))
	row.FirstName = strings.TrimSpace(row.FirstName)
	row.LastName = strings.TrimSpace(row.LastName)
	row.City = strings.TrimSpace(row.City)
	row.Phone = strings.TrimSpace(row.Phone)
	row.Interests = strings.TrimSpace(row.Interests)
	row.Status = strings.TrimSpace(row.Status)
	if row.Status == "" {
		row.Status = "new"
	}
	return row
}

func validateRow(row ImportRow) []string {
	errors := []string{}
	if !strings.Contains(row.Email, "@") {
		errors = append(errors, "Некорректный email")
	}
	if row.FirstName == "" {
		errors = append(errors, "Не указано имя")
	}
	if row.LastName == "" {
		errors = append(errors, "Не указана фамилия")
	}
	if !contains([]string{"new", "active", "unavailable", "archived"}, row.Status) {
		errors = append(errors, "Некорректный статус")
	}
	return errors
}

func optional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
