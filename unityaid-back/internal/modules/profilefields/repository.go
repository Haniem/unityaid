package profilefields

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("profile field not found")
var ErrProtected = errors.New("system profile field cannot be modified")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) EnsureSystemSchema(ctx context.Context, organizationID string) error {
	if _, err := r.db.Exec(ctx, `
		INSERT INTO profile_field_groups (organization_id, code, name, sort_order, is_system)
		SELECT $1, seed.code, seed.name, seed.sort_order, true
		FROM (VALUES ('personal', 'Личная информация', 10), ('contacts', 'Контакты', 20), ('work', 'Работа', 30), ('results', 'Результаты', 40)) AS seed(code, name, sort_order)
		ON CONFLICT (organization_id, code) DO NOTHING
	`, organizationID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO profile_field_definitions (organization_id, group_id, code, name, field_type, sort_order, is_system, editable_by_user)
		SELECT g.organization_id, g.id, seed.code, seed.name, seed.field_type, seed.sort_order, true, seed.editable_by_user
		FROM profile_field_groups g
		JOIN (VALUES
			('personal', 'firstName', 'Имя', 'text', 10, true),
			('personal', 'lastName', 'Фамилия', 'text', 20, true),
			('personal', 'patronymic', 'Отчество', 'text', 30, true),
			('contacts', 'email', 'Электронная почта', 'email', 10, false),
			('contacts', 'phone', 'Телефон', 'tel', 20, true),
			('work', 'city', 'Город', 'text', 10, true),
			('work', 'bio', 'Опыт и описание', 'textarea', 20, true),
			('work', 'interests', 'Интересы', 'textarea', 30, true),
			('results', 'totalHours', 'Подтвержденные часы', 'number', 10, false),
			('results', 'points', 'Баллы', 'number', 20, false)
		) AS seed(group_code, code, name, field_type, sort_order, editable_by_user) ON seed.group_code = g.code
		WHERE g.organization_id = $1 AND g.is_system = true
		ON CONFLICT (organization_id, code) DO NOTHING
	`, organizationID)
	return err
}

func (r *Repository) ListSchema(ctx context.Context, organizationID string) ([]Group, error) {
	if err := r.EnsureSystemSchema(ctx, organizationID); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, organization_id::text, code, name, description, sort_order, is_system, is_active
		FROM profile_field_groups WHERE organization_id = $1 ORDER BY sort_order, name
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Group{}
	for rows.Next() {
		var item Group
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Description, &item.SortOrder, &item.IsSystem, &item.IsActive); err != nil {
			return nil, err
		}
		item.Fields, err = r.listFields(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) listFields(ctx context.Context, groupID string) ([]Field, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, organization_id::text, group_id::text, code, name, field_type, required, is_system,
			is_active, editable_by_user, sort_order, placeholder, help
		FROM profile_field_definitions WHERE group_id = $1 ORDER BY sort_order, name
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Field{}
	for rows.Next() {
		var item Field
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.GroupID, &item.Code, &item.Name, &item.Type, &item.Required, &item.IsSystem, &item.IsActive, &item.EditableByUser, &item.SortOrder, &item.Placeholder, &item.Help); err != nil {
			return nil, err
		}
		item.Options, err = r.listOptions(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) listOptions(ctx context.Context, fieldID string) ([]Option, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, value, label, sort_order, is_active FROM profile_field_options WHERE field_id = $1 ORDER BY sort_order, label`, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Option{}
	for rows.Next() {
		var item Option
		if err := rows.Scan(&item.ID, &item.Value, &item.Label, &item.SortOrder, &item.IsActive); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateGroup(ctx context.Context, request GroupRequest) (Group, error) {
	code := "custom_" + strings.ReplaceAll(strings.ToLower(strings.TrimSpace(request.Name)), " ", "_")
	var item Group
	err := r.db.QueryRow(ctx, `
		INSERT INTO profile_field_groups (organization_id, code, name, description, sort_order, is_active)
		VALUES ($1, $2 || '_' || substr(gen_random_uuid()::text, 1, 8), $3, $4, $5, $6)
		RETURNING id::text, organization_id::text, code, name, description, sort_order, is_system, is_active
	`, request.OrganizationID, code, request.Name, request.Description, request.SortOrder, request.IsActive).
		Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Description, &item.SortOrder, &item.IsSystem, &item.IsActive)
	item.Fields = []Field{}
	return item, err
}

func (r *Repository) UpdateGroup(ctx context.Context, id string, request GroupRequest) (Group, error) {
	var protected bool
	if err := r.db.QueryRow(ctx, `SELECT is_system FROM profile_field_groups WHERE id = $1 AND organization_id = $2`, id, request.OrganizationID).Scan(&protected); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Group{}, ErrNotFound
		}
		return Group{}, err
	}
	var item Group
	err := r.db.QueryRow(ctx, `
		UPDATE profile_field_groups SET name = $3, description = $4, sort_order = $5, is_active = CASE WHEN is_system THEN true ELSE $6 END, updated_at = now()
		WHERE id = $1 AND organization_id = $2
		RETURNING id::text, organization_id::text, code, name, description, sort_order, is_system, is_active
	`, id, request.OrganizationID, request.Name, request.Description, request.SortOrder, request.IsActive).
		Scan(&item.ID, &item.OrganizationID, &item.Code, &item.Name, &item.Description, &item.SortOrder, &item.IsSystem, &item.IsActive)
	if err == nil {
		item.Fields, err = r.listFields(ctx, item.ID)
	}
	return item, err
}

func (r *Repository) DeleteGroup(ctx context.Context, id string, organizationID string) error {
	var system bool
	if err := r.db.QueryRow(ctx, `SELECT is_system FROM profile_field_groups WHERE id = $1 AND organization_id = $2`, id, organizationID).Scan(&system); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if system {
		return ErrProtected
	}
	_, err := r.db.Exec(ctx, `DELETE FROM profile_field_groups WHERE id = $1 AND organization_id = $2`, id, organizationID)
	return err
}

func (r *Repository) CreateField(ctx context.Context, request FieldRequest) (Field, error) {
	code := fmt.Sprintf("custom_%s", strings.ReplaceAll(strings.ToLower(strings.TrimSpace(request.Name)), " ", "_"))
	var item Field
	err := r.db.QueryRow(ctx, `
		INSERT INTO profile_field_definitions (organization_id, group_id, code, name, field_type, required, is_active, editable_by_user, sort_order, placeholder, help)
		VALUES ($1, $2, $3 || '_' || substr(gen_random_uuid()::text, 1, 8), $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id::text, organization_id::text, group_id::text, code, name, field_type, required, is_system, is_active, editable_by_user, sort_order, placeholder, help
	`, request.OrganizationID, request.GroupID, code, request.Name, request.Type, request.Required, request.IsActive, request.EditableByUser, request.SortOrder, request.Placeholder, request.Help).
		Scan(&item.ID, &item.OrganizationID, &item.GroupID, &item.Code, &item.Name, &item.Type, &item.Required, &item.IsSystem, &item.IsActive, &item.EditableByUser, &item.SortOrder, &item.Placeholder, &item.Help)
	if err == nil {
		err = r.replaceOptions(ctx, item.ID, request.Options)
		item.Options = request.Options
	}
	return item, err
}

func (r *Repository) UpdateField(ctx context.Context, id string, request FieldRequest) (Field, error) {
	var system bool
	if err := r.db.QueryRow(ctx, `SELECT is_system FROM profile_field_definitions WHERE id = $1 AND organization_id = $2`, id, request.OrganizationID).Scan(&system); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Field{}, ErrNotFound
		}
		return Field{}, err
	}
	var item Field
	err := r.db.QueryRow(ctx, `
		UPDATE profile_field_definitions SET group_id = CASE WHEN is_system THEN group_id ELSE $3 END,
			name = CASE WHEN is_system THEN name ELSE $4 END, field_type = CASE WHEN is_system THEN field_type ELSE $5 END,
			required = CASE WHEN is_system THEN required ELSE $6 END, is_active = CASE WHEN is_system THEN true ELSE $7 END,
			editable_by_user = CASE WHEN is_system THEN editable_by_user ELSE $8 END, sort_order = $9,
			placeholder = $10, help = $11, updated_at = now()
		WHERE id = $1 AND organization_id = $2
		RETURNING id::text, organization_id::text, group_id::text, code, name, field_type, required, is_system, is_active, editable_by_user, sort_order, placeholder, help
	`, id, request.OrganizationID, request.GroupID, request.Name, request.Type, request.Required, request.IsActive, request.EditableByUser, request.SortOrder, request.Placeholder, request.Help).
		Scan(&item.ID, &item.OrganizationID, &item.GroupID, &item.Code, &item.Name, &item.Type, &item.Required, &item.IsSystem, &item.IsActive, &item.EditableByUser, &item.SortOrder, &item.Placeholder, &item.Help)
	if err == nil && !system {
		err = r.replaceOptions(ctx, item.ID, request.Options)
	}
	if err == nil {
		item.Options, err = r.listOptions(ctx, item.ID)
	}
	return item, err
}

func (r *Repository) DeleteField(ctx context.Context, id string, organizationID string) error {
	var system bool
	if err := r.db.QueryRow(ctx, `SELECT is_system FROM profile_field_definitions WHERE id = $1 AND organization_id = $2`, id, organizationID).Scan(&system); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if system {
		return ErrProtected
	}
	_, err := r.db.Exec(ctx, `DELETE FROM profile_field_definitions WHERE id = $1 AND organization_id = $2`, id, organizationID)
	return err
}

func (r *Repository) replaceOptions(ctx context.Context, fieldID string, options []Option) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM profile_field_options WHERE field_id = $1`, fieldID); err != nil {
		return err
	}
	for index, option := range options {
		if strings.TrimSpace(option.Value) == "" {
			continue
		}
		if _, err := r.db.Exec(ctx, `INSERT INTO profile_field_options (field_id, value, label, sort_order, is_active) VALUES ($1, $2, $3, $4, true)`, fieldID, strings.TrimSpace(option.Value), strings.TrimSpace(option.Label), index*10); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Values(ctx context.Context, organizationID string, userID string) (map[string]json.RawMessage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.code, v.value FROM profile_field_values v
		JOIN profile_field_definitions f ON f.id = v.field_id
		WHERE v.organization_id = $1 AND v.user_id = $2
	`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := map[string]json.RawMessage{}
	for rows.Next() {
		var code string
		var value []byte
		if err := rows.Scan(&code, &value); err != nil {
			return nil, err
		}
		values[code] = json.RawMessage(value)
	}
	return values, rows.Err()
}

func (r *Repository) SaveValues(ctx context.Context, organizationID string, userID string, updatedBy string, values map[string]json.RawMessage) error {
	for code, value := range values {
		if !json.Valid(value) {
			return errors.New("invalid value")
		}
		_, err := r.db.Exec(ctx, `
			INSERT INTO profile_field_values (organization_id, user_id, field_id, value, updated_by)
			SELECT $1::uuid, $2::uuid, f.id, $4::jsonb, $3::uuid
			FROM profile_field_definitions f
			WHERE f.organization_id = $1::uuid AND f.code = $5 AND f.is_active = true AND (f.editable_by_user = true OR $2::uuid <> $3::uuid)
			ON CONFLICT (organization_id, user_id, field_id)
			DO UPDATE SET value = EXCLUDED.value, updated_by = EXCLUDED.updated_by, updated_at = now()
		`, organizationID, userID, updatedBy, string(value), code)
		if err != nil {
			return err
		}
	}
	return nil
}
