# AGENTS.md — Codurity Site App

## Стек

- React 19 + TypeScript 6 + Vite 8
- Bun (пакетный менеджер)
- TanStack Query (серверный стейт)
- React Router (маршрутизация, Library Mode)
- Zod (валидация)
- Orval (генерация API-клиентов из OpenAPI)
- oxlint + oxfmt (линтер + форматтер вместо ESLint/Prettier)
- @iconify/react (иконки)
- CSS Modules / глобальные стили, без препроцессоров

## Запуск

```bash
bun install          # установка зависимостей
bun run dev          # dev-сервер
bun run build        # production-сборка (tsc + vite build)
bun run lint         # oxlint
bun run lint:fix     # oxlint --fix
bun run fmt          # oxfmt (форматирование)
bun run fmt:check    # проверка форматирования
```

## Структура проекта (FSD)

Проект построен по архитектуре [Feature-Sliced Design](https://fsd.how/ru/).

```
src/
├── app/                  # Инициализация приложения
│   ├── index.tsx         # App — маршруты (Routes)
│   ├── providers.tsx     # QueryClientProvider + BrowserRouter
│   └── styles.css        # Импорт дизайн-системы
│
├── pages/                # Страницы (каждая страница — слайс)
│   └── home/
│       ├── index.ts      # Публичный API страницы
│       └── ui/
│           └── HomePage.tsx
│
├── widgets/              # Крупные переиспользуемые блоки UI
│   └── header/
│       ├── index.ts
│       └── ui/
│           └── Header.tsx
│
├── features/             # Пользовательские действия (авторизация, фильтры и т.д.)
│   └── auth/
│       ├── index.ts
│       ├── ui/
│       ├── model/
│       └── api/
│
├── entities/             # Бизнес-сущности (User, Post и т.д.)
│   └── user/
│       ├── index.ts
│       ├── ui/
│       └── model/
│
└── shared/               # Общий код, переиспользуемый везде
    ├── api/              # API-клиент (orval), типы
    ├── assets/           # Лого, шрифты
    ├── config/           # Дизайн-система (CSS-токены)
    ├── lib/              # Утилиты, хелперы
    ├── router/           # Конфигурация роутинга
    └── ui/               # Общие UI-компоненты (кнопки, инпуты)
```

### Слои FSD (сверху вниз)

| Слой       | Назначение                                      |
|------------|------------------------------------------------|
| **app**    | Инициализация: провайдеры, роутинг, глобальные стили |
| **pages**  | Страницы приложения                             |
| **widgets**| Крупные переиспользуемые блоки UI               |
| **features**| Действия пользователя (логин, поиск и т.д.)   |
| **entities**| Бизнес-сущности (User, Post)                   |
| **shared** | Общий код: UI-кит, API, утилиты, конфиги       |

### Правило импортов

Модуль может импортировать только из слоёв **ниже**:
- `pages` → `widgets`, `features`, `entities`, `shared`
- `widgets` → `features`, `entities`, `shared`
- `features` → `entities`, `shared`
- `entities` → `shared`
- `shared` → ничего (нижний слой)

Запрещено: импорт между страницами, импорт с нижних слоёв на верхние.

### Сегменты внутри слайсов

- `ui/` — React-компоненты
- `model/` — бизнес-логика, типы, zod-схемы, состояние
- `api/` — запросы к бэкенду, orval-хуки
- `lib/` — утилиты
- `config/` — конфигурация, переменные окружения

### Публичный API

Каждый слайс/сегмент экспортирует через `index.ts`. Внутренняя структура папки — приватная.

```ts
// pages/home/index.ts
export { HomePage } from './ui/HomePage'
```

Импорт всегда через слайс:
```ts
import { HomePage } from '@pages/home'  // ✅
import { HomePage } from '@pages/home/ui/HomePage'  // ❌
```

## Алиасы

Настроены в `vite.config.ts` и `tsconfig.app.json`:

```
@app/*      → src/app/*
@pages/*    → src/pages/*
@widgets/*  → src/widgets/*
@features/* → src/features/*
@entities/* → src/entities/*
@shared/*   → src/shared/*
```

## Дизайн-система

Файлы лежат в `src/shared/config/design-system/`:

| Файл        | Содержимое                                    |
|-------------|-----------------------------------------------|
| `tokens.css`| Базовые токены: палитра, шрифты, размеры, радиусы, тени |
| `themes.css`| Семантические токены для light/dark тем       |
| `globals.css`| CSS-сброс, типографика, layout-примитивы, компоненты |

### Подключение

Импортируются через `src/app/styles.css`:

```css
@import '../shared/config/design-system/tokens.css';
@import '../shared/config/design-system/themes.css';
@import '../shared/config/design-system/globals.css';
```

### Токены

Основные цвета:
- `--cd-primary: #1758d1` (синий брендовый)
- `--cd-bg`, `--cd-surface`, `--cd-text`, `--cd-border`

Шрифты:
- `--cd-font-body: Onest` — основной текст
- `--cd-font-heading: Manrope` — заголовки
- `--cd-font-mono: JetBrains Mono` — код

### Компоненты (CSS-классы)

Кнопки:
```html
<button class="cd-button">Default</button>
<button class="cd-button cd-button-primary">Primary</button>
<button class="cd-button cd-button-soft">Soft</button>
<button class="cd-button cd-button-ghost">Ghost</button>
```

Карточки:
```html
<div class="cd-card">Обычная</div>
<div class="cd-card-soft">Мягкая</div>
<div class="cd-card-hard">Акцентная</div>
```

Формы:
```html
<div class="cd-field">
  <label class="cd-label">Email</label>
  <input class="cd-input" placeholder="anna@example.com" />
  <div class="cd-help-text">Подсказка</div>
</div>
```

Layout:
```html
<div class="cd-container">Обёртка</div>
<section class="cd-section">Секция</section>
<div class="cd-grid-3">Сетка 3 колонки</div>
<div class="cd-stack-lg">Стек с большим gap</div>
```

Темы:
```html
<html data-theme="light"> <!-- или "dark" -->
```

### Дизайн-спецификация

Полное описание системы: `src/shared/config/design-system/design-system.md`

## Когда добавлять новый код

1. **Новая страница** → `src/pages/<name>/` с сегментами `ui/`, `model/`, `api/`
2. **Новая фича** → `src/features/<name>/` (действие пользователя)
3. **Новая сущность** → `src/entities/<name>/` (модель данных)
4. **Переиспользуемый UI** → `src/shared/ui/`
5. **API-клиент** → `src/shared/api/` (через orval)
6. **Утилита** → `src/shared/lib/`

## Orval (API-клиент)

Для генерации API-клиента из OpenAPI-спецификации:

```bash
bunx orval           # генерация из orval.config.ts
```

Конфигурация: `orval.config.ts` в корне проекта.
Сгенерированные файлы: `src/shared/api/`.

## Иконки (Iconify)

Пакет: `@iconify/react`

### Использование

```tsx
import { Icon } from '@iconify/react'

// По названию иконки (ленивая загрузка)
<Icon icon="mdi:home" />

// С кастомизацией
<Icon icon="mdi:home" width={24} color="var(--cd-primary)" />

// Конкретный набор (без ленивой загрузки)
import HomeIcon from '@iconify-icons/mdi/home'
<Icon icon={HomeIcon} />
```

### Популярные наборы иконок

| Набор       | Префикс    | Описание                    |
|-------------|------------|-----------------------------|
| Material Design Icons | `mdi:` | Material Design |
| Heroicons    | `heroicons:` | Tailwind-стиль           |
| Lucide       | `lucide:`  | Минималистичные              |
| Phosphor     | `phosphor:`| Универсальные                |
| Carbon       | `carbon:`  | IBM Carbon                   |

### Поиск иконок

https://icon-sets.iconify.design/ — каталог всех иконок с поиском.
