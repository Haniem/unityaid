# UnityAid API Overview

Базовый путь API: `/api/v1`.

## Авторизация

- `POST /auth/register` - регистрация.
- `POST /auth/login` - вход.
- `POST /auth/refresh` - обновление сессии.
- `GET /auth/me` - текущий пользователь.
- `POST /auth/logout` - выход.
- `POST /auth/forgot-password` - запрос сброса пароля.
- `POST /auth/reset-password` - установка нового пароля.
- `POST /auth/verify-email` - подтверждение email.
- `POST /auth/change-password` - смена пароля.

## Основные модули

- `/users`, `/volunteers`, `/skills`, `/system-roles` - пользователи, профили, навыки, системные роли.
- `/organizations` - организации и участники организаций.
- `/events` - мероприятия, заявки, посещаемость, смены, обратная связь.
- `/tasks` - задачи, назначения, комментарии, вложения, история статусов, учет часов.
- `/time-entries` - учет и подтверждение волонтерских часов.
- `/news` - новости, категории, публикации и черновики.
- `/gamification`, `/achievements` - достижения, баллы, лидерборд.
- `/analytics` - отчеты по волонтерам, мероприятиям, задачам, геймификации и действиям пользователей.
- `/notifications` - in-app уведомления.
- `/knowledge-base` - база знаний и категории.
- `/certificates` - сертификаты и справки.
- `/admin` - системная админ-панель.

## Сертификаты

- `GET /certificates` - список сертификатов текущего пользователя или всех доступных менеджеру.
- `POST /certificates/generate` - генерация сертификата или справки.
- `GET /certificates/download/{id}` - скачивание PDF.
- `GET /certificates/verify/{code}` - публичная проверка по коду.

## Системная админ-панель

- `GET /admin/entities` - список разрешенных сущностей.
- `GET /admin/{entity}?page=1&perPage=20&search=text` - постраничный список.
- `POST /admin/{entity}` - создание записи.
- `PUT /admin/{entity}/{id}` - редактирование записи.
- `DELETE /admin/{entity}/{id}` - удаление записи.

Доступ к `/admin` есть только у супер-администратора приложения.
