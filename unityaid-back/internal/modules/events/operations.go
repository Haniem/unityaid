package events

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) ListApplications(ctx context.Context, eventID string) ([]Application, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ea.id::text, ea.event_id::text, ea.user_id::text, concat_ws(' ', u.last_name, u.first_name), u.email,
			ea.status::text, ea.message, COALESCE(ea.rejection_reason, ''), ea.created_at, ea.updated_at
		FROM event_applications ea
		JOIN users u ON u.id = ea.user_id
		WHERE ea.event_id = $1
		ORDER BY ea.created_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Application{}
	for rows.Next() {
		var item Application
		if err := rows.Scan(&item.ID, &item.EventID, &item.UserID, &item.UserName, &item.Email, &item.Status, &item.Message, &item.RejectionReason, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateApplication(ctx context.Context, eventID string, userID string, message string) (Application, error) {
	var existingID string
	err := r.db.QueryRow(ctx, `SELECT id::text FROM event_applications WHERE event_id = $1 AND user_id = $2`, eventID, userID).Scan(&existingID)
	if err == nil {
		return Application{}, ErrAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Application{}, err
	}
	status, err := r.nextApplicationStatus(ctx, eventID)
	if err != nil {
		return Application{}, err
	}
	var id string
	err = r.db.QueryRow(ctx, `
		INSERT INTO event_applications (event_id, user_id, status, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, eventID, userID, status, message).Scan(&id)
	if err != nil {
		return Application{}, err
	}
	return r.findApplication(ctx, eventID, id)
}

func (r *Repository) UpdateApplicationStatus(ctx context.Context, eventID string, applicationID string, status string, rejectionReason string) (Application, error) {
	normalized := normalizeApplicationStatus(status)
	tag, err := r.db.Exec(ctx, `
		UPDATE event_applications
		SET status = $3,
			rejection_reason = CASE WHEN $3 = 'rejected' THEN $4 ELSE '' END,
			updated_at = now()
		WHERE event_id = $1 AND id = $2
	`, eventID, applicationID, normalized, rejectionReason)
	if err != nil {
		return Application{}, err
	}
	if tag.RowsAffected() == 0 {
		return Application{}, ErrNotFound
	}
	if normalized == "approved" {
		if err := r.RebalanceWaitlist(ctx, eventID); err != nil {
			return Application{}, err
		}
	}
	return r.findApplication(ctx, eventID, applicationID)
}

func (r *Repository) BulkUpdateApplications(ctx context.Context, eventID string, request BulkApplicationStatusRequest) ([]Application, error) {
	normalized := normalizeApplicationStatus(request.Status)
	rows, err := r.db.Query(ctx, `
		UPDATE event_applications
		SET status = $3,
			rejection_reason = CASE WHEN $3 = 'rejected' THEN $4 ELSE '' END,
			updated_at = now()
		WHERE event_id = $1 AND id::text = ANY($2::text[])
		RETURNING id::text
	`, eventID, request.ApplicationIDs, normalized, request.RejectionReason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if normalized == "approved" {
		if err := r.RebalanceWaitlist(ctx, eventID); err != nil {
			return nil, err
		}
	}
	items := make([]Application, 0, len(ids))
	for _, id := range ids {
		item, err := r.findApplication(ctx, eventID, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) DeleteApplication(ctx context.Context, eventID string, applicationID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM event_applications WHERE event_id = $1 AND id = $2`, eventID, applicationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteApplicationForUser(ctx context.Context, eventID string, applicationID string, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM event_applications WHERE event_id = $1 AND id = $2 AND user_id = $3`, eventID, applicationID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListAttendance(ctx context.Context, eventID string) ([]Attendance, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ea.id::text, ea.event_id::text, ea.user_id::text, concat_ws(' ', u.last_name, u.first_name), u.email,
			ea.check_in_at, ea.check_out_at, ea.hours::float8
		FROM event_attendance ea
		JOIN users u ON u.id = ea.user_id
		WHERE ea.event_id = $1
		ORDER BY u.last_name, u.first_name
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Attendance{}
	for rows.Next() {
		var item Attendance
		var inAt, outAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.EventID, &item.UserID, &item.UserName, &item.Email, &inAt, &outAt, &item.Hours); err != nil {
			return nil, err
		}
		if inAt.Valid {
			item.CheckInAt = &inAt.Time
		}
		if outAt.Valid {
			item.CheckOutAt = &outAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) MarkAttendance(ctx context.Context, eventID string, req AttendanceRequest) (Attendance, error) {
	event, err := r.FindByID(ctx, eventID)
	if err != nil {
		return Attendance{}, err
	}
	if req.CheckinCode != "" && req.CheckinCode != event.CheckinCode {
		return Attendance{}, errors.New("invalid checkin code")
	}
	hours := req.Hours
	if hours <= 0 {
		hours = event.EndsAt.Sub(event.StartsAt).Hours()
	}
	var id string
	err = r.db.QueryRow(ctx, `
		INSERT INTO event_attendance (event_id, user_id, check_in_at, hours)
		VALUES ($1, $2, now(), $3)
		ON CONFLICT (event_id, user_id)
		DO UPDATE SET check_in_at = COALESCE(event_attendance.check_in_at, now()), hours = EXCLUDED.hours, updated_at = now()
		RETURNING id::text
	`, eventID, req.UserID, hours).Scan(&id)
	if err != nil {
		return Attendance{}, err
	}
	items, err := r.ListAttendance(ctx, eventID)
	if err != nil {
		return Attendance{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return Attendance{}, ErrNotFound
}

func (r *Repository) BulkMarkAttendance(ctx context.Context, eventID string, req BulkAttendanceRequest) ([]Attendance, error) {
	event, err := r.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	hours := req.Hours
	if hours <= 0 {
		hours = event.EndsAt.Sub(event.StartsAt).Hours()
	}
	for _, userID := range req.UserIDs {
		_, err := r.db.Exec(ctx, `
			INSERT INTO event_attendance (event_id, user_id, check_in_at, hours)
			VALUES ($1, $2, now(), $3)
			ON CONFLICT (event_id, user_id)
			DO UPDATE SET check_in_at = COALESCE(event_attendance.check_in_at, now()), hours = EXCLUDED.hours, updated_at = now()
		`, eventID, userID, hours)
		if err != nil {
			return nil, err
		}
	}
	return r.ListAttendance(ctx, eventID)
}

func (r *Repository) UpdateAttendance(ctx context.Context, eventID string, attendanceID string, hours *float64, checkOutAt *time.Time) (Attendance, error) {
	if hours != nil && *hours < 0 {
		return Attendance{}, errors.New("hours must be non-negative")
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE event_attendance
		SET hours = COALESCE($3, hours),
			check_out_at = COALESCE($4, check_out_at),
			updated_at = now()
		WHERE event_id = $1 AND id = $2
	`, eventID, attendanceID, hours, checkOutAt)
	if err != nil {
		return Attendance{}, err
	}
	if tag.RowsAffected() == 0 {
		return Attendance{}, ErrNotFound
	}
	items, err := r.ListAttendance(ctx, eventID)
	if err != nil {
		return Attendance{}, err
	}
	for _, item := range items {
		if item.ID == attendanceID {
			return item, nil
		}
	}
	return Attendance{}, ErrNotFound
}

func (r *Repository) ListShifts(ctx context.Context, eventID string) ([]Shift, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, event_id::text, title, starts_at, ends_at, capacity, created_at FROM event_shifts WHERE event_id = $1 ORDER BY starts_at`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Shift{}
	for rows.Next() {
		var item Shift
		var capacity sql.NullInt32
		if err := rows.Scan(&item.ID, &item.EventID, &item.Title, &item.StartsAt, &item.EndsAt, &capacity, &item.CreatedAt); err != nil {
			return nil, err
		}
		if capacity.Valid {
			v := int(capacity.Int32)
			item.Capacity = &v
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateShift(ctx context.Context, eventID string, req ShiftRequest, startsAt, endsAt time.Time) (Shift, error) {
	var item Shift
	var capacity sql.NullInt32
	err := r.db.QueryRow(ctx, `
		INSERT INTO event_shifts (event_id, title, starts_at, ends_at, capacity)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, event_id::text, title, starts_at, ends_at, capacity, created_at
	`, eventID, req.Title, startsAt, endsAt, req.Capacity).Scan(&item.ID, &item.EventID, &item.Title, &item.StartsAt, &item.EndsAt, &capacity, &item.CreatedAt)
	if capacity.Valid {
		v := int(capacity.Int32)
		item.Capacity = &v
	}
	return item, err
}

func (r *Repository) ListFeedback(ctx context.Context, eventID string) ([]Feedback, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ef.id::text, ef.event_id::text, ef.user_id::text, concat_ws(' ', u.last_name, u.first_name), ef.rating, ef.comment, ef.created_at
		FROM event_feedback ef JOIN users u ON u.id = ef.user_id WHERE ef.event_id = $1 ORDER BY ef.created_at DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Feedback{}
	for rows.Next() {
		var item Feedback
		if err := rows.Scan(&item.ID, &item.EventID, &item.UserID, &item.UserName, &item.Rating, &item.Comment, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FeedbackSummary(ctx context.Context, eventID string) (float64, int, error) {
	var average float64
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(rating), 0)::float8, COUNT(*)::int
		FROM event_feedback
		WHERE event_id = $1
	`, eventID).Scan(&average, &count)
	return average, count, err
}

func (r *Repository) CreateFeedback(ctx context.Context, eventID, userID string, req FeedbackRequest) (Feedback, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO event_feedback (event_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (event_id, user_id) DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment
		RETURNING id::text
	`, eventID, userID, req.Rating, req.Comment).Scan(&id)
	if err != nil {
		return Feedback{}, err
	}
	items, err := r.ListFeedback(ctx, eventID)
	if err != nil {
		return Feedback{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return Feedback{}, ErrNotFound
}

func (r *Repository) CompleteEvent(ctx context.Context, eventID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status::text FROM events WHERE id = $1 FOR UPDATE`, eventID).Scan(&status); err != nil {
		return err
	}
	if status == "completed" {
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `
		UPDATE events SET status = 'completed', updated_at = now() WHERE id = $1;
		UPDATE volunteer_profiles vp
		SET total_hours = vp.total_hours + a.hours, updated_at = now()
		FROM event_attendance a
		WHERE a.event_id = $1 AND a.user_id = vp.user_id;
		SELECT recalculate_user_gamification(a.user_id)
		FROM event_attendance a
		WHERE a.event_id = $1;
	`, eventID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) findApplication(ctx context.Context, eventID string, applicationID string) (Application, error) {
	var item Application
	err := r.db.QueryRow(ctx, `
		SELECT ea.id::text, ea.event_id::text, ea.user_id::text, concat_ws(' ', u.last_name, u.first_name), u.email,
			ea.status::text, ea.message, COALESCE(ea.rejection_reason, ''), ea.created_at, ea.updated_at
		FROM event_applications ea JOIN users u ON u.id = ea.user_id
		WHERE ea.event_id = $1 AND ea.id = $2
	`, eventID, applicationID).Scan(&item.ID, &item.EventID, &item.UserID, &item.UserName, &item.Email, &item.Status, &item.Message, &item.RejectionReason, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) RebalanceWaitlist(ctx context.Context, eventID string) error {
	var maxParticipants sql.NullInt32
	if err := r.db.QueryRow(ctx, `SELECT max_participants FROM events WHERE id = $1`, eventID).Scan(&maxParticipants); err != nil {
		return err
	}
	if !maxParticipants.Valid {
		_, err := r.db.Exec(ctx, `
			UPDATE event_applications
			SET status = 'pending', updated_at = now()
			WHERE event_id = $1 AND status = 'waitlisted'
		`, eventID)
		return err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text
		FROM event_applications
		WHERE event_id = $1 AND status IN ('approved', 'waitlisted')
		ORDER BY
			CASE status WHEN 'approved' THEN 0 ELSE 1 END,
			created_at
	`, eventID)
	if err != nil {
		return err
	}
	defer rows.Close()
	approvedIDs := []string{}
	waitlistedIDs := []string{}
	index := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if index < int(maxParticipants.Int32) {
			approvedIDs = append(approvedIDs, id)
		} else {
			waitlistedIDs = append(waitlistedIDs, id)
		}
		index++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(approvedIDs) > 0 {
		if _, err := r.db.Exec(ctx, `UPDATE event_applications SET status = 'approved', updated_at = now() WHERE event_id = $1 AND id::text = ANY($2::text[])`, eventID, approvedIDs); err != nil {
			return err
		}
	}
	if len(waitlistedIDs) > 0 {
		if _, err := r.db.Exec(ctx, `UPDATE event_applications SET status = 'waitlisted', updated_at = now() WHERE event_id = $1 AND id::text = ANY($2::text[])`, eventID, waitlistedIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) nextApplicationStatus(ctx context.Context, eventID string) (string, error) {
	var maxParticipants sql.NullInt32
	var approved int
	err := r.db.QueryRow(ctx, `
		SELECT e.max_participants, COUNT(ea.id)
		FROM events e
		LEFT JOIN event_applications ea ON ea.event_id = e.id AND ea.status = 'approved'
		WHERE e.id = $1
		GROUP BY e.max_participants
	`, eventID).Scan(&maxParticipants, &approved)
	if err != nil {
		return "", err
	}
	if maxParticipants.Valid && approved >= int(maxParticipants.Int32) {
		return "waitlisted", nil
	}
	return "pending", nil
}

func normalizeApplicationStatus(status string) string {
	switch status {
	case "pending", "approved", "waitlisted", "rejected", "cancelled":
		return status
	default:
		return "pending"
	}
}
