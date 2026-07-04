# tasks

`tasks` - сервис каталога заданий и отчетов. Он отдает список задач, фильтры, детальную карточку, git metadata для CLI и принимает report data от внешнего runner/CI.

## Возможности

- Список задач с фильтрами и пагинацией.
- Детальная карточка задачи.
- Справочники языков и тегов.
- Выдача git-информации по имени задачи.
- Создание и чтение CI/report результатов.
- Валидация user auth через JWT.
- Интеграция с внешним taskrunner/git service.

## Основной стек

- Go 1.26.4
- Echo HTTP server
- PostgreSQL + Goose migrations
- TypeSpec/OpenAPI + oapi-codegen
- Testcontainers

## API

Контракт лежит в `openapi/openapi.yaml`.

Endpoints:

- `/tasks`
- `/tasks/{task_id}`
- `/tasks/{task_name}/git`
- `/tasks/languages`
- `/tasks/tags`
- `/reports`
- `/reports/{reportId}`

## Структура

- `cmd/tasks-http` - HTTP entrypoint.
- `internal/services/git_tasks` - логика git/task metadata.
- `internal/services/reports` - логика reports.
- `internal/repository/postgres` - tasks, tags, reports и связанные записи.
- `internal/http/handlers` - HTTP handlers.
- `internal/http/middleware` - auth и request middleware.
- `pkg/http_clients` - клиенты внешних сервисов.
- `scripts` - вспомогательная post-processing логика для OpenAPI.
- `test_data` - fixtures для проверок.
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
task go -- test ./...
task stop
```

Sample tokens из `.env.example` подходят только для локальной разработки.
