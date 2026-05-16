package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrUnknownEntity = errors.New("unknown admin entity")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Entities() []EntityConfig {
	items := make([]EntityConfig, 0, len(entityConfigs))
	for _, item := range entityConfigs {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

func (r *Repository) List(ctx context.Context, code string, page int, limit int, search string) (ListResponse, error) {
	config, ok := entityConfigs[code]
	if !ok {
		return ListResponse{}, ErrUnknownEntity
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	where, args := searchWhere(config, search)
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM "+config.Table+where, args...).Scan(&total); err != nil {
		return ListResponse{}, err
	}
	args = append(args, limit, (page-1)*limit)
	rows, err := r.db.Query(ctx, "SELECT row_to_json(t)::text FROM (SELECT "+strings.Join(config.Columns, ", ")+" FROM "+config.Table+where+" ORDER BY "+config.OrderBy+" LIMIT $"+fmt.Sprint(len(args)-1)+" OFFSET $"+fmt.Sprint(len(args))+") t", args...)
	if err != nil {
		return ListResponse{}, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return ListResponse{}, err
		}
		item := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return ListResponse{}, err
		}
		items = append(items, item)
	}
	return ListResponse{Items: items, Total: total, Page: page, Limit: limit}, rows.Err()
}

func (r *Repository) Create(ctx context.Context, code string, data map[string]any) (map[string]any, error) {
	config, ok := entityConfigs[code]
	if !ok {
		return nil, ErrUnknownEntity
	}
	if !config.CanCreate {
		return nil, errors.New("entity cannot be created")
	}
	if code == "users" {
		if _, ok := data["password_hash"]; !ok {
			hash, _ := bcrypt.GenerateFromPassword([]byte("ChangeMe123!"), bcrypt.DefaultCost)
			data["password_hash"] = string(hash)
		}
		if _, ok := data["is_email_verified"]; !ok {
			data["is_email_verified"] = true
		}
	}
	columns, values := editableData(config, data)
	if len(columns) == 0 {
		return nil, errors.New("empty payload")
	}
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "$" + fmt.Sprint(i+1)
	}
	query := "INSERT INTO " + config.Table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ") RETURNING " + config.PrimaryKey + "::text"
	var id string
	if err := r.db.QueryRow(ctx, query, values...).Scan(&id); err != nil {
		return nil, err
	}
	return r.Get(ctx, code, id)
}

func (r *Repository) Get(ctx context.Context, code string, id string) (map[string]any, error) {
	config, ok := entityConfigs[code]
	if !ok {
		return nil, ErrUnknownEntity
	}
	var raw string
	err := r.db.QueryRow(ctx, "SELECT row_to_json(t)::text FROM (SELECT "+strings.Join(config.Columns, ", ")+" FROM "+config.Table+" WHERE "+config.PrimaryKey+" = $1) t", id).Scan(&raw)
	if err != nil {
		return nil, err
	}
	item := map[string]any{}
	return item, json.Unmarshal([]byte(raw), &item)
}

func (r *Repository) Update(ctx context.Context, code string, id string, data map[string]any) (map[string]any, error) {
	config, ok := entityConfigs[code]
	if !ok {
		return nil, ErrUnknownEntity
	}
	columns, values := editableData(config, data)
	if len(columns) == 0 {
		return r.Get(ctx, code, id)
	}
	assignments := make([]string, len(columns))
	for i, column := range columns {
		assignments[i] = column + " = $" + fmt.Sprint(i+2)
	}
	if hasColumn(config.Editable, "updated_at") {
		assignments = append(assignments, "updated_at = now()")
	}
	args := append([]any{id}, values...)
	_, err := r.db.Exec(ctx, "UPDATE "+config.Table+" SET "+strings.Join(assignments, ", ")+" WHERE "+config.PrimaryKey+" = $1", args...)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, code, id)
}

func (r *Repository) Delete(ctx context.Context, code string, id string) error {
	config, ok := entityConfigs[code]
	if !ok {
		return ErrUnknownEntity
	}
	if !config.CanDelete {
		return errors.New("entity cannot be deleted")
	}
	_, err := r.db.Exec(ctx, "DELETE FROM "+config.Table+" WHERE "+config.PrimaryKey+" = $1", id)
	return err
}

func editableData(config EntityConfig, data map[string]any) ([]string, []any) {
	editable := map[string]bool{}
	for _, column := range config.Editable {
		editable[column] = true
	}
	columns := []string{}
	for column := range data {
		if editable[column] && column != "updated_at" {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)
	values := make([]any, len(columns))
	for i, column := range columns {
		values[i] = data[column]
	}
	return columns, values
}

func searchWhere(config EntityConfig, search string) (string, []any) {
	search = strings.TrimSpace(search)
	if search == "" || len(config.Search) == 0 {
		return "", []any{}
	}
	parts := make([]string, len(config.Search))
	for i, column := range config.Search {
		parts[i] = column + "::text ILIKE '%' || $1 || '%'"
	}
	return " WHERE (" + strings.Join(parts, " OR ") + ")", []any{search}
}

func hasColumn(columns []string, target string) bool {
	for _, column := range columns {
		if column == target {
			return true
		}
	}
	return false
}
