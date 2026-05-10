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

## Инфраструктурные команды

Для частых операций добавлен `Makefile`:

```powershell
make up
make down
make logs
make ps
make migrate
make seed
```

Отдельные backend-команды миграций и сидов можно запускать напрямую:

```powershell
docker compose run --rm backend go run ./cmd/migrate
docker compose run --rm backend go run ./cmd/seed
```

Для production-сборки добавлены отдельные Dockerfile:

```powershell
docker build -f unityaid-back/Dockerfile.prod -t unityaid-back:prod ./unityaid-back
docker build -f unityaid-front/Dockerfile.prod -t unityaid-front:prod ./unityaid-front
```

## Auth API

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET /api/v1/auth/me
POST /api/v1/auth/logout
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
POST /api/v1/auth/verify-email
POST /api/v1/auth/change-password
```

В development-режиме endpoints регистрации, подтверждения email и восстановления пароля возвращают dev-токен в поле `token`, чтобы сценарий можно было проверить без почтового сервиса.

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
