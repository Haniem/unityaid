package tasks

import (
	"context"
	"database/sql"
)

func (r *Repository) AddAssignment(ctx context.Context, taskID string, req AssignmentRequest) (Task, error) {
	role := req.Role
	if role == "" {
		role = "assignee"
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO task_assignments (task_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (task_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, taskID, req.UserID, role)
	if err != nil {
		return Task{}, err
	}
	return r.FindByID(ctx, taskID)
}

func (r *Repository) RemoveAssignment(ctx context.Context, taskID string, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM task_assignments WHERE task_id = $1 AND user_id = $2`, taskID, userID)
	return err
}

func (r *Repository) AddComment(ctx context.Context, taskID, userID string, req CommentRequest) (TaskComment, error) {
	var item TaskComment
	err := r.db.QueryRow(ctx, `
		INSERT INTO task_comments (task_id, user_id, content) VALUES ($1, $2, $3)
		RETURNING id::text, user_id::text, (SELECT concat_ws(' ', last_name, first_name) FROM users WHERE id = $2), content, created_at
	`, taskID, userID, req.Content).Scan(&item.ID, &item.UserID, &item.UserName, &item.Content, &item.CreatedAt)
	return item, err
}

func (r *Repository) ListComments(ctx context.Context, taskID string) ([]TaskComment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT tc.id::text, tc.user_id::text, concat_ws(' ', u.last_name, u.first_name), tc.content, tc.created_at
		FROM task_comments tc JOIN users u ON u.id = tc.user_id WHERE tc.task_id = $1 ORDER BY tc.created_at
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskComment{}
	for rows.Next() {
		var item TaskComment
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Content, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) AddAttachment(ctx context.Context, taskID, userID string, req AttachmentRequest) (TaskAttachment, error) {
	var item TaskAttachment
	err := r.db.QueryRow(ctx, `
		INSERT INTO task_attachments (task_id, user_id, file_name, file_url) VALUES ($1, $2, $3, $4)
		RETURNING id::text, file_name, file_url, created_at
	`, taskID, userID, req.FileName, req.FileURL).Scan(&item.ID, &item.FileName, &item.FileURL, &item.CreatedAt)
	return item, err
}

func (r *Repository) ListAttachments(ctx context.Context, taskID string) ([]TaskAttachment, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, file_name, file_url, created_at FROM task_attachments WHERE task_id = $1 ORDER BY created_at DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskAttachment{}
	for rows.Next() {
		var item TaskAttachment
		if err := rows.Scan(&item.ID, &item.FileName, &item.FileURL, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListStatusHistory(ctx context.Context, taskID string) ([]TaskStatusHistory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT tsh.id::text, tsh.from_status, tsh.to_status, concat_ws(' ', u.last_name, u.first_name), tsh.created_at
		FROM task_status_history tsh LEFT JOIN users u ON u.id = tsh.changed_by WHERE tsh.task_id = $1 ORDER BY tsh.created_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskStatusHistory{}
	for rows.Next() {
		var item TaskStatusHistory
		var fromStatus, userName sql.NullString
		if err := rows.Scan(&item.ID, &fromStatus, &item.ToStatus, &userName, &item.CreatedAt); err != nil {
			return nil, err
		}
		if fromStatus.Valid {
			item.FromStatus = &fromStatus.String
		}
		if userName.Valid {
			item.UserName = &userName.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) AddTimeEntry(ctx context.Context, taskID, userID string, req TimeEntryRequest) (TaskTimeEntry, error) {
	var item TaskTimeEntry
	err := r.db.QueryRow(ctx, `
		INSERT INTO task_time_entries (task_id, user_id, hours, note) VALUES ($1, $2, $3, $4)
		RETURNING id::text, user_id::text, (SELECT concat_ws(' ', last_name, first_name) FROM users WHERE id = $2), hours::float8, note, status, created_at
	`, taskID, userID, req.Hours, req.Note).Scan(&item.ID, &item.UserID, &item.UserName, &item.Hours, &item.Note, &item.Status, &item.CreatedAt)
	return item, err
}

func (r *Repository) ListTimeEntries(ctx context.Context, taskID string) ([]TaskTimeEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT te.id::text, te.user_id::text, concat_ws(' ', u.last_name, u.first_name), te.hours::float8, te.note, te.status, te.created_at
		FROM task_time_entries te JOIN users u ON u.id = te.user_id WHERE te.task_id = $1 ORDER BY te.created_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskTimeEntry{}
	for rows.Next() {
		var item TaskTimeEntry
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.Hours, &item.Note, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Approve(ctx context.Context, taskID, userID string) (Task, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE tasks
		SET status = 'completed', completion_confirmed_by = $2, completion_confirmed_at = now(), updated_at = now()
		WHERE id = $1;
		SELECT recalculate_user_gamification(ta.user_id)
		FROM task_assignments ta
		WHERE ta.task_id = $1;
	`, taskID, userID)
	if err != nil {
		return Task{}, err
	}
	return r.FindByID(ctx, taskID)
}
