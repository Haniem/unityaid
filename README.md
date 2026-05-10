# UnityAid

Информационная система управления волонтерами с элементами геймификации и аналитики.

## Локальный запуск через Docker

1. Скопируйте переменные окружения:

```powershell
Copy-Item .env.example .env
```

2. Соберите и запустите сервисы:

```powershell
docker compose up --build
```

3. Откройте приложения:

- Frontend: http://localhost:5173
- Backend health: http://localhost:8080/api/v1/health
- Backend DB health: http://localhost:8080/api/v1/health/db
- PostgreSQL: `localhost:5432`

## Демо-доступ

Все демо-пользователи используют пароль:

```text
password
```

Основной аккаунт:

```text
admin@unityaid.test
```

Другие аккаунты:

```text
org.admin@dobrye-ruki.test
coord@dobrye-ruki.test
volunteer1@test.local
volunteer2@test.local
org.admin@ecopulse.test
coord@ecopulse.test
volunteer3@test.local
```

## Auth API

```http
POST /api/v1/auth/login
GET /api/v1/auth/me
POST /api/v1/auth/logout
```

## News API

```http
GET /api/v1/news
POST /api/v1/news
GET /api/v1/news/{id}
PUT /api/v1/news/{id}
DELETE /api/v1/news/{id}
POST /api/v1/files/news-images
```

## Core CRUD API

```http
GET /api/v1/organizations
POST /api/v1/organizations
GET /api/v1/organizations/{id}
PUT /api/v1/organizations/{id}
DELETE /api/v1/organizations/{id}

GET /api/v1/events
POST /api/v1/events
GET /api/v1/events/{id}
PUT /api/v1/events/{id}
DELETE /api/v1/events/{id}

GET /api/v1/tasks
POST /api/v1/tasks
GET /api/v1/tasks/{id}
PUT /api/v1/tasks/{id}
DELETE /api/v1/tasks/{id}
```

После изменений backend-кода перезапустите backend-контейнер:

```powershell
docker compose restart backend
```

Миграции и сиды применяются при старте backend. Для применения новых миграций без пересоздания базы достаточно:

```powershell
docker compose restart backend
```

## Остановка

```powershell
docker compose down
```

Остановка с удалением данных БД:

```powershell
docker compose down -v
```

## Структура

```text
unityaid-back/   Go backend на Gin
unityaid-front/  Vue 3 frontend на Vite
docs/            документация проекта
```
