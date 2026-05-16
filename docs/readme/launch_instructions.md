# Инструкция запуска UnityAid

## Docker

1. Скопировать `.env.example` в `.env` и проверить параметры PostgreSQL, JWT и upload-директории.
2. Запустить инфраструктуру:

```bash
docker compose up --build
```

3. Backend доступен на `http://localhost:8080`.
4. Frontend доступен на `http://localhost:5173`.

## Локальный запуск backend

```bash
cd unityaid-back
go run ./cmd/api
```

Перед запуском должен быть доступен PostgreSQL, а миграции должны быть применены.

## Локальный запуск frontend

```bash
cd unityaid-front
npm install
npm run dev
```

## Проверка

```bash
cd unityaid-back
go test ./...

cd ../unityaid-front
npm run build
```
