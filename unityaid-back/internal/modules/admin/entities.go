package admin

var entityConfigs = map[string]EntityConfig{
	"users": {
		Code: "users", Label: "Пользователи", Table: "users", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "email", "first_name", "last_name", "patronymic", "avatar_url", "locale", "is_email_verified", "is_active", "last_login_at", "created_at", "updated_at"},
		Editable: []string{"email", "password_hash", "first_name", "last_name", "patronymic", "avatar_url", "locale", "is_email_verified", "is_active", "updated_at"},
		Search:   []string{"email", "first_name", "last_name"}, CanCreate: true, CanDelete: false,
		Description: "Аккаунты пользователей. Если password_hash не передан, используется временный пароль ChangeMe123!.",
	},
	"organizations": {
		Code: "organizations", Label: "Организации", Table: "organizations", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "name", "slug", "description", "contact_email", "logo_url", "website_url", "phone", "address", "deleted_at", "created_at", "updated_at"},
		Editable: []string{"name", "slug", "description", "contact_email", "logo_url", "website_url", "phone", "address", "deleted_at", "updated_at"},
		Search:   []string{"name", "slug", "description", "contact_email"}, CanCreate: true, CanDelete: true,
	},
	"organization_members": {
		Code: "organization_members", Label: "Участники организаций", Table: "organization_members", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "organization_id", "user_id", "role", "status", "created_at", "updated_at"},
		Editable: []string{"organization_id", "user_id", "role", "status", "updated_at"},
		Search:   []string{"role", "status"}, CanCreate: true, CanDelete: true,
	},
	"volunteer_profiles": {
		Code: "volunteer_profiles", Label: "Профили волонтеров", Table: "volunteer_profiles", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "user_id", "city", "phone", "bio", "status", "interests", "total_hours", "points", "level", "created_at", "updated_at"},
		Editable: []string{"user_id", "city", "phone", "bio", "status", "interests", "total_hours", "points", "level", "updated_at"},
		Search:   []string{"city", "phone", "bio", "status", "interests"}, CanCreate: true, CanDelete: true,
	},
	"skills": {
		Code: "skills", Label: "Навыки", Table: "skills", PrimaryKey: "id", OrderBy: "name ASC",
		Columns: []string{"id", "name", "created_at"}, Editable: []string{"name"}, Search: []string{"name"}, CanCreate: true, CanDelete: true,
	},
	"system_roles": {
		Code: "system_roles", Label: "Системные роли", Table: "system_roles", PrimaryKey: "id", OrderBy: "code ASC",
		Columns: []string{"id", "code", "name", "description", "created_at"}, Editable: []string{"code", "name", "description"}, Search: []string{"code", "name", "description"}, CanCreate: true, CanDelete: false,
	},
	"user_system_roles": {
		Code: "user_system_roles", Label: "Роли пользователей", Table: "user_system_roles", PrimaryKey: "user_id", OrderBy: "assigned_at DESC",
		Columns: []string{"user_id", "role_id", "assigned_at"}, Editable: []string{"user_id", "role_id"}, Search: []string{}, CanCreate: true, CanDelete: true,
	},
	"events": {
		Code: "events", Label: "Мероприятия", Table: "events", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "organization_id", "title", "description", "format", "status", "starts_at", "ends_at", "location", "max_participants", "checkin_code", "created_by", "created_at", "updated_at"},
		Editable: []string{"organization_id", "title", "description", "format", "status", "starts_at", "ends_at", "location", "max_participants", "checkin_code", "created_by", "updated_at"},
		Search:   []string{"title", "description", "location", "status"}, CanCreate: true, CanDelete: true,
	},
	"event_applications": {
		Code: "event_applications", Label: "Заявки на мероприятия", Table: "event_applications", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "event_id", "user_id", "status", "message", "created_at", "updated_at"},
		Editable: []string{"event_id", "user_id", "status", "message", "updated_at"},
		Search:   []string{"status", "message"}, CanCreate: true, CanDelete: true,
	},
	"event_attendance": {
		Code: "event_attendance", Label: "Посещаемость", Table: "event_attendance", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "event_id", "user_id", "check_in_at", "check_out_at", "hours", "created_at", "updated_at"},
		Editable: []string{"event_id", "user_id", "check_in_at", "check_out_at", "hours", "updated_at"},
		Search:   []string{}, CanCreate: true, CanDelete: true,
	},
	"event_shifts": {
		Code: "event_shifts", Label: "Смены мероприятий", Table: "event_shifts", PrimaryKey: "id", OrderBy: "starts_at DESC",
		Columns:  []string{"id", "event_id", "title", "starts_at", "ends_at", "capacity", "created_at"},
		Editable: []string{"event_id", "title", "starts_at", "ends_at", "capacity"},
		Search:   []string{"title"}, CanCreate: true, CanDelete: true,
	},
	"event_feedback": {
		Code: "event_feedback", Label: "Отзывы", Table: "event_feedback", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "event_id", "user_id", "rating", "comment", "created_at"},
		Editable: []string{"event_id", "user_id", "rating", "comment"},
		Search:   []string{"comment"}, CanCreate: true, CanDelete: true,
	},
	"tasks": {
		Code: "tasks", Label: "Задачи", Table: "tasks", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "organization_id", "event_id", "title", "description", "status", "priority", "due_at", "created_by", "completion_confirmed_by", "completion_confirmed_at", "created_at", "updated_at"},
		Editable: []string{"organization_id", "event_id", "title", "description", "status", "priority", "due_at", "created_by", "completion_confirmed_by", "completion_confirmed_at", "updated_at"},
		Search:   []string{"title", "description", "status", "priority"}, CanCreate: true, CanDelete: true,
	},
	"task_assignments": {
		Code: "task_assignments", Label: "Назначения задач", Table: "task_assignments", PrimaryKey: "id", OrderBy: "assigned_at DESC",
		Columns: []string{"id", "task_id", "user_id", "role", "assigned_at"}, Editable: []string{"task_id", "user_id", "role"}, Search: []string{"role"}, CanCreate: true, CanDelete: true,
	},
	"task_comments": {
		Code: "task_comments", Label: "Комментарии задач", Table: "task_comments", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "task_id", "user_id", "content", "created_at"}, Editable: []string{"task_id", "user_id", "content"}, Search: []string{"content"}, CanCreate: true, CanDelete: true,
	},
	"task_attachments": {
		Code: "task_attachments", Label: "Вложения задач", Table: "task_attachments", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "task_id", "user_id", "file_name", "file_url", "created_at"}, Editable: []string{"task_id", "user_id", "file_name", "file_url"}, Search: []string{"file_name", "file_url"}, CanCreate: true, CanDelete: true,
	},
	"task_status_history": {
		Code: "task_status_history", Label: "История статусов задач", Table: "task_status_history", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "task_id", "from_status", "to_status", "changed_by", "created_at"}, Editable: []string{"task_id", "from_status", "to_status", "changed_by"}, Search: []string{"from_status", "to_status"}, CanCreate: true, CanDelete: true,
	},
	"task_time_entries": {
		Code: "task_time_entries", Label: "Часы по задачам", Table: "task_time_entries", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "task_id", "user_id", "hours", "note", "status", "created_at"}, Editable: []string{"task_id", "user_id", "hours", "note", "status"}, Search: []string{"note", "status"}, CanCreate: true, CanDelete: true,
	},
	"time_entries": {
		Code: "time_entries", Label: "Учет часов", Table: "time_entries", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "organization_id", "user_id", "event_id", "task_id", "hours", "description", "status", "reviewed_by", "reviewed_at", "created_at", "updated_at"},
		Editable: []string{"organization_id", "user_id", "event_id", "task_id", "hours", "description", "status", "reviewed_by", "reviewed_at", "updated_at"},
		Search:   []string{"description", "status"}, CanCreate: true, CanDelete: true,
	},
	"news": {
		Code: "news", Label: "Новости", Table: "news", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "organization_id", "title", "slug", "summary", "content_html", "cover_image_url", "category_id", "status", "author_id", "scheduled_at", "published_at", "created_at", "updated_at"},
		Editable: []string{"organization_id", "title", "slug", "summary", "content_html", "cover_image_url", "category_id", "status", "author_id", "scheduled_at", "published_at", "updated_at"},
		Search:   []string{"title", "summary", "content_html", "status"}, CanCreate: true, CanDelete: true,
	},
	"news_categories": {
		Code: "news_categories", Label: "Категории новостей", Table: "news_categories", PrimaryKey: "id", OrderBy: "name ASC",
		Columns: []string{"id", "name", "slug", "created_at"}, Editable: []string{"name", "slug"}, Search: []string{"name", "slug"}, CanCreate: true, CanDelete: true,
	},
	"knowledge_articles": {
		Code: "knowledge_articles", Label: "База знаний", Table: "knowledge_articles", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "category_id", "title", "slug", "summary", "content_html", "status", "author_id", "published_at", "created_at", "updated_at"},
		Editable: []string{"category_id", "title", "slug", "summary", "content_html", "status", "author_id", "published_at", "updated_at"},
		Search:   []string{"title", "summary", "content_html", "status"}, CanCreate: true, CanDelete: true,
	},
	"knowledge_categories": {
		Code: "knowledge_categories", Label: "Категории базы знаний", Table: "knowledge_categories", PrimaryKey: "id", OrderBy: "name ASC",
		Columns: []string{"id", "name", "slug", "description", "created_at"}, Editable: []string{"name", "slug", "description"}, Search: []string{"name", "slug", "description"}, CanCreate: true, CanDelete: true,
	},
	"achievements": {
		Code: "achievements", Label: "Достижения", Table: "achievements", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "code", "name", "description", "icon", "points_reward", "created_at"}, Editable: []string{"code", "name", "description", "icon", "points_reward"}, Search: []string{"code", "name", "description"}, CanCreate: true, CanDelete: true,
	},
	"achievement_rules": {
		Code: "achievement_rules", Label: "Правила достижений", Table: "achievement_rules", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "achievement_id", "code", "trigger_type", "threshold", "points", "created_at"}, Editable: []string{"achievement_id", "code", "trigger_type", "threshold", "points"}, Search: []string{"code", "trigger_type"}, CanCreate: true, CanDelete: true,
	},
	"volunteer_achievements": {
		Code: "volunteer_achievements", Label: "Достижения волонтеров", Table: "volunteer_achievements", PrimaryKey: "id", OrderBy: "earned_at DESC",
		Columns: []string{"id", "user_id", "achievement_id", "earned_at"}, Editable: []string{"user_id", "achievement_id", "earned_at"}, Search: []string{}, CanCreate: true, CanDelete: true,
	},
	"points_transactions": {
		Code: "points_transactions", Label: "Начисления баллов", Table: "points_transactions", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns: []string{"id", "user_id", "achievement_id", "source_type", "source_id", "points", "reason", "created_at"}, Editable: []string{"user_id", "achievement_id", "source_type", "source_id", "points", "reason"}, Search: []string{"source_type", "reason"}, CanCreate: true, CanDelete: true,
	},
	"notifications": {
		Code: "notifications", Label: "Уведомления", Table: "notifications", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "user_id", "type", "title", "body", "link", "entity_type", "entity_id", "is_read", "read_at", "created_at"},
		Editable: []string{"user_id", "type", "title", "body", "link", "entity_type", "entity_id", "is_read", "read_at"},
		Search:   []string{"type", "title", "body", "entity_type", "entity_id"}, CanCreate: true, CanDelete: true,
	},
	"certificates": {
		Code: "certificates", Label: "Сертификаты", Table: "certificates", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "user_id", "organization_id", "type", "title", "description", "total_hours", "verify_code", "issued_by", "issued_at", "created_at"},
		Editable: []string{"user_id", "organization_id", "type", "title", "description", "total_hours", "verify_code", "issued_by", "issued_at"},
		Search:   []string{"type", "title", "description", "verify_code"}, CanCreate: true, CanDelete: true,
	},
	"audit_log": {
		Code: "audit_log", Label: "Аудит", Table: "audit_log", PrimaryKey: "id", OrderBy: "created_at DESC",
		Columns:  []string{"id", "user_id", "method", "action", "entity_type", "entity_id", "path", "status_code", "ip_address", "user_agent", "created_at"},
		Editable: []string{}, Search: []string{"method", "action", "entity_type", "path"}, CanCreate: false, CanDelete: false,
	},
}
