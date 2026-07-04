# codurity

`codurity` - CLI-клиент для Codurity. Он связывает пользователя с backend-платформой: логинится через device flow, хранит локальные credentials, получает задачи и помогает клонировать git-репозитории.

## Возможности

- Device login через `auth` service.
- Локальное хранение token/config.
- Проверка статуса авторизации.
- Получение задачи по имени.
- Клонирование GitHub-репозитория.
- Вывод версии сборки.

## Основной стек

- Go 1.26.4
- Cobra для CLI-команд
- Viper для конфигурации
- `pkg/browser` для открытия browser flow

## Команды

```bash
codurity auth login
codurity auth status
codurity auth logout
codurity get <task_name>
codurity clone <owner/repo>
codurity version
```

## Сборка

CLI получает build-time значения через `ldflags`: `Version`, `AuthAPIURL`, `FrontendURL`.

```bash
cp .env.example .env
task build
./codurity version
```

Запуск без отдельной сборки:

```bash
task run -- version
task run -- auth status
```

## Структура

- `cmd` - Cobra commands.
- `internal/api` - HTTP-клиент backend API.
- `internal/auth` - auth/device flow.
- `internal/backend` - backend endpoint orchestration.
- `internal/config` - локальная конфигурация.
- `internal/tasks` - получение и подготовка задач.
- `internal/token` - хранение access/refresh данных.

## Проверки

```bash
task test
task vet
task tidy
```
