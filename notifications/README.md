# notifications

`notifications` - сервис уведомлений и отчетов. Он состоит из HTTP API и worker-процесса: HTTP-часть отдает report data и принимает webhook, worker слушает NATS-события для email-кодов и отправляет письма через SMTP.

## Возможности

- Отправка verification/reset/change/delete email codes.
- Идемпотентная обработка NATS events через таблицу processed events.
- HTML, text и MJML email templates.
- Получение report данных по API.
- Webhook endpoint для внешних интеграций.
- Кэширование CI/report данных в Redis.

## Основной стек

- Go 1.26.4
- Echo HTTP server
- PostgreSQL + Goose migrations
- Redis
- NATS
- SMTP через `go-mail`
- TypeSpec/OpenAPI + oapi-codegen
- Testcontainers

## API

Контракт лежит в `openapi/openapi.yaml`.

Endpoints:

- `/reports/{uid}`
- `/webhook`

## События

Worker подписывается на:

- `notification.email.verification_code.send`
- `notification.email.password_reset_code.send`
- `notification.email.password_change_code.send`
- `notification.email.email_change_code.send`
- `notification.email.account_delete_code.send`

Каждое событие проходит idempotency check перед отправкой письма.

## Структура

- `cmd/notifications-http` - HTTP entrypoint.
- `cmd/notifications-worker` - worker entrypoint.
- `internal/http` - HTTP handlers, middleware, validation.
- `internal/worker/consumer` - NATS consumer.
- `internal/templates` - email templates.
- `internal/repository/mailer` - отправка email.
- `internal/repository/postgres` - reports и processed events.
- `internal/services/ci_report` и `internal/services/report` - бизнес-логика отчетов.
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
task codegen
task mocks
task go -- test ./...
task stop
```

Для реальной отправки писем нужно заменить SMTP-переменные в `.env`.
