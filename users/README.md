# users

`users` - сервис профиля пользователя. Он хранит пользовательские данные поверх identity из `auth`: профиль, username, avatar, настройки и публичные данные для git/task сценариев.

## Возможности

- Получение профиля текущего пользователя.
- Обновление профильных настроек.
- Загрузка, хранение и раздача аватаров.
- Получение git-user данных для внутренних интеграций.
- Worker-синхронизация профиля по identity events.
- Идемпотентная обработка событий через processed events.

## Основной стек

- Go 1.26.4
- Echo HTTP server
- PostgreSQL + Goose migrations
- NATS
- TypeSpec/OpenAPI + oapi-codegen
- Testcontainers

## API

Контракт лежит в `openapi/openapi.yaml`.

Endpoints:

- `/profile/me`
- `/profile/me/avatar`
- `/profile/me/settings`
- `/git-user/me`

## События

Worker подписывается на identity events из `auth`:

- `identity.created`
- `identity.updated`
- `identity.deleted`

На основе этих событий сервис создает, обновляет или удаляет профильные записи.

## Структура

- `cmd/users-server` - HTTP entrypoint.
- `cmd/users-worker` - worker entrypoint.
- `internal/services` - бизнес-логика пользователей.
- `internal/worker/consumer` - NATS consumer.
- `internal/generate/avatar` и `internal/generate/username` - генерация дефолтных данных.
- `internal/repository/postgres` - persistent storage.
- `internal/repository/storage` - avatar storage.
- `internal/http` - HTTP слой.
- `migrations` - SQL migrations.

## Локальный запуск

```bash
cp .env.example .env
task create-network
task deploy
task goose -- up
```

Полезные команды:

```bash
task go -- test ./...
task stop
```

Для production-like окружения замените `SERVICE_TOKEN`, JWT secret и internal client tokens.
