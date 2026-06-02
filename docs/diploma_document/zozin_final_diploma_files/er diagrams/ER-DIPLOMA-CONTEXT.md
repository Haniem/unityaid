# Контекст: ER-диаграммы UnityAid для диплома

Документ для переноса в другой чат / главы диплома. Описывает, как получена схема БД, какие проблемы были решены и какие артефакты использовать.

---

## 1. Проект и база данных

| Параметр | Значение |
|----------|----------|
| Проект | UnityAid (диплом) |
| Расположение кода | `E:\diploma` (в WSL: `/mnt/e/diploma`) |
| СУБД | PostgreSQL 17 (Docker) |
| Контейнер | `unityaid-db` |
| Образ | `postgres:17-alpine` |
| База / пользователь / пароль | `unityaid` / `unityaid` / `unityaid` |
| Порт на хосте | **5433** (в `.env`: `POSTGRES_PORT=5433`) |
| Запуск БД | `cd /mnt/e/diploma && docker compose up -d db` |

**Важно:** на Windows отдельно установлен PostgreSQL 17 на порту **5432**. DBeaver при подключении к `localhost:5432` попадает в **пустую** Windows-БД, а не в Docker. Для работы с реальными данными нужен порт **5433**.

---

## 2. Что было не так (типичные ошибки)

### 2.1. Неверный тип дампа в DBeaver

Использовался **PostgreSQL global dump** (`pg_dumpall`) вместо **Dump database** (`pg_dump` одной БД).

Симптомы:
- В файле заголовок `PostgreSQL database cluster dump`
- Только роли, `CREATE DATABASE`, без `CREATE TABLE`
- Или пустые секции для `unityaid`

**Правильно:** ПКМ по базе **`unityaid`** (порт 5433) → **Tools → Dump database** → Format: **Plain**, при необходимости **schema only**.

### 2.2. Подключение не к той БД

- Docker (WSL): таблицы есть, ~46 в `public`
- Windows PostgreSQL 5432: база `unityaid` пустая, пользователя `unityaid` нет

### 2.3. Формат Custom вместо Plain

Файлы с началом `PGDMP` — бинарный **Custom**, не текстовый SQL. Для ER-сайтов не подходят.

### 2.4. schema.biz не рисовал связи

Причины:
1. Вставляли только `CREATE TYPE` + `CREATE TABLE`, **без FK** (в pg_dump связи часто в `ALTER TABLE ... ADD CONSTRAINT` в конце).
2. Отдельные `ALTER TABLE ADD FOREIGN KEY` в конце файла **schema.biz не отображает** — нужны **inline** `REFERENCES` внутри `CREATE TABLE`.
3. Служебные строки PostgreSQL 17: `\restrict`, `\connect`, функции с `$$`, `gen_random_uuid()` — парсер ломается (ошибки `Unexpected token '\'`, `'$'`).

---

## 3. Итоговая полная схema (46 таблиц)

Полный список таблиц в `public` (актуально на момент экспорта):

**Пользователи и доступ:** `users`, `organization_members`, `system_roles`, `user_system_roles`, `user_invitations`, `refresh_sessions`, `revoked_access_tokens`, `email_verification_tokens`, `password_reset_tokens`

**Организации:** `organizations`, `tenant_settings`

**Волонтёры:** `volunteer_profiles`, `volunteer_skills`, `skills`

**Мероприятия:** `events`, `event_templates`, `event_applications`, `event_attendance`, `event_feedback`, `event_shifts`, `field_qr_tokens`, `field_checkins`

**Задачи:** `tasks`, `task_assignments`, `task_comments`, `task_attachments`, `task_status_history`, `task_time_entries`

**Учёт времени:** `time_entries`

**Геймификация:** `achievements`, `achievement_rules`, `volunteer_achievements`, `points_transactions`

**Контент:** `news`, `news_categories`, `knowledge_articles`, `knowledge_categories`, `notifications`

**Профиль (кастомные поля):** `profile_field_groups`, `profile_field_definitions`, `profile_field_options`, `profile_field_values`

**Прочее:** `certificates`, `audit_log`, `schema_migrations`, `seed_migrations`

---

## 4. Сгенерированные файлы

### 4.1. Папка для диплома (основное)

**Путь:** `c:\Users\Павел\diploma-er\`

| Файл | Назначение |
|------|------------|
| `01-core.sql` / `01-core.mmd` | **Основная ER для главы диплома** (10 таблиц) |
| `02-gamification.sql` / `02-gamification.mmd` | Геймификация (8 таблиц) |
| `03-content.sql` / `03-content.mmd` | Контент и уведомления (7 таблиц) |
| `04-roles.sql` / `04-roles.mmd` | Роли и членство (5 таблиц) |
| `README.md` | Краткая справка |
| `ER-DIPLOMA-CONTEXT.md` | Этот документ |

**Визуализация:**
- `.sql` → [schema.biz](https://schema.biz/database/sql-to-diagram/) (вставить целиком, zoom out / fit)
- `.mmd` → [mermaid.live](https://mermaid.live) → Export PNG/SVG для Word

### 4.2. Полная схема (справочно)

| Файл | Описание |
|------|----------|
| `c:\Users\Павел\unityaid-er-diagram.sql` | Все 46 таблиц, 72 FK, inline `REFERENCES`, без `ALTER TABLE` |
| `c:\Users\Павел\dump-unityaid-202606020905-er.sql` | Промежуточная версия с ALTER FK |
| `E:\diploma\unityaid-schema-only.sql` | schema-only из Docker (`pg_dump --schema-only`) |

### 4.3. Скрипты генерации (при необходимости пересобрать)

| Скрипт | Что делает |
|--------|------------|
| `c:\Users\Павел\Downloads\build_er.py` | Полная ER из дампа |
| `c:\Users\Павел\Downloads\build_er_inline.py` | Inline REFERENCES для schema.biz |
| `c:\Users\Павел\Downloads\build_diploma_er.py` | 4 дипломных диаграммы |

Запуск (WSL): `wsl -e python3 /mnt/c/Users/Павел/Downloads/build_diploma_er.py`

---

## 5. Четыре ER-диаграммы для диплома

### 5.1. `01-core` — ядро платформы (10 таблиц) ★ основная

**Таблицы:** `users`, `organizations`, `organization_members`, `volunteer_profiles`, `events`, `event_applications`, `event_attendance`, `tasks`, `task_assignments`, `time_entries`

**Смысл:** пользователи, организации, членство, профиль волонтёра, мероприятия (заявки, посещаемость), задачи, назначения, учёт часов.

**Связи (логика):**
- `users` ↔ `organization_members` ↔ `organizations`
- `users` → `volunteer_profiles`
- `organizations` → `events` → `event_applications` / `event_attendance` ← `users`
- `organizations` → `tasks` ← `events` (опционально)
- `tasks` → `task_assignments` ← `users`
- `time_entries` → org, user, event, task

### 5.2. `02-gamification` — геймификация (8 таблиц)

**Таблицы:** `users`, `volunteer_profiles`, `achievements`, `achievement_rules`, `volunteer_achievements`, `points_transactions`, `skills`, `volunteer_skills`

**Смысл:** достижения, правила начисления, история баллов, навыки волонтёра.

### 5.3. `03-content` — контент (7 таблиц)

**Таблицы:** `users`, `organizations`, `news`, `news_categories`, `knowledge_articles`, `knowledge_categories`, `notifications`

**Смысл:** новости, база знаний, уведомления пользователям.

### 5.4. `04-roles` — роли (5 таблиц, приложение)

**Таблицы:** `users`, `system_roles`, `user_system_roles`, `organization_members`, `organizations`

**Смысл:** системные роли (super_admin и т.д.) и роли внутри организации (org_admin, coordinator, volunteer).

---

## 6. Что намеренно убрали из дипломных диаграмм

Не показывать на ER (упомянуть текстом или в приложении):

| Категория | Таблицы / поля |
|-----------|----------------|
| Технические миграции | `schema_migrations`, `seed_migrations` |
| Сессии и токены | `refresh_sessions`, `revoked_access_tokens`, `email_verification_tokens`, `password_reset_tokens` |
| Аудит | `audit_log` |
| Служебные ID | `display_id` (человекочитаемые номера) |
| Детализация задач | `task_comments`, `task_attachments`, `task_status_history`, `task_time_entries` |
| Полевые QR | `field_qr_tokens`, `field_checkins` |
| Кастомные поля профиля | `profile_field_*` (4 таблицы) |
| Настройки тенанта | `tenant_settings` |
| Шаблоны/смены/отзывы | `event_templates`, `event_shifts`, `event_feedback` |
| Сертификаты | `certificates` (можно отдельным рисунком) |

Из колонок убраны: `created_at`, `updated_at`, хеши, длинные тексты (`content_html`, `password_hash`), чтобы диаграмма читалась на А4.

---

## 7. Рекомендации для текста диплома

### Глава «Проектирование БД» / «Модель данных»

1. Указать СУБД: **PostgreSQL 17**, развёртывание в **Docker**.
2. Привести **логическую ER-диаграмму ядра** (`01-core`) — ~10 сущностей, читаемо на одной странице.
3. Текстом: полная физическая модель содержит **46 таблиц**; детализация по подсистемам — на рисунках 2–4 или в приложении.
4. Кратко описать нормализацию: отдельные сущности для заявок, посещаемости, назначений на задачи, транзакций баллов.

### Формулировки для пояснительной записки

> Логическая модель данных платформы UnityAid построена вокруг сущностей «Пользователь», «Организация», «Мероприятие», «Задача» и «Учёт времени». Связь пользователя с организацией реализована через ассоциативную сущность «Членство в организации» с атрибутом роли. Участие волонтёра в мероприятии описывается заявкой и фактом посещаемости. Подсистемы геймификации, контента и управления доступом вынесены на отдельные ER-диаграммы для наглядности.

### Рисунки (пример нумерации)

- Рисунок X — ER-диаграмма ядра информационной системы UnityAid  
- Рисунок X+1 — ER-диаграмма подсистемы геймификации  
- Рисунок X+2 — ER-диаграмма подсистемы контента  
- Рисунок X+3 — ER-диаграмма ролей и доступа (приложение)

---

## 8. Команды для повторного экспорта

### Schema-only из Docker

```bash
cd /mnt/e/diploma
docker exec unityaid-db pg_dump -U unityaid -d unityaid \
  --schema-only --format=plain --no-owner \
  > /mnt/e/diploma/unityaid-schema-only.sql
```

### DBeaver

- Host: `localhost`, Port: **5433**, DB/User/Pass: `unityaid`
- Dump database → Plain → Schema only → без create/drop database

### Проверка таблиц

```bash
docker exec unityaid-db psql -U unityaid -d unityaid -c "\dt public.*"
```

---

## 9. Чеклист для нового чата (AI / соавтор)

- [ ] БД в Docker на порту **5433**, не Windows 5432  
- [ ] Для ER использовать файлы из **`c:\Users\Павел\diploma-er\`**, не сырой global dump  
- [ ] Основная диаграмма диплома: **`01-core`** (10 таблиц)  
- [ ] Полная модель: 46 таблиц, файл `unityaid-er-diagram.sql`  
- [ ] schema.biz: только `.sql` с inline `REFERENCES`, без `\restrict` и `ALTER TABLE` FK  
- [ ] Картинки: mermaid.live из `.mmd` или schema.biz из `.sql`  
- [ ] Технические таблицы (tokens, migrations, audit) — не на основной ER  

---

## 10. История действий (кратко)

1. Попытки global dump и подключение к Windows PostgreSQL → пустые/неправильные дампы.  
2. Выявлен Docker `unityaid-db`, порт 5433, учётные данные `unityaid`.  
3. Получен Plain dump с 46 таблицами и данными.  
4. Подготовлен `unityaid-er-diagram.sql` с inline FK для schema.biz.  
5. 46 таблиц признаны избыточными для одной страницы диплома.  
6. Схема разбита на 4 предметные ER-диаграммы в `diploma-er/`.  
7. Пользователь построил 4 визуальные ER-диаграммы по этим файлам.  

---

*Дата актуализации контекста: 2026-06-02*
