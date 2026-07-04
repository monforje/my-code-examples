# Codurity Design System

## 1. Идея

Codurity — это dev hire platform. Интерфейс должен ощущаться как технический продукт: строгий, быстрый, понятный, но не слишком нишевый.

Основной визуальный принцип:

````txt
чистый SaaS + лёгкая графичная жёсткость

Система строится на трёх вещах:

Чёрно-белая база.
Синий брендовый цвет #1758d1.
Заметные бордеры, но без агрессивного brutalist-перегруза.
2. Основной цвет
--cd-blue-700: #1758d1;

Полная шкала:

--cd-blue-50:  #e8f3ff;
--cd-blue-100: #d3e2fd;
--cd-blue-200: #a5c1f5;
--cd-blue-300: #759fee;
--cd-blue-400: #4d82e9;
--cd-blue-500: #3470e6;
--cd-blue-600: #2566e5;
--cd-blue-700: #1758d1;
--cd-blue-800: #0b4cb8;
--cd-blue-900: #0041a3;
Как использовать

blue-700 — основной action-цвет.

Использовать для:

primary-кнопок;
активных состояний;
важных badge;
прогресса;
ключевых метрик;
ссылок;
focus-состояний.

Не использовать синий для всего подряд. В интерфейсе он должен быть акцентом, а не фоном всего продукта.

3. Темы
Light theme

База:

фон: белый
текст: чёрный
акцент: синий

Использовать для основного UI, таблиц, форм, страниц с большим количеством данных.

Dark theme

База:

фон: почти чёрный
текст: белый
акцент: синий

Использовать для code review, editor-like страниц, дашбордов, технических экранов.

4. Шрифты
Body
--cd-font-body: Onest;

Используется для:

обычного текста;
форм;
таблиц;
кнопок;
навигации;
карточек.

Onest хорошо работает на русском и выглядит современно.

Headings
--cd-font-heading: Manrope;

Используется для:

h1-h6;
больших чисел;
ключевых метрик;
заголовков карточек;
hero-секций.

Manrope даёт мягкость, которая хорошо сочетается с округлым логотипом Codurity.

Mono
--cd-font-mono: JetBrains Mono;

Используется для:

кода;
id;
API-ответов;
логов;
технических значений.
5. Логотип

Логотип Codurity использует отдельный шрифт:

Geometry Soft Pro Bold A

Важно:

не использовать шрифт логотипа в интерфейсе;
не пытаться сделать весь UI таким же мягким, как лого;
интерфейс должен поддерживать лого, а не копировать его.

Лого мягкое, поэтому UI компенсирует это чёткой сеткой, бордерами и строгой типографикой.

6. Радиусы

Базовые радиусы:

--cd-radius-xs: 6px;
--cd-radius-sm: 10px;
--cd-radius-md: 14px;
--cd-radius-lg: 22px;

Рекомендация:

6px — badge, маленькие элементы;
10px — кнопки, inputs;
14px — карточки, таблицы;
22px — большие панели.

Не делать полностью квадратный UI. Это ломает связь с мягким логотипом.

7. Бордеры

Бордеры — важная часть стиля.

Базовый бордер:

border: 1.5px solid var(--cd-border);

Использовать для:

карточек;
кнопок;
input;
таблиц;
code-block;
крупных layout-блоков.

Мягкие бордеры:

border-color: var(--cd-border-soft);

Использовать внутри карточек:

разделители строк;
вторичные карточки;
вложенные блоки.
8. Тени

Есть два типа теней.

Hard shadow
box-shadow: var(--cd-shadow-hard);

Использовать редко:

важный preview-блок;
главная карточка;
промо-блок;
выделенный экран.
Soft shadow
box-shadow: var(--cd-shadow-soft);

Использовать для обычных elevated-блоков.

Не ставить hard-shadow на каждую карточку. Иначе интерфейс станет слишком brutalist.

9. Кнопки
Primary

Для главного действия:

<button class="cd-button cd-button-primary">
  Открыть отчёт
</button>
Default

Для вторичных действий:

<button class="cd-button">
  Сравнить кандидатов
</button>
Soft

Для мягкого синего action:

<button class="cd-button cd-button-soft">
  Назначить интервью
</button>
Ghost

Для навигации и лёгких действий:

<button class="cd-button cd-button-ghost">
  Отмена
</button>
10. Карточки
Обычная карточка
<div class="cd-card">
  Контент
</div>
Мягкая карточка
<div class="cd-card-soft">
  Контент
</div>
Акцентная карточка
<div class="cd-card-hard">
  Контент
</div>

Использовать cd-card-hard только для визуально важных блоков.

11. Формы
<div class="cd-field">
  <label class="cd-label">Email кандидата</label>
  <input class="cd-input" placeholder="anna@example.com" />
  <div class="cd-help-text">На этот email придёт приглашение.</div>
</div>

Правила:

label всегда жирный;
input всегда с чётким бордером;
focus всегда с синим ring;
placeholder приглушённый.
12. Таблицы
<div class="cd-table-wrap">
  <table class="cd-table">
    <thead>
      <tr>
        <th>Кандидат</th>
        <th>Роль</th>
        <th>Балл</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Анна Смирнова</td>
        <td>Backend Go</td>
        <td>84</td>
      </tr>
    </tbody>
  </table>
</div>

Таблицы должны быть спокойными. Не перегружать цветом.

13. Badge
<span class="cd-badge">Hire</span>
<span class="cd-badge cd-badge-success">Пройдено</span>
<span class="cd-badge cd-badge-warning">Ревью</span>
<span class="cd-badge cd-badge-danger">Ошибка</span>

Badge должен быть компактным. Не использовать большие pill-элементы везде.

14. Code block
<pre class="cd-code">func ValidateEmail(email string) error {
    if email == "" {
        return ErrEmptyEmail
    }

    return nil
}</pre>

Правила:

код всегда в JetBrains Mono;
фон кода тёмный даже в light theme;
code block должен иметь чёткий border.
15. Метрики
<div class="cd-metric">
  <span class="cd-metric-label">Средний балл</span>
  <strong class="cd-metric-value">84</strong>
</div>

Метрики используют heading-шрифт, потому что это ключевые данные.

16. Layout

Использовать простые layout-primitives:

<div class="cd-container">
  <section class="cd-section">
    <div class="cd-grid-3">
      ...
    </div>
  </section>
</div>

Доступные классы:

.cd-container
.cd-page
.cd-section
.cd-stack
.cd-stack-sm
.cd-stack-lg
.cd-row
.cd-row-between
.cd-grid
.cd-grid-2
.cd-grid-3
.cd-grid-4
17. Пример страницы
<main class="cd-page">
  <section class="cd-section">
    <div class="cd-container cd-stack-lg">

      <div class="cd-row-between">
        <div>
          <span class="cd-badge">Backend Go</span>
          <h1>Техническое ревью кандидата</h1>
          <p>
            Система оценивает решение по тестам, читаемости,
            архитектуре и устойчивости.
          </p>
        </div>

        <button class="cd-button cd-button-primary">
          Открыть отчёт
        </button>
      </div>

      <div class="cd-grid-3">
        <div class="cd-card">
          <div class="cd-metric">
            <span class="cd-metric-label">Средний балл</span>
            <strong class="cd-metric-value">84</strong>
          </div>
        </div>

        <div class="cd-card">
          <div class="cd-metric">
            <span class="cd-metric-label">Решено задач</span>
            <strong class="cd-metric-value">7 / 9</strong>
          </div>
        </div>

        <div class="cd-card">
          <div class="cd-metric">
            <span class="cd-metric-label">Покрытие тестами</span>
            <strong class="cd-metric-value">86%</strong>
          </div>
        </div>
      </div>

      <pre class="cd-code">func ValidateEmail(email string) error {
    if email == "" {
        return ErrEmptyEmail
    }

    return nil
}</pre>

    </div>
  </section>
</main>
18. Рекомендации
Делать
использовать белый/чёрный как основу;
использовать синий для действий;
держать UI чистым и техническим;
использовать бордеры как часть визуального языка;
оставлять достаточно воздуха;
использовать hard-shadow редко.
Не делать
не заливать весь интерфейс синим;
не делать слишком brutalist;
не использовать шрифт логотипа в UI;
не делать все карточки с hard-shadow;
не использовать много разных цветов статусов;
не делать огромные скругления везде.

---

Минимальный пример подключения:

```html
<!doctype html>
<html lang="ru" data-theme="light">
<head>
  <meta charset="UTF-8" />
  <title>Codurity</title>

  <link rel="stylesheet" href="./styles/tokens.css" />
  <link rel="stylesheet" href="./styles/themes.css" />
  <link rel="stylesheet" href="./styles/globals.css" />
</head>

<body>
  <main class="cd-page">
    <section class="cd-section">
      <div class="cd-container cd-stack-lg">
        <div class="cd-row-between">
          <div class="cd-stack">
            <span class="cd-badge">Backend Go</span>
            <h1>Техническое ревью кандидата</h1>
            <p>Оценка решения по тестам, архитектуре и качеству кода.</p>
          </div>

          <button class="cd-button cd-button-primary">
            Открыть отчёт
          </button>
        </div>

        <div class="cd-grid-3">
          <div class="cd-card">
            <div class="cd-metric">
              <span class="cd-metric-label">Средний балл</span>
              <strong class="cd-metric-value">84</strong>
            </div>
          </div>

          <div class="cd-card">
            <div class="cd-metric">
              <span class="cd-metric-label">Решено задач</span>
              <strong class="cd-metric-value">7 / 9</strong>
            </div>
          </div>

          <div class="cd-card">
            <div class="cd-metric">
              <span class="cd-metric-label">Покрытие тестами</span>
              <strong class="cd-metric-value">86%</strong>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</body>
</html>
````
