export const TYPE_CONTENT: Record<
  string,
  { label: string; title: string; copy: string; checklist: string[] }
> = {
  backend: {
    label: "Backend",
    title: "Серверная задача",
    copy: "Собери решение вокруг контракта, обработки ошибок и устойчивого поведения по ТЗ.",
    checklist: [
      "Проверь входные и выходные контракты",
      "Обработай ошибки и пограничные случаи",
      "Сохрани логику тестируемой и изолированной",
    ],
  },
  frontend: {
    label: "Frontend",
    title: "Frontend-задача",
    copy: "Собери интерфейс вокруг состояния, сценариев взаимодействия и аккуратной обработки edge cases.",
    checklist: [
      "Проверь состояния загрузки, ошибки и пустые данные",
      "Сохрани доступность базовых элементов",
      "Раздели UI, состояние и побочные эффекты",
    ],
  },
};

export const LANGUAGE_CONTENT: Record<string, { label: string; checklist: string[] }> = {
  typescript: {
    label: "TypeScript",
    checklist: [
      "Опиши типы входов и результата",
      "Избегай any в ключевой логике",
      "Обработай невалидные состояния",
    ],
  },
  ts: {
    label: "TypeScript",
    checklist: [
      "Опиши типы входов и результата",
      "Избегай any в ключевой логике",
      "Обработай невалидные состояния",
    ],
  },
  javascript: {
    label: "JavaScript",
    checklist: [
      "Проверь асинхронные сценарии",
      "Не смешивай бизнес-логику и представление",
      "Обработай невалидные входные данные",
    ],
  },
  js: {
    label: "JavaScript",
    checklist: [
      "Проверь асинхронные сценарии",
      "Не смешивай бизнес-логику и представление",
      "Обработай невалидные входные данные",
    ],
  },
  go: {
    label: "Go",
    checklist: [
      "Проверь работу с context и ошибками",
      "Держи интерфейсы небольшими",
      "Добавь тесты для ключевой логики",
    ],
  },
  golang: {
    label: "Go",
    checklist: [
      "Проверь работу с context и ошибками",
      "Держи интерфейсы небольшими",
      "Добавь тесты для ключевой логики",
    ],
  },
  python: {
    label: "Python",
    checklist: [
      "Сделай поведение явным через тесты",
      "Проверь обработку исключений",
      "Сохрани код читаемым и простым",
    ],
  },
  java: {
    label: "Java",
    checklist: [
      "Выдели контракты и модели",
      "Проверь исключения и null-сценарии",
      "Сохрани слои решения независимыми",
    ],
  },
  rust: {
    label: "Rust",
    checklist: [
      "Опиши ошибки через Result",
      "Проверь ownership на границах модулей",
      "Добавь тесты для core-логики",
    ],
  },
  sql: {
    label: "SQL",
    checklist: [
      "Проверь миграции на обратимость",
      "Сохрани индексы для частых запросов",
      "Тестируй на пустых данных",
    ],
  },
  yaml: {
    label: "YAML",
    checklist: [
      "Проверь синтаксис и валидность",
      "Опиши секреты отдельно",
      "Добавь комментарии к неочевидным шагам",
    ],
  },
};

const LANGUAGE_ICONS: Record<string, string> = {
  typescript: "logos:typescript-icon",
  ts: "logos:typescript-icon",
  javascript: "logos:javascript",
  js: "logos:javascript",
  python: "logos:python",
  java: "logos:java",
  go: "logos:go",
  golang: "logos:go",
  rust: "logos:rust",
  sql: "vscode-icons:file-type-sql",
  yaml: "vscode-icons:file-type-yaml",
  yml: "vscode-icons:file-type-yaml",
};

export function normalizedType(type: string): string {
  return String(type || "")
    .toLowerCase()
    .replace(/\s+/g, "")
    .replace("-", "");
}

export function typeContent(type: string) {
  return (
    TYPE_CONTENT[normalizedType(type)] || {
      label: type || "Тип не указан",
      title: type ? `${type}-задача` : "Задача",
      copy: "Собери решение строго по условию и проверь ключевые сценарии из ТЗ.",
      checklist: [
        "Разбери требования из ТЗ",
        "Зафиксируй входы и ожидаемый результат",
        "Проверь пограничные сценарии",
      ],
    }
  );
}

export function languageContent(language: string) {
  const normalized = normalizedLanguage(language);
  return (
    LANGUAGE_CONTENT[normalized] || {
      label: language || "Не указан",
      checklist: [
        "Соблюдай выбранный стек решения",
        "Держи код читаемым",
        "Проверь основную логику перед отправкой",
      ],
    }
  );
}

export function languageIcon(language: string): string {
  return LANGUAGE_ICONS[normalizedLanguage(language)] || "carbon:code";
}

export function languageIconUrl(language: string): string {
  return `https://api.iconify.design/${languageIcon(language)}.svg`;
}

function normalizedLanguage(language: string): string {
  return String(language || "")
    .toLowerCase()
    .trim();
}

export function tagSummary(tags: string[]): string {
  if (!tags.length) return "фокус указан в ТЗ";
  return tags.slice(0, 4).join(", ");
}
