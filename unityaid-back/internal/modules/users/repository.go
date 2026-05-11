package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")
var ErrSkillNotFound = errors.New("skill not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListUsers(ctx context.Context, filters UserFilters) ([]User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT u.id::text, u.email, u.first_name, u.last_name, u.patronymic, u.avatar_url,
			u.locale, u.is_email_verified, u.is_active, u.last_login_at, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN organization_members om ON om.user_id = u.id
		WHERE (
				$1 = ''
				OR u.email ILIKE '%' || $1 || '%'
				OR u.first_name ILIKE '%' || $1 || '%'
				OR u.last_name ILIKE '%' || $1 || '%'
			)
			AND ($2 = '' OR om.role::text = $2)
			AND ($3 = '' OR om.organization_id::text = $3)
		ORDER BY u.last_name, u.first_name
	`, filters.Search, filters.Role, filters.OrganizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []User{}
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		if err := r.loadUserMemberships(ctx, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (User, error) {
	item, err := scanUser(r.db.QueryRow(ctx, `
		SELECT id::text, email, first_name, last_name, patronymic, avatar_url,
			locale, is_email_verified, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if err := r.loadUserMemberships(ctx, &item); err != nil {
		return User{}, err
	}
	return item, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id string, request UpdateUserRequest) (User, error) {
	isActiveSQL := "is_active"
	args := []any{id, request.FirstName, request.LastName, request.Patronymic, request.AvatarURL, request.Locale}
	if request.IsActive != nil {
		isActiveSQL = "$7"
		args = append(args, *request.IsActive)
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE users
		SET first_name = $2,
			last_name = $3,
			patronymic = $4,
			avatar_url = $5,
			locale = $6,
			is_active = `+isActiveSQL+`,
			updated_at = now()
		WHERE id = $1
	`, args...)
	if err != nil {
		return User{}, err
	}
	if tag.RowsAffected() == 0 {
		return User{}, ErrNotFound
	}
	return r.FindUserByID(ctx, id)
}

func (r *Repository) ListVolunteers(ctx context.Context, filters VolunteerFilters) ([]VolunteerProfile, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT vp.id::text, u.id::text, u.email, u.first_name, u.last_name, u.patronymic, u.avatar_url,
			vp.city, vp.phone, vp.bio, vp.total_hours::float8, vp.points, vp.level, vp.created_at, vp.updated_at
		FROM volunteer_profiles vp
		JOIN users u ON u.id = vp.user_id
		LEFT JOIN volunteer_skills vs ON vs.volunteer_profile_id = vp.id
		LEFT JOIN organization_members om ON om.user_id = u.id
		WHERE (
				$1 = ''
				OR u.email ILIKE '%' || $1 || '%'
				OR u.first_name ILIKE '%' || $1 || '%'
				OR u.last_name ILIKE '%' || $1 || '%'
				OR vp.city ILIKE '%' || $1 || '%'
				OR vp.bio ILIKE '%' || $1 || '%'
			)
			AND ($2 = '' OR vs.skill_id::text = $2)
			AND ($3 = '' OR om.organization_id::text = $3)
		ORDER BY u.last_name, u.first_name
	`, filters.Search, filters.SkillID, filters.OrganizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []VolunteerProfile{}
	for rows.Next() {
		item, err := scanVolunteer(rows)
		if err != nil {
			return nil, err
		}
		if err := r.enrichVolunteer(ctx, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindVolunteerByUserID(ctx context.Context, userID string) (VolunteerProfile, error) {
	item, err := scanVolunteer(r.db.QueryRow(ctx, `
		SELECT vp.id::text, u.id::text, u.email, u.first_name, u.last_name, u.patronymic, u.avatar_url,
			vp.city, vp.phone, vp.bio, vp.total_hours::float8, vp.points, vp.level, vp.created_at, vp.updated_at
		FROM volunteer_profiles vp
		JOIN users u ON u.id = vp.user_id
		WHERE u.id = $1
	`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VolunteerProfile{}, ErrNotFound
	}
	if err != nil {
		return VolunteerProfile{}, err
	}
	if err := r.enrichVolunteer(ctx, &item); err != nil {
		return VolunteerProfile{}, err
	}
	return item, nil
}

func (r *Repository) UpdateVolunteer(ctx context.Context, userID string, request UpdateVolunteerRequest) (VolunteerProfile, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return VolunteerProfile{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE users
		SET first_name = $2, last_name = $3, patronymic = $4, avatar_url = $5, updated_at = now()
		WHERE id = $1
	`, userID, request.FirstName, request.LastName, request.Patronymic, request.AvatarURL)
	if err != nil {
		return VolunteerProfile{}, err
	}
	if tag.RowsAffected() == 0 {
		return VolunteerProfile{}, ErrNotFound
	}

	var profileID string
	err = tx.QueryRow(ctx, `
		INSERT INTO volunteer_profiles (user_id, city, phone, bio)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET city = EXCLUDED.city, phone = EXCLUDED.phone, bio = EXCLUDED.bio, updated_at = now()
		RETURNING id::text
	`, userID, request.City, request.Phone, request.Bio).Scan(&profileID)
	if err != nil {
		return VolunteerProfile{}, err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM volunteer_skills WHERE volunteer_profile_id = $1", profileID); err != nil {
		return VolunteerProfile{}, err
	}
	for _, skillID := range request.SkillIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO volunteer_skills (volunteer_profile_id, skill_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, profileID, skillID); err != nil {
			return VolunteerProfile{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return VolunteerProfile{}, err
	}
	return r.FindVolunteerByUserID(ctx, userID)
}

func (r *Repository) ListSkills(ctx context.Context) ([]Skill, error) {
	rows, err := r.db.Query(ctx, "SELECT id::text, name, created_at FROM skills ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Skill{}
	for rows.Next() {
		var item Skill
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateSkill(ctx context.Context, request CreateSkillRequest) (Skill, error) {
	var item Skill
	err := r.db.QueryRow(ctx, `
		INSERT INTO skills (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id::text, name, created_at
	`, request.Name).Scan(&item.ID, &item.Name, &item.CreatedAt)
	return item, err
}

func (r *Repository) DeleteSkill(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM skills WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSkillNotFound
	}
	return nil
}

func (r *Repository) enrichVolunteer(ctx context.Context, item *VolunteerProfile) error {
	skills, err := r.loadVolunteerSkills(ctx, item.ID)
	if err != nil {
		return err
	}
	memberships, err := r.loadMemberships(ctx, item.UserID)
	if err != nil {
		return err
	}
	item.Skills = skills
	item.Organizations = memberships
	return nil
}

func (r *Repository) loadVolunteerSkills(ctx context.Context, profileID string) ([]Skill, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id::text, s.name, s.created_at
		FROM volunteer_skills vs
		JOIN skills s ON s.id = vs.skill_id
		WHERE vs.volunteer_profile_id = $1
		ORDER BY s.name
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Skill{}
	for rows.Next() {
		var item Skill
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadUserMemberships(ctx context.Context, user *User) error {
	memberships, err := r.loadMemberships(ctx, user.ID)
	if err != nil {
		return err
	}
	user.Organizations = memberships
	return nil
}

func (r *Repository) loadMemberships(ctx context.Context, userID string) ([]Membership, error) {
	rows, err := r.db.Query(ctx, `
		SELECT om.id::text, om.organization_id::text, o.name, om.role::text, om.status::text
		FROM organization_members om
		JOIN organizations o ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY
			CASE om.role
				WHEN 'super_admin' THEN 1
				WHEN 'org_admin' THEN 2
				WHEN 'coordinator' THEN 3
				ELSE 4
			END,
			o.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Membership{}
	for rows.Next() {
		var item Membership
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.OrganizationName, &item.Role, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var item User
	var patronymic, avatarURL sql.NullString
	err := row.Scan(
		&item.ID,
		&item.Email,
		&item.FirstName,
		&item.LastName,
		&patronymic,
		&avatarURL,
		&item.Locale,
		&item.IsEmailVerified,
		&item.IsActive,
		&item.LastLoginAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if patronymic.Valid {
		item.Patronymic = &patronymic.String
	}
	if avatarURL.Valid {
		item.AvatarURL = &avatarURL.String
	}
	return item, err
}

func scanVolunteer(row scanner) (VolunteerProfile, error) {
	var item VolunteerProfile
	var patronymic, avatarURL, city, phone sql.NullString
	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Email,
		&item.FirstName,
		&item.LastName,
		&patronymic,
		&avatarURL,
		&city,
		&phone,
		&item.Bio,
		&item.TotalHours,
		&item.Points,
		&item.Level,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if patronymic.Valid {
		item.Patronymic = &patronymic.String
	}
	if avatarURL.Valid {
		item.AvatarURL = &avatarURL.String
	}
	if city.Valid {
		item.City = &city.String
	}
	if phone.Valid {
		item.Phone = &phone.String
	}
	return item, err
}
