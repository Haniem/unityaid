package gamification

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAchievements(ctx context.Context) ([]Achievement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, code, name, description, icon, points_reward, created_at
		FROM achievements
		ORDER BY points_reward, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Achievement{}
	for rows.Next() {
		item, err := scanAchievement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Profile(ctx context.Context, userID string) (Profile, error) {
	if err := r.RecalculateUser(ctx, userID); err != nil {
		return Profile{}, err
	}
	var item Profile
	err := r.db.QueryRow(ctx, `
		SELECT u.id::text, concat_ws(' ', u.last_name, u.first_name), u.email, u.avatar_url,
			vp.total_hours::float8, vp.points, vp.level
		FROM users u
		JOIN volunteer_profiles vp ON vp.user_id = u.id
		WHERE u.id = $1
	`, userID).Scan(&item.UserID, &item.UserName, &item.Email, &item.AvatarURL, &item.TotalHours, &item.Points, &item.Level)
	if err != nil {
		return Profile{}, err
	}
	item.NextLevelAt = item.Level * 100
	item.LevelProgress = float64(item.Points%100) / 100
	achievements, err := r.UserAchievements(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	item.Achievements = achievements
	next, err := r.NextAchievement(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	item.NextAchievement = next
	return item, nil
}

func (r *Repository) UserAchievements(ctx context.Context, userID string) ([]UserAchievement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id::text, a.code, a.name, a.description, a.icon, a.points_reward, a.created_at, va.earned_at
		FROM volunteer_achievements va
		JOIN achievements a ON a.id = va.achievement_id
		WHERE va.user_id = $1
		ORDER BY va.earned_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserAchievement{}
	for rows.Next() {
		var item UserAchievement
		achievement, err := scanAchievementPrefix(rows, &item.EarnedAt)
		if err != nil {
			return nil, err
		}
		item.Achievement = achievement
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) NextAchievement(ctx context.Context, userID string) (*Achievement, error) {
	item, err := scanAchievement(r.db.QueryRow(ctx, `
		SELECT a.id::text, a.code, a.name, a.description, a.icon, a.points_reward, a.created_at
		FROM achievements a
		WHERE NOT EXISTS (
			SELECT 1 FROM volunteer_achievements va WHERE va.achievement_id = a.id AND va.user_id = $1
		)
		ORDER BY a.points_reward, a.name
		LIMIT 1
	`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Leaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT row_number() OVER (ORDER BY vp.points DESC, vp.total_hours DESC, u.last_name, u.first_name)::int,
			u.id::text, concat_ws(' ', u.last_name, u.first_name), u.email, u.avatar_url,
			vp.total_hours::float8, vp.points, vp.level
		FROM volunteer_profiles vp
		JOIN users u ON u.id = vp.user_id
		ORDER BY vp.points DESC, vp.total_hours DESC, u.last_name, u.first_name
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LeaderboardEntry{}
	for rows.Next() {
		var item LeaderboardEntry
		if err := rows.Scan(&item.Rank, &item.UserID, &item.UserName, &item.Email, &item.AvatarURL, &item.TotalHours, &item.Points, &item.Level); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) RecalculateAll(ctx context.Context) error {
	rows, err := r.db.Query(ctx, `SELECT id::text FROM users`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		if err := r.RecalculateUser(ctx, userID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *Repository) UserIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text FROM users WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (r *Repository) RecalculateUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `SELECT recalculate_user_gamification($1)`, userID)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAchievement(row scanner) (Achievement, error) {
	var item Achievement
	err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.Icon, &item.PointsReward, &item.CreatedAt)
	return item, err
}

func scanAchievementPrefix(row scanner, earnedAt *time.Time) (Achievement, error) {
	var item Achievement
	err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.Icon, &item.PointsReward, &item.CreatedAt, earnedAt)
	return item, err
}
