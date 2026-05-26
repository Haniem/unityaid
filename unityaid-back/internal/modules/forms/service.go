package forms

import (
	"context"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context, entity string, action string, id string) (Form, error) {
	form, err := s.baseForm(ctx, entity, action)
	if err != nil {
		return Form{}, err
	}
	if action == "edit" {
		form.ID = &id
		values, err := s.values(ctx, entity, id)
		if err != nil {
			return Form{}, err
		}
		applyValues(&form, values)
	}
	return form, nil
}

func (s *Service) ResourceOrganizationID(ctx context.Context, entity string, id string) (string, error) {
	return s.repository.resourceOrganizationID(ctx, entity, id)
}

func (s *Service) baseForm(ctx context.Context, entity string, action string) (Form, error) {
	switch entity {
	case "organizations":
		return s.organizationForm(action), nil
	case "news":
		categories, err := s.repository.newsCategories(ctx)
		if err != nil {
			return Form{}, err
		}
		return newsForm(action, categories), nil
	case "events":
		organizations, err := s.repository.organizations(ctx)
		if err != nil {
			return Form{}, err
		}
		return eventForm(action, organizations), nil
	case "tasks":
		organizations, err := s.repository.organizations(ctx)
		if err != nil {
			return Form{}, err
		}
		events, err := s.repository.events(ctx, "")
		if err != nil {
			return Form{}, err
		}
		return taskForm(action, organizations, events), nil
	case "volunteers":
		skills, err := s.repository.skills(ctx)
		if err != nil {
			return Form{}, err
		}
		return volunteerForm(action, skills), nil
	case "skills":
		return skillForm(action), nil
	default:
		return Form{}, ErrNotFound
	}
}

func (s *Service) values(ctx context.Context, entity string, id string) (map[string]any, error) {
	switch entity {
	case "organizations":
		return s.repository.organizationValues(ctx, id)
	case "news":
		return s.repository.newsValues(ctx, id)
	case "events":
		return s.repository.eventValues(ctx, id)
	case "tasks":
		return s.repository.taskValues(ctx, id)
	case "volunteers":
		return s.repository.volunteerValues(ctx, id)
	default:
		return nil, ErrNotFound
	}
}

func applyValues(form *Form, values map[string]any) {
	for index := range form.Fields {
		if value, ok := values[form.Fields[index].Code]; ok {
			form.Fields[index].Value = value
		}
	}
}

func textField(name string, code string, required bool, maxLength int, help string) FormField {
	return FormField{Name: name, Code: code, Type: "text", Require: required, Value: nil, MaxLength: maxLength, Help: help}
}

func selectField(name string, code string, required bool, value any, values []PossibleValue, help string) FormField {
	return FormField{Name: name, Code: code, Type: "select", Require: required, Value: value, PossibleValues: values, Help: help}
}

func option(id string, name string) PossibleValue {
	return PossibleValue{ID: id, Name: name}
}

func statusValues(items ...PossibleValue) []PossibleValue {
	return items
}

func (s *Service) organizationForm(action string) Form {
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "организации")},
		Fields: []FormField{
			textField("Название", "name", true, 255, "Публичное название организации."),
			textField("Slug", "slug", false, 120, "Короткий адрес латиницей; если оставить пустым, система создаст его из названия."),
			{Name: "Email", Code: "contactEmail", Type: "email", Require: false, Value: nil, MaxLength: 255, Help: "Контактный email для связи с организацией."},
			{Name: "Телефон", Code: "phone", Type: "tel", Require: false, Value: nil, Placeholder: "+7 900 000-00-00", Help: "Телефон в международном или российском формате."},
			{Name: "Сайт", Code: "websiteUrl", Type: "url", Require: false, Value: nil, Placeholder: "https://example.org", Help: "Официальный сайт или страница организации."},
			{Name: "Логотип", Code: "logoUrl", Type: "file", Require: false, Value: nil, UploadEndpoint: "/files/organization-logos", Accept: "image/*", Help: "Изображение будет сохранено на backend и подставлено в профиль организации."},
			textField("Адрес", "address", false, 255, "Фактический адрес или основная площадка."),
			{Name: "Описание", Code: "description", Type: "textarea", Require: false, Value: "", Rows: 5, Help: "Кратко опишите миссию и направление работы."},
		},
	}
}

func newsForm(action string, categories []PossibleValue) Form {
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "новости")},
		Fields: []FormField{
			textField("Заголовок", "title", true, 255, "Название новости в списке и детальной странице."),
			{Name: "Краткое описание", Code: "summary", Type: "textarea", Require: false, Value: "", Rows: 3, Help: "Короткий анонс для карточки новости."},
			selectField("Статус", "status", true, "published", statusValues(option("published", "Опубликовано"), option("scheduled", "Запланировано"), option("draft", "Черновик")), "Запланированные новости публикуются после выбранной даты."),
			{Name: "Дата публикации", Code: "scheduledAt", Type: "datetime", Require: false, Value: nil, Help: "Нужна только для статуса «Запланировано»."},
			selectField("Категория", "categoryId", false, nil, append([]PossibleValue{option("", "Без категории")}, categories...), "Помогает фильтровать новости."),
			{Name: "Обложка", Code: "coverImageUrl", Type: "file", Require: false, Value: nil, UploadEndpoint: "/files/news-images", Accept: "image/*", Help: "Изображение будет храниться на backend."},
		},
	}
}

func eventForm(action string, organizations []PossibleValue) Form {
	start := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	end := time.Now().Add(26 * time.Hour).Format(time.RFC3339)
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "мероприятия")},
		Fields: []FormField{
			selectField("Организация", "organizationId", true, firstValue(organizations), organizations, "Организация, которая проводит мероприятие."),
			textField("Название", "title", true, 255, "Короткое и понятное название события."),
			{Name: "Описание", Code: "description", Type: "textarea", Require: false, Value: "", Rows: 4, Help: "Что будет происходить и кто нужен."},
			selectField("Формат", "format", true, "offline", statusValues(option("offline", "Офлайн"), option("online", "Онлайн"), option("hybrid", "Гибрид")), "Формат участия волонтеров."),
			selectField("Статус", "status", true, "published", statusValues(option("draft", "Черновик"), option("published", "Опубликовано"), option("completed", "Завершено"), option("cancelled", "Отменено")), "Опубликованные мероприятия видны волонтерам."),
			{Name: "Начало", Code: "startsAt", Type: "datetime", Require: true, Value: start, Help: "Дата и время старта мероприятия."},
			{Name: "Окончание", Code: "endsAt", Type: "datetime", Require: true, Value: end, Help: "Дата и время завершения."},
			textField("Место", "location", false, 255, "Адрес, ссылка или площадка."),
			{Name: "Лимит участников", Code: "maxParticipants", Type: "number", Require: false, Value: nil, Min: intPtr(1), Help: "Оставьте пустым, если ограничения нет."},
		},
	}
}

func taskForm(action string, organizations []PossibleValue, events []PossibleValue) Form {
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "задачи")},
		Fields: []FormField{
			selectField("Организация", "organizationId", true, firstValue(organizations), organizations, "Задача будет привязана к выбранной организации."),
			selectField("Мероприятие", "eventId", false, nil, append([]PossibleValue{option("", "Без мероприятия")}, events...), "Необязательная связь с мероприятием."),
			textField("Название", "title", true, 255, "Что нужно сделать."),
			{Name: "Описание", Code: "description", Type: "textarea", Require: false, Value: "", Rows: 4, Help: "Детали, критерии готовности, ссылки."},
			selectField("Статус", "status", true, "created", statusValues(option("created", "Создана"), option("assigned", "Назначена"), option("in_progress", "В работе"), option("review", "На проверке"), option("completed", "Выполнена"), option("cancelled", "Отменена")), "Текущий этап выполнения."),
			selectField("Приоритет", "priority", true, "medium", statusValues(option("low", "Низкий"), option("medium", "Средний"), option("high", "Высокий")), "Помогает координатору сортировать работу."),
			{Name: "Срок", Code: "dueAt", Type: "datetime", Require: false, Value: nil, Help: "Дата, к которой задачу желательно завершить."},
		},
	}
}

func volunteerForm(action string, skills []PossibleValue) Form {
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "профиля волонтера")},
		Fields: []FormField{
			textField("Имя", "firstName", true, 120, "Имя волонтера."),
			textField("Фамилия", "lastName", true, 120, "Фамилия волонтера."),
			textField("Отчество", "patronymic", false, 120, "Если используется в документах."),
			{Name: "Аватар", Code: "avatarUrl", Type: "file", Require: false, Value: nil, UploadEndpoint: "/files/profile-avatars", Accept: "image/*", Help: "Загрузите JPG, PNG, WEBP или GIF размером до 5 МБ."},
			textField("Город", "city", false, 120, "Город, где волонтер чаще всего участвует."),
			{Name: "Телефон", Code: "phone", Type: "tel", Require: false, Value: nil, Placeholder: "+7 900 000-00-00", Help: "Телефон для связи координатора."},
			selectField("Статус", "status", true, "active", statusValues(option("new", "Новый"), option("active", "Активный"), option("unavailable", "Временно недоступен"), option("archived", "Архивный")), "Рабочий статус волонтера для координаторов и фильтрации команды."),
			{Name: "Интересы", Code: "interests", Type: "textarea", Require: false, Value: "", Rows: 3, Help: "Направления, в которых волонтер хочет участвовать: помощь людям, экология, события, медиа."},
			{Name: "О себе", Code: "bio", Type: "textarea", Require: false, Value: "", Rows: 5, Help: "Опыт, интересы, доступность."},
			{Name: "Навыки", Code: "skillIds", Type: "select", Require: false, Value: []string{}, Multi: true, PossibleValues: skills, Help: "Можно выбрать несколько навыков."},
		},
	}
}

func skillForm(action string) Form {
	return Form{
		Lang: "ru",
		Meta: FormMeta{Title: titleByAction(action, "навыка")},
		Fields: []FormField{
			textField("Название", "name", true, 120, "Короткое название навыка."),
		},
	}
}

func titleByAction(action string, entity string) string {
	if action == "edit" {
		return "Редактирование " + entity
	}
	return "Создание " + entity
}

func firstValue(values []PossibleValue) any {
	if len(values) == 0 {
		return nil
	}
	return values[0].ID
}

func intPtr(value int) *int {
	return &value
}
