package dataexports

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

type exportDefinition struct {
	filename string
	query    string
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

var definitions = map[string]exportDefinition{
	"volunteers": {
		filename: "unityaid-volunteers.csv",
		query: `
			SELECT u.id::text, u.email, u.last_name, u.first_name, COALESCE(u.patronymic, '') AS patronymic,
				COALESCE(vp.status::text, '') AS status, COALESCE(vp.city, '') AS city, COALESCE(vp.phone, '') AS phone,
				COALESCE(vp.interests, '') AS interests, COALESCE(vp.total_hours, 0) AS total_hours,
				COALESCE(vp.points, 0) AS points, COALESCE(vp.level, 1) AS level, u.created_at
			FROM users u
			LEFT JOIN volunteer_profiles vp ON vp.user_id = u.id
			ORDER BY u.last_name, u.first_name, u.email
		`,
	},
	"events": {
		filename: "unityaid-events.csv",
		query: `
			SELECT e.id::text, o.name AS organization, e.title, e.description, e.format::text, e.status::text,
				e.starts_at, e.ends_at, COALESCE(e.location, '') AS location, COALESCE(e.max_participants, 0) AS max_participants,
				e.created_at
			FROM events e
			JOIN organizations o ON o.id = e.organization_id
			ORDER BY e.starts_at DESC
		`,
	},
	"applications": {
		filename: "unityaid-applications.csv",
		query: `
			SELECT ea.id::text, e.title AS event, o.name AS organization, u.email, u.last_name, u.first_name,
				ea.status::text, ea.message, ea.created_at, ea.updated_at
			FROM event_applications ea
			JOIN events e ON e.id = ea.event_id
			JOIN organizations o ON o.id = e.organization_id
			JOIN users u ON u.id = ea.user_id
			ORDER BY ea.created_at DESC
		`,
	},
	"time-entries": {
		filename: "unityaid-time-entries.csv",
		query: `
			SELECT te.id::text, o.name AS organization, u.email, u.last_name, u.first_name,
				COALESCE(e.title, '') AS event, COALESCE(t.title, '') AS task, te.hours, te.status,
				te.description, te.reviewed_at, te.created_at
			FROM time_entries te
			JOIN organizations o ON o.id = te.organization_id
			JOIN users u ON u.id = te.user_id
			LEFT JOIN events e ON e.id = te.event_id
			LEFT JOIN tasks t ON t.id = te.task_id
			ORDER BY te.created_at DESC
		`,
	},
	"tasks": {
		filename: "unityaid-tasks.csv",
		query: `
			SELECT t.id::text, o.name AS organization, COALESCE(e.title, '') AS event, t.title, t.description,
				t.status::text, t.priority, t.due_at, t.completion_confirmed_at, t.created_at
			FROM tasks t
			JOIN organizations o ON o.id = t.organization_id
			LEFT JOIN events e ON e.id = t.event_id
			ORDER BY t.created_at DESC
		`,
	},
	"certificates": {
		filename: "unityaid-certificates.csv",
		query: `
			SELECT c.id::text, u.email, u.last_name, u.first_name, COALESCE(o.name, '') AS organization,
				c.type, c.title, c.description, c.total_hours, c.verify_code, c.issued_at, c.created_at
			FROM certificates c
			JOIN users u ON u.id = c.user_id
			LEFT JOIN organizations o ON o.id = c.organization_id
			ORDER BY c.issued_at DESC
		`,
	},
}

func (h *Handler) Download(c *gin.Context) {
	kind := c.Param("kind")
	definition, ok := definitions[kind]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Export type not found"})
		return
	}

	rows, err := h.db.Query(c.Request.Context(), definition.query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not build export"})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, definition.filename))

	writer := csv.NewWriter(c.Writer)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	fields := rows.FieldDescriptions()
	headers := make([]string, 0, len(fields))
	for _, field := range fields {
		headers = append(headers, string(field.Name))
	}
	if err := writer.Write(headers); err != nil {
		return
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return
		}
		record := make([]string, 0, len(values))
		for _, value := range values {
			record = append(record, formatCSVValue(value))
		}
		if err := writer.Write(record); err != nil {
			return
		}
	}
	writer.Flush()
}

func formatCSVValue(value any) string {
	switch item := value.(type) {
	case nil:
		return ""
	case time.Time:
		return item.Format(time.RFC3339)
	case []byte:
		return string(item)
	default:
		return strings.TrimSpace(fmt.Sprint(item))
	}
}
