# Контекст для следующего ИИ-агента по проекту «Пульс»

Дата контекста: 2026-06-02.

Рабочая папка проекта: `E:\diploma`.

## 1. Общая цель

Проект — информационная система управления волонтерами. Текущее пользовательское название: **«Пульс»**.

Историческое название проекта и части технических идентификаторов: `UnityAid` / `unityaid`.

Важно: не переименовывать технические идентификаторы механически, если это может сломать проект. Названия папок, Go module, Docker service/container names, localStorage keys и импорты могут оставаться `unityaid`. Пользовательский интерфейс и диплом должны использовать название **«Пульс»**.

Основные части:

- Backend: `E:\diploma\unityaid-back`
- Frontend: `E:\diploma\unityaid-front`
- Диплом: `E:\diploma\docs\Документ диплома\Финальный файлы диплома Зозин`
- Контексты: `E:\diploma\docs\contexts`

## 2. Версии диплома

Не изменять старые версии:

- `Zozin_UnityAid_VKR_final_V1.docx` — исходная V1.
- `Zozin_UnityAid_VKR_final_V2_dorabotannaya.docx` — V2.
- `Zozin_Puls_VKR_final_V3_dorabotannaya.docx` — V3 после ребрендинга.

Рабочая V4:

- Генератор: `build_diploma_v4_puls.py`
- Целевой DOCX: `Zozin_Puls_VKR_final_V4_dorabotannaya.docx`
- Скриншоты V4: `V4_generated/screenshots`

Связанные файлы:

- `ТЗ_на_финальную_доработку_ВКР_UnityAid_V1.md`
- `Сравнение_ВКР_Афонькина_с_V3_и_план_V4_Пульс.md`
- `Краткое_описание_доработок_V2_относительно_V1.md`
- `Краткое_описание_доработок_V4_относительно_V3.md`

## 3. Что уже реализовано в приложении

### 3.1 Ребрендинг

Пользовательские упоминания `UnityAid` заменены на **«Пульс»** во frontend/backend и в пояснительной записке V3/V4.

Часть seed-файлов уже изменена на `admin@puls.test`, но если локальная база была создана раньше, в ней может остаться `admin@unityaid.test`.

### 3.2 Backend V4: настраиваемые поля профиля

Добавлена миграция:

- `unityaid-back/migrations/020_profile_fields.sql`

Добавлены таблицы:

- `profile_field_groups`
- `profile_field_definitions`
- `profile_field_options`
- `profile_field_values`

Добавлен модуль:

- `unityaid-back/internal/modules/profilefields/domain.go`
- `unityaid-back/internal/modules/profilefields/repository.go`
- `unityaid-back/internal/modules/profilefields/service.go`
- `unityaid-back/internal/modules/profilefields/handler.go`

Подключено в:

- `unityaid-back/internal/http/router.go`

API:

- `GET /api/v1/profile-fields/schema?organizationId=...`
- `GET /api/v1/profile-fields/users/:userId?organizationId=...`
- `PUT /api/v1/profile-fields/users/:userId`
- `POST /api/v1/profile-fields/groups`
- `PUT /api/v1/profile-fields/groups/:id`
- `DELETE /api/v1/profile-fields/groups/:id`
- `POST /api/v1/profile-fields/fields`
- `PUT /api/v1/profile-fields/fields/:id`
- `DELETE /api/v1/profile-fields/fields/:id`

Особенности:

- системные группы и поля создаются автоматически;
- системные группы/поля нельзя удалить;
- удаление системного поля проверялось и возвращало HTTP `409`;
- значения пользовательских полей сохраняются в `jsonb`;
- доступ проверяется через существующий `auth.Authorizer`;
- audit middleware уже применяется к изменяющим маршрутам.

### 3.3 Backend V4: аватар

Добавлен endpoint:

- `POST /api/v1/files/profile-avatars`

Файл:

- `unityaid-back/internal/modules/files/handler.go`

В форме волонтера поле аватара изменено с URL на file-upload:

- `unityaid-back/internal/modules/forms/service.go`

### 3.4 Frontend V4

Добавлены:

- `unityaid-front/src/entities/profileFields/types.ts`
- `unityaid-front/src/entities/profileFields/api.ts`
- `unityaid-front/src/pages/ProfileFieldsSettingsPage.vue`

Изменены:

- `unityaid-front/src/pages/ProfilePage.vue`
- `unityaid-front/src/pages/SettingsPage.vue`
- `unityaid-front/src/router/index.ts`
- `unityaid-front/src/assets/main.css`

Новый маршрут настроек:

- `/settings/profile-fields/manage`

Профиль стал вкладочным:

- «Личная информация»
- «Задачи»
- «Работа»
- «Мероприятия»
- «Достижения и документы»

В профиле:

- можно загрузить аватар;
- можно сохранить базовые данные;
- вкладка «Работа» показывает кастомные группы/поля;
- фактически проверено поле «Готовность к выездам».

В настройках профиля:

- отображаются системные группы с защитой;
- можно создать группу;
- можно создать поле;
- можно редактировать поля/группы;
- пользовательские поля можно удалить;
- системные удалить нельзя.

## 4. GitHub Actions

Добавлен workflow:

- `.github/workflows/quality.yml`

Jobs:

- `backend`: `go mod download`, `go vet ./...`, `go test ./...`, `go build ./...`
- `frontend`: Node.js 22, `npm ci`, `npm run build`
- `containers`: `docker compose config --quiet`, build backend/frontend, `docker compose up -d --wait`, health checks

Локально `docker compose config --quiet` проходил.

Важно: контейнерный frontend на Windows/Docker ранее падал с `SIGBUS` при запуске Vite. Для визуальной проверки frontend запускался локально через `npm run dev`. Production build frontend проходит.

## 5. Фактические проверки

Успешно выполнялись:

```powershell
cd E:\diploma\unityaid-back
go test ./...
go vet ./...
go build ./...
```

`go test ./...` успешен, но почти все пакеты показывают `[no test files]`.

```powershell
cd E:\diploma\unityaid-front
npm.cmd run build
```

Frontend build успешен.

```powershell
cd E:\diploma
docker compose config --quiet
```

Compose config успешен.

Проверенный API-сценарий V4:

- применена миграция `020_profile_fields.sql`;
- созданы системные группы;
- создана группа «Подготовка волонтера»;
- создано поле «Готовность к выездам»;
- волонтер сохранил значение `ready`;
- удаление системного поля вернуло HTTP `409`;
- загрузка PNG через `/files/profile-avatars` вернула URL `/uploads/avatars/...png`.

## 6. Скриншоты V4

Созданы:

- `V4_generated/screenshots/09_profile_fields_admin.png`
- `V4_generated/screenshots/10_profile_work_custom_fields.png`

Также в `V4_generated/screenshots` скопированы V3-скриншоты `01`-`08`.

Эти изображения используются в V4:

- Рисунок 25 — конструктор групп и полей профиля.
- Рисунок 26 — вкладка работы волонтера с настраиваемым полем.
- Приложение Д — рисунки Д.9 и Д.10.

## 7. Статус DOCX V4

V4 уже собиралась, но при визуальной проверке был найден дефект:

- страница содержания была пустой, потому что LibreOffice не развернул автоматическое поле TOC.

В `build_diploma_v4_puls.py` уже добавлен `STATIC_CONTENTS`, который должен заменить автоматическое содержание на статическое.

Следующий агент должен:

1. Пересобрать V4.
2. Отрендерить V4 в PDF через LibreOffice.
3. Растеризовать PDF в PNG/contact sheets.
4. Проверить все страницы визуально.
5. Если дефектов нет — отдать пользователю финальный DOCX.

Структурная проверка до исправления содержания показывала:

- новые разделы есть;
- `GitHub Actions` есть;
- `Готовность к выездам` есть;
- `HTTP 409` есть;
- `Рисунок Д.10` есть;
- комментариев нет;
- `w:highlight` нет;
- tracked changes нет;
- после рендера было около 122 страниц;
- таблиц около 35.

## 8. Команды для V4

Пересборка V4:

```powershell
& 'C:\Users\Павел\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe' `
  'E:\diploma\docs\Документ диплома\Финальный файлы диплома Зозин\build_diploma_v4_puls.py'
```

Рендер через LibreOffice:

```powershell
New-Item -ItemType Directory -Force 'E:\diploma\.codex\v4_render' | Out-Null

Copy-Item `
  -LiteralPath 'E:\diploma\docs\Документ диплома\Финальный файлы диплома Зозин\Zozin_Puls_VKR_final_V4_dorabotannaya.docx' `
  -Destination 'E:\diploma\.codex\Zozin_Puls_V4.docx' `
  -Force

& 'C:\Program Files\LibreOffice\program\soffice.com' `
  --headless `
  --convert-to pdf `
  --outdir 'E:\diploma\.codex\v4_render' `
  'E:\diploma\.codex\Zozin_Puls_V4.docx'
```

Предыдущий PDF:

- `E:\diploma\.codex\v4_render\Zozin_Puls_V4.pdf`

Предыдущие contact sheets:

- `E:\diploma\.codex\v4_render\contact-001-020.png`
- `contact-021-040.png`
- ...

## 9. Проблема DBeaver / PostgreSQL

Пользователь сообщил, что не может подключиться к базе в DBeaver.

На скрине DBeaver стоял порт `5432`.

В `.env` проекта сейчас:

```env
POSTGRES_USER=unityaid
POSTGRES_PASSWORD=unityaid
POSTGRES_DB=unityaid
POSTGRES_PORT=5433
```

В `docker-compose.yml`:

```yaml
ports:
  - "${POSTGRES_PORT:-5432}:5432"
```

Значит DBeaver должен подключаться к:

- Host: `localhost`
- Port: `5433`
- Database: `unityaid`
- User: `unityaid`
- Password: `unityaid`

Если `docker ps` показывает только `5432/tcp` без `0.0.0.0:5433->5432/tcp`, порт на хост не опубликован. Нужно пересоздать контейнер:

```powershell
cd E:\diploma
docker compose down
docker compose up -d db
docker ps --filter "name=unityaid-db" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

Ожидаемо:

```text
0.0.0.0:5433->5432/tcp
```

Если порт `5433` занят, можно поменять `.env`:

```env
POSTGRES_PORT=55432
```

и пересоздать контейнер.

## 10. Команды для приложения

Backend:

```powershell
cd E:\diploma\unityaid-back
go test ./...
go vet ./...
go build ./...
```

Frontend:

```powershell
cd E:\diploma\unityaid-front
npm.cmd run build
```

Поднять db/backend:

```powershell
cd E:\diploma
docker compose up -d db backend
```

Запустить frontend локально:

```powershell
cd E:\diploma\unityaid-front
npm.cmd run dev -- --host 127.0.0.1 --port 5173
```

## 11. Что делать дальше

Приоритет:

1. Помочь пользователю с DBeaver: скорее всего нужен порт `5433`, а не `5432`.
2. Пересобрать V4 после `STATIC_CONTENTS`.
3. Повторить render/visual QA всех страниц.
4. Если V4 чистая — финально отдать ссылку на DOCX и результаты проверок.
5. Не откатывать dirty worktree: там много уже сделанных изменений по ребрендингу, профилю и V4.

## 12. Осторожность

Не удалять и не перезаписывать:

- V1
- V2
- V3

V4 можно пересобирать.

Не выполнять `git reset --hard` и не откатывать чужие изменения.

