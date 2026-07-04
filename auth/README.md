# auth

`auth` - сервис аутентификации и управления identity для Codurity. Он отвечает за регистрацию, логин, refresh-сессии, текущего пользователя, смену email/пароля, удаление аккаунта и device authorization flow для CLI.

## Возможности

- Регистрация пользователя с кодом подтверждения email.
- Login/logout и refresh token flow.
- Получение текущей identity через `/auth/me`.
- Смена email с подтверждением старого и нового адреса.
- Смена и восстановление пароля через verification codes.
- Удаление аккаунта через email confirmation.
- CLI device flow: start, confirm, token и refresh.
- Публикация событий в NATS и Kafka.

## Основной стек

- Go 1.26.4
- Echo HTTP server
- PostgreSQL + Goose migrations
- Redis для verification/rate-limit state
- NATS и Kafka для событий
- TypeSpec/OpenAPI + oapi-codegen
- Testcontainers для интеграционных тестов

## API

Контракт лежит в `openapi/openapi.yaml`, исходники контракта - в `typespec`.

Основные группы endpoints:

- `/auth/register`, `/auth/register/verify`, `/auth/register/code/resend`
- `/auth/login`, `/auth/logout`, `/auth/refresh`
- `/auth/me`
- `/auth/password/change/*`, `/auth/password/forgot/*`, `/auth/password/reset`
- `/auth/me/email/change/*`
- `/auth/me/delete/*`
- `/auth/device/start`, `/auth/device/confirm`, `/auth/device/token`
- `/auth/cli/refresh`

## События

Сервис публикует identity events и email-code events:

- `identity.created`
- `identity.updated`
- `identity.deleted`
- `identity.login`
- `identity.logout`
- `notification.email.verification_code.send`
- `notification.email.password_reset_code.send`
- `notification.email.password_change_code.send`
- `notification.email.email_change_code.send`
- `notification.email.account_delete_code.send`

## Структура

- `cmd/auth` - entrypoint HTTP-сервиса.
- `internal/services/auth` - бизнес-логика.
- `internal/http/handlers` - HTTP handlers.
- `internal/http/validation` - request validation.
- `internal/repository/postgres` - persistent storage.
- `internal/repository/redis` - cache/rate-limit state.
- `internal/repository/nats` и `internal/repository/kafka` - event publishers.
- `migrations` - SQL migrations.
- `tests/e2e` - end-to-end tests.

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

Перед использованием вне локального окружения замените sample secrets из `.env.example`.
