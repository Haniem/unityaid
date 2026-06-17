package analytics

import (
	"context"
	"errors"
)

func (r *Repository) Management(ctx context.Context, kind string, filters Filters) (ManagementReport, error) {
	metrics, err := r.metrics(ctx, filters)
	if err != nil {
		return ManagementReport{}, err
	}
	rows, err := r.managementRows(ctx, kind, filters)
	if err != nil {
		return ManagementReport{}, err
	}
	risks, err := r.riskRows(ctx, filters)
	if err != nil {
		return ManagementReport{}, err
	}
	return ManagementReport{
		Code:    kind,
		Title:   managementTitle(kind),
		Metrics: selectManagementMetrics(kind, metrics),
		Rows:    rows,
		Risks:   risks,
	}, nil
}

func managementTitle(kind string) string {
	switch kind {
	case "executive":
		return "Executive dashboard"
	case "management":
		return "Отчет для руководства"
	case "grant":
		return "Отчет для грантодателя"
	case "branches":
		return "Отчет по филиалам и организациям"
	case "coordinators":
		return "Отчет по координаторам и сотрудникам"
	case "risks":
		return "Отчет по проблемным зонам"
	default:
		return ""
	}
}

func selectManagementMetrics(kind string, metrics []Metric) []Metric {
	if kind == "risks" {
		return metrics
	}
	if len(metrics) > 6 {
		return metrics[:6]
	}
	return metrics
}

func (r *Repository) managementRows(ctx context.Context, kind string, filters Filters) ([]ReportRow, error) {
	switch kind {
	case "executive", "management":
		return r.reportRows(ctx, `
			SELECT 'Динамика заявок', to_char(date_trunc('day', created_at), 'DD.MM'), COUNT(*)::float8, 'Заявки за день'
			FROM event_applications
			WHERE created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			GROUP BY date_trunc('day', created_at)
			ORDER BY date_trunc('day', created_at)
		`, filters)
	case "grant":
		return r.reportRows(ctx, `
			SELECT o.name, 'Подтвержденные часы', COALESCE(SUM(te.hours), 0)::float8, 'Часы волонтеров за период'
			FROM organizations o
			LEFT JOIN time_entries te ON te.organization_id = o.id AND te.status = 'approved' AND te.reviewed_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			GROUP BY o.name
			ORDER BY COALESCE(SUM(te.hours), 0) DESC
		`, filters)
	case "branches":
		return r.reportRows(ctx, `
			SELECT o.name, 'Активность', COUNT(DISTINCT e.id)::float8 + COUNT(DISTINCT t.id)::float8, 'Мероприятия и задачи'
			FROM organizations o
			LEFT JOIN events e ON e.organization_id = o.id AND e.starts_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			LEFT JOIN tasks t ON t.organization_id = o.id AND t.created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			GROUP BY o.name
			ORDER BY COUNT(DISTINCT e.id)::float8 + COUNT(DISTINCT t.id)::float8 DESC
		`, filters)
	case "coordinators":
		return r.reportRows(ctx, `
			SELECT concat_ws(' ', u.last_name, u.first_name), 'Действия', COUNT(al.id)::float8, u.email
			FROM users u
			JOIN audit_log al ON al.user_id = u.id
			WHERE al.created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			GROUP BY u.id, u.last_name, u.first_name, u.email
			ORDER BY COUNT(al.id) DESC
			LIMIT 50
		`, filters)
	case "risks":
		return r.riskRows(ctx, filters)
	default:
		return nil, errors.New("unknown report")
	}
}

func (r *Repository) riskRows(ctx context.Context, filters Filters) ([]ReportRow, error) {
	return r.reportRows(ctx, `
		SELECT label, group_name, value, details
		FROM (
			SELECT 'Ожидающие заявки' AS label, e.title AS group_name, COUNT(ea.id)::float8 AS value, 'Требуют решения координатора' AS details
			FROM events e
			JOIN event_applications ea ON ea.event_id = e.id AND ea.status = 'pending'
			GROUP BY e.title
			UNION ALL
			SELECT 'Непроверенные часы', o.name, COUNT(te.id)::float8, 'Записи времени ожидают подтверждения'
			FROM organizations o
			JOIN time_entries te ON te.organization_id = o.id AND te.status = 'pending'
			GROUP BY o.name
			UNION ALL
			SELECT 'Ошибки API', COALESCE(entity_type, 'system'), COUNT(*)::float8, 'Аудит действий со статусом >= 400'
			FROM audit_log
			WHERE status_code >= 400 AND created_at BETWEEN parse_period_from($1) AND parse_period_to($2)
			GROUP BY COALESCE(entity_type, 'system')
		) risks
		ORDER BY value DESC
		LIMIT 30
	`, filters)
}

func (r *Repository) reportRows(ctx context.Context, query string, filters Filters) ([]ReportRow, error) {
	rows, err := r.db.Query(ctx, query, filters.From, filters.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ReportRow{}
	for rows.Next() {
		var item ReportRow
		if err := rows.Scan(&item.Label, &item.Group, &item.Value, &item.Details); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
