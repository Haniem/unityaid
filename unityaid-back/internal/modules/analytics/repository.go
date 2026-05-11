package analytics

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Overview(ctx context.Context, filters Filters) (OverviewReport, error) {
	metrics, err := r.metrics(ctx, filters)
	if err != nil {
		return OverviewReport{}, err
	}
	applications, err := r.series(ctx, `event_applications`, `created_at`, filters)
	if err != nil {
		return OverviewReport{}, err
	}
	attendance, err := r.series(ctx, `event_attendance`, `created_at`, filters)
	if err != nil {
		return OverviewReport{}, err
	}
	top, err := r.topVolunteers(ctx)
	if err != nil {
		return OverviewReport{}, err
	}
	return OverviewReport{Metrics: metrics, Applications: applications, Attendance: attendance, TopVolunteers: top}, nil
}

func (r *Repository) Volunteers(ctx context.Context, filters Filters) (VolunteersReport, error) {
	metrics, err := r.volunteerMetrics(ctx, filters)
	if err != nil {
		return VolunteersReport{}, err
	}
	status, err := r.group(ctx, `SELECT vp.status::text, COUNT(*)::float8 FROM volunteer_profiles vp GROUP BY vp.status::text ORDER BY 1`)
	if err != nil {
		return VolunteersReport{}, err
	}
	hoursByLevel, err := r.group(ctx, `SELECT CONCAT('Уровень ', level), SUM(total_hours)::float8 FROM volunteer_profiles GROUP BY level ORDER BY level`)
	if err != nil {
		return VolunteersReport{}, err
	}
	top, err := r.topVolunteers(ctx)
	if err != nil {
		return VolunteersReport{}, err
	}
	return VolunteersReport{Metrics: metrics, Status: status, HoursByLevel: hoursByLevel, TopVolunteers: top}, nil
}

func (r *Repository) Events(ctx context.Context, filters Filters) (EventsReport, error) {
	metrics, err := r.eventMetrics(ctx, filters)
	if err != nil {
		return EventsReport{}, err
	}
	byStatus, err := r.group(ctx, `SELECT status::text, COUNT(*)::float8 FROM events GROUP BY status::text ORDER BY 1`)
	if err != nil {
		return EventsReport{}, err
	}
	applications, err := r.series(ctx, `event_applications`, `created_at`, filters)
	if err != nil {
		return EventsReport{}, err
	}
	attendance, err := r.series(ctx, `event_attendance`, `created_at`, filters)
	if err != nil {
		return EventsReport{}, err
	}
	return EventsReport{Metrics: metrics, ByStatus: byStatus, Applications: applications, Attendance: attendance}, nil
}

func (r *Repository) Tasks(ctx context.Context, filters Filters) (TasksReport, error) {
	metrics, err := r.taskMetrics(ctx, filters)
	if err != nil {
		return TasksReport{}, err
	}
	byStatus, err := r.group(ctx, `SELECT status::text, COUNT(*)::float8 FROM tasks GROUP BY status::text ORDER BY 1`)
	if err != nil {
		return TasksReport{}, err
	}
	byPriority, err := r.group(ctx, `SELECT priority, COUNT(*)::float8 FROM tasks GROUP BY priority ORDER BY priority`)
	if err != nil {
		return TasksReport{}, err
	}
	completed, err := r.series(ctx, `tasks`, `completion_confirmed_at`, filters)
	if err != nil {
		return TasksReport{}, err
	}
	return TasksReport{Metrics: metrics, ByStatus: byStatus, ByPriority: byPriority, Completed: completed}, nil
}

func (r *Repository) Gamification(ctx context.Context, filters Filters) (GamificationReport, error) {
	metrics, err := r.gamificationMetrics(ctx, filters)
	if err != nil {
		return GamificationReport{}, err
	}
	byReason, err := r.groupWithFilters(ctx, `
		SELECT COALESCE(NULLIF(a.name, ''), NULLIF(pt.reason, ''), pt.source_type), SUM(pt.points)::float8
		FROM points_transactions pt
		LEFT JOIN achievements a ON a.id = pt.achievement_id
		WHERE pt.created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
		GROUP BY COALESCE(NULLIF(a.name, ''), NULLIF(pt.reason, ''), pt.source_type)
		ORDER BY SUM(pt.points) DESC
		LIMIT 10
	`, filters)
	if err != nil {
		return GamificationReport{}, err
	}
	pointsByDay, err := r.seriesSum(ctx, `points_transactions`, `created_at`, `points`, filters)
	if err != nil {
		return GamificationReport{}, err
	}
	transactions, err := r.gamificationTransactions(ctx, filters)
	if err != nil {
		return GamificationReport{}, err
	}
	return GamificationReport{Metrics: metrics, ByReason: byReason, PointsByDay: pointsByDay, Transactions: transactions}, nil
}

func (r *Repository) Audit(ctx context.Context, filters Filters) (AuditReport, error) {
	metrics, err := r.auditMetrics(ctx, filters)
	if err != nil {
		return AuditReport{}, err
	}
	byAction, err := r.groupWithFilters(ctx, `
		SELECT action, COUNT(*)::float8
		FROM audit_log
		WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
		GROUP BY action
		ORDER BY COUNT(*) DESC
	`, filters)
	if err != nil {
		return AuditReport{}, err
	}
	byEntity, err := r.groupWithFilters(ctx, `
		SELECT entity_type, COUNT(*)::float8
		FROM audit_log
		WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
		GROUP BY entity_type
		ORDER BY COUNT(*) DESC
	`, filters)
	if err != nil {
		return AuditReport{}, err
	}
	activity, err := r.series(ctx, `audit_log`, `created_at`, filters)
	if err != nil {
		return AuditReport{}, err
	}
	entries, err := r.auditEntries(ctx, filters)
	if err != nil {
		return AuditReport{}, err
	}
	return AuditReport{Metrics: metrics, ByAction: byAction, ByEntity: byEntity, Activity: activity, Entries: entries}, nil
}

func (r *Repository) metrics(ctx context.Context, filters Filters) ([]Metric, error) {
	var volunteers, activeVolunteers, newVolunteers, events, applications, attendance, hours, completedTasks float64
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::float8 FROM volunteer_profiles),
			(SELECT COUNT(*)::float8 FROM volunteer_profiles WHERE status = 'active'),
			(SELECT COUNT(*)::float8 FROM volunteer_profiles WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)),
			(SELECT COUNT(*)::float8 FROM events WHERE starts_at BETWEEN parse_period_from($1) AND parse_period_to($2)),
			(SELECT COUNT(*)::float8 FROM event_applications WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)),
			(SELECT COUNT(*)::float8 FROM event_attendance WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)),
			(SELECT COALESCE(SUM(hours), 0)::float8 FROM time_entries WHERE status = 'approved' AND reviewed_at BETWEEN parse_period_from($1) AND parse_period_to($2)),
			(SELECT COUNT(*)::float8 FROM tasks WHERE status = 'completed' AND completion_confirmed_at BETWEEN parse_period_from($1) AND parse_period_to($2))
	`, filters.From, filters.To).Scan(&volunteers, &activeVolunteers, &newVolunteers, &events, &applications, &attendance, &hours, &completedTasks)
	if err != nil {
		return nil, err
	}
	return []Metric{
		{Code: "volunteers", Label: "Волонтеры", Value: volunteers},
		{Code: "activeVolunteers", Label: "Активные волонтеры", Value: activeVolunteers},
		{Code: "newVolunteers", Label: "Новые за период", Value: newVolunteers},
		{Code: "events", Label: "Мероприятия", Value: events},
		{Code: "applications", Label: "Заявки", Value: applications},
		{Code: "attendance", Label: "Посещения", Value: attendance},
		{Code: "hours", Label: "Подтвержденные часы", Value: hours},
		{Code: "completedTasks", Label: "Выполненные задачи", Value: completedTasks},
	}, nil
}

func (r *Repository) volunteerMetrics(ctx context.Context, filters Filters) ([]Metric, error) {
	all, err := r.metrics(ctx, filters)
	if err != nil {
		return nil, err
	}
	return all[:3], nil
}

func (r *Repository) eventMetrics(ctx context.Context, filters Filters) ([]Metric, error) {
	all, err := r.metrics(ctx, filters)
	if err != nil {
		return nil, err
	}
	return all[3:6], nil
}

func (r *Repository) taskMetrics(ctx context.Context, filters Filters) ([]Metric, error) {
	all, err := r.metrics(ctx, filters)
	if err != nil {
		return nil, err
	}
	return []Metric{all[7]}, nil
}

func (r *Repository) gamificationMetrics(ctx context.Context, filters Filters) ([]Metric, error) {
	var totalPoints, transactions, recipients, reasons float64
	err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(points), 0)::float8,
			COUNT(*)::float8,
			COUNT(DISTINCT user_id)::float8,
			COUNT(DISTINCT COALESCE(achievement_id::text, reason, source_type))::float8
		FROM points_transactions
		WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
	`, filters.From, filters.To).Scan(&totalPoints, &transactions, &recipients, &reasons)
	if err != nil {
		return nil, err
	}
	return []Metric{
		{Code: "totalPoints", Label: "Начислено баллов", Value: totalPoints},
		{Code: "transactions", Label: "Начислений", Value: transactions},
		{Code: "recipients", Label: "Получателей", Value: recipients},
		{Code: "reasons", Label: "Оснований", Value: reasons},
	}, nil
}

func (r *Repository) auditMetrics(ctx context.Context, filters Filters) ([]Metric, error) {
	var actions, users, failed, deletes float64
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)::float8,
			COUNT(DISTINCT user_id)::float8,
			COUNT(*) FILTER (WHERE status_code >= 400)::float8,
			COUNT(*) FILTER (WHERE action = 'delete')::float8
		FROM audit_log
		WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
	`, filters.From, filters.To).Scan(&actions, &users, &failed, &deletes)
	if err != nil {
		return nil, err
	}
	return []Metric{
		{Code: "actions", Label: "Действий", Value: actions},
		{Code: "users", Label: "Сотрудников", Value: users},
		{Code: "failed", Label: "С ошибкой", Value: failed},
		{Code: "deletes", Label: "Удалений", Value: deletes},
	}, nil
}

func (r *Repository) topVolunteers(ctx context.Context) ([]TopVolunteer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id::text, concat_ws(' ', u.last_name, u.first_name), u.email, vp.total_hours::float8, vp.points, vp.level
		FROM volunteer_profiles vp
		JOIN users u ON u.id = vp.user_id
		ORDER BY vp.total_hours DESC, vp.points DESC, u.last_name, u.first_name
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TopVolunteer{}
	for rows.Next() {
		var item TopVolunteer
		if err := rows.Scan(&item.UserID, &item.UserName, &item.Email, &item.TotalHours, &item.Points, &item.Level); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) group(ctx context.Context, query string) ([]ChartPoint, error) {
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartPoint{}
	for rows.Next() {
		var item ChartPoint
		if err := rows.Scan(&item.Label, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) groupWithFilters(ctx context.Context, query string, filters Filters) ([]ChartPoint, error) {
	rows, err := r.db.Query(ctx, query, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartPoint{}
	for rows.Next() {
		var item ChartPoint
		if err := rows.Scan(&item.Label, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) series(ctx context.Context, table string, field string, filters Filters) ([]ChartPoint, error) {
	query := `
		SELECT to_char(date_trunc('day', ` + field + `), 'DD.MM'), COUNT(*)::float8
		FROM ` + table + `
		WHERE ` + field + ` IS NOT NULL AND ` + field + ` BETWEEN parse_period_from($1) AND parse_period_to($2)
		GROUP BY date_trunc('day', ` + field + `)
		ORDER BY date_trunc('day', ` + field + `)
	`
	rows, err := r.db.Query(ctx, query, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartPoint{}
	for rows.Next() {
		var item ChartPoint
		if err := rows.Scan(&item.Label, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) seriesSum(ctx context.Context, table string, field string, valueField string, filters Filters) ([]ChartPoint, error) {
	query := `
		SELECT to_char(date_trunc('day', ` + field + `), 'DD.MM'), COALESCE(SUM(` + valueField + `), 0)::float8
		FROM ` + table + `
		WHERE ` + field + ` IS NOT NULL AND ` + field + ` BETWEEN parse_period_from($1) AND parse_period_to($2)
		GROUP BY date_trunc('day', ` + field + `)
		ORDER BY date_trunc('day', ` + field + `)
	`
	rows, err := r.db.Query(ctx, query, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChartPoint{}
	for rows.Next() {
		var item ChartPoint
		if err := rows.Scan(&item.Label, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) gamificationTransactions(ctx context.Context, filters Filters) ([]GamificationTransaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pt.id::text,
			pt.user_id::text,
			concat_ws(' ', u.last_name, u.first_name),
			u.email,
			COALESCE(a.name, ''),
			pt.source_type,
			COALESCE(pt.source_id::text, ''),
			pt.points,
			pt.reason,
			to_char(pt.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM points_transactions pt
		JOIN users u ON u.id = pt.user_id
		LEFT JOIN achievements a ON a.id = pt.achievement_id
		WHERE pt.created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
		ORDER BY pt.created_at DESC
		LIMIT 100
	`, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GamificationTransaction{}
	for rows.Next() {
		var item GamificationTransaction
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Email, &item.AchievementName, &item.SourceType, &item.SourceID, &item.Points, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) auditEntries(ctx context.Context, filters Filters) ([]AuditEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			al.id::text,
			COALESCE(al.user_id::text, ''),
			concat_ws(' ', u.last_name, u.first_name),
			COALESCE(u.email, ''),
			al.method,
			al.action,
			al.entity_type,
			COALESCE(al.entity_id, ''),
			al.path,
			al.status_code,
			to_char(al.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM audit_log al
		LEFT JOIN users u ON u.id = al.user_id
		WHERE al.created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
		ORDER BY al.created_at DESC
		LIMIT 100
	`, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AuditEntry{}
	for rows.Next() {
		var item AuditEntry
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Email, &item.Method, &item.Action, &item.EntityType, &item.EntityID, &item.Path, &item.StatusCode, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
