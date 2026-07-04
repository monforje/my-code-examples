import type { TaskFilters } from "../model/store";
import type { Tag, Language } from "@shared/api";
import styles from "./TaskFiltersSidebar.module.css";

interface TaskFiltersSidebarProps {
  filters: TaskFilters;
  tags: Tag[];
  languages: Language[];
  taskCounts: Record<string, number>;
  langCounts: Record<string, number>;
  onToggleTag: (tagId: string) => void;
  onToggleLanguage: (langId: string) => void;
  onClear: () => void;
  hasActive: boolean;
}

function Chevron() {
  return (
    <svg
      className={styles.filterCardChevron}
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="6 9 12 15 18 9" />
    </svg>
  );
}

function CheckSvg() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

export function TaskFiltersSidebar({
  filters,
  tags,
  languages,
  langCounts,
  onToggleTag,
  onToggleLanguage,
  onClear,
  hasActive,
}: TaskFiltersSidebarProps) {
  const langSummary = filters.languages.length > 0 ? `${filters.languages.length} выбрано` : "Все";
  const tagsSummary = filters.tags.length > 0 ? `${filters.tags.length} выбрано` : "Все";

  return (
    <aside className={styles.sidebar} aria-label="Фильтры задач">
      <div className={styles.sidebarHeader}>
        <span>Фильтры</span>
        <button type="button" className={styles.clearBtn} onClick={onClear} hidden={!hasActive}>
          Сбросить
        </button>
      </div>

      {/* Language filter */}
      <details className={styles.filterCard} open={filters.languages.length > 0}>
        <summary className={styles.filterCardSummary}>
          <span className={styles.filterCardSummaryLeft}>
            <span className={styles.filterCardLabel}>Язык</span>
            <strong>{langSummary}</strong>
          </span>
          <Chevron />
        </summary>
        <div className={styles.filterList}>
          {languages.map((lang) => {
            const active = filters.languages.includes(lang.id);
            const count = langCounts[lang.id] || 0;
            return (
              <button
                key={lang.id}
                type="button"
                role="option"
                aria-pressed={active}
                className={`${styles.filterRow} ${active ? styles.filterRowActive : ""}`}
                onClick={() => onToggleLanguage(lang.id)}
              >
                <span className={styles.filterRowCheck}>
                  <CheckSvg />
                </span>
                <span className={styles.filterRowText}>{lang.name}</span>
                <span className={styles.filterRowCount}>{count}</span>
              </button>
            );
          })}
        </div>
      </details>

      {/* Tags filter */}
      <details className={styles.filterCard} open={filters.tags.length > 0}>
        <summary className={styles.filterCardSummary}>
          <span className={styles.filterCardSummaryLeft}>
            <span className={styles.filterCardLabel}>Теги</span>
            <strong>{tagsSummary}</strong>
          </span>
          <Chevron />
        </summary>
        <div className={styles.filterTags}>
          {tags.map((tag) => {
            const active = filters.tags.includes(tag.id);
            return (
              <button
                key={tag.id}
                type="button"
                aria-pressed={active}
                className={`${styles.tagFilter} ${active ? styles.tagFilterActive : ""}`}
                onClick={() => onToggleTag(tag.id)}
              >
                <span className={styles.tagFilterText}>{tag.name}</span>
                {active && (
                  <svg
                    className={styles.tagFilterX}
                    width="10"
                    height="10"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="3"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                )}
              </button>
            );
          })}
        </div>
      </details>
    </aside>
  );
}
