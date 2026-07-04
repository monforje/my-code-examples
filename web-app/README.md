# web-app

`web-app` - frontend Codurity на React и TypeScript. Он собирает пользовательский сценарий: регистрация и вход, CLI login confirmation, профиль, настройки безопасности, список задач, детальная страница задачи, sandbox и просмотр отчетов.

## Возможности

- Auth screens: login, register, verify, forgot/reset password.
- CLI login page для device flow.
- Protected routes через общий auth provider.
- Профиль, редактирование профиля и security settings.
- Каталог задач с фильтрами, тегами и level/type UI.
- Детальная страница задачи.
- Sandbox и просмотр task reports.
- Polling pending-отчетов до финального статуса.
- Typed API client, сгенерированный по OpenAPI.

## Основной стек

- React 19
- TypeScript
- Vite
- React Router
- TanStack Query
- Orval для API generation
- oxlint/oxfmt
- Bun

## Маршруты

- `/login`
- `/register`
- `/verify`
- `/forgot`
- `/forgot/verify`
- `/reset`
- `/cli/login`
- `/profile`
- `/settings/profile`
- `/settings/security`
- `/tasks`
- `/tasks/:taskId`
- `/sandbox`
- `/sandbox/tasks`
- `/sandbox/tasks/reports/:reportId`

## Структура

Проект организован в feature-sliced стиле:

- `src/app` - root app, providers и routes.
- `src/pages` - page-level screens.
- `src/widgets` - крупные layout/header блоки.
- `src/features` - пользовательские сценарии.
- `src/entities` - доменные сущности: task, report, user.
- `src/shared` - API, UI-kit, assets, config и utilities.
- `openapi` - OpenAPI specs backend-сервисов.

## Локальный запуск

```bash
bun install
bun run dev
```

Через Taskfile/Docker:

```bash
task create-network
task deploy
```

Полезные команды:

```bash
bun run build
bun run lint
bun run fmt:check
task orval
task stop
```
