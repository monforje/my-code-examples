import { useCallback, useMemo, useRef, useState } from "react";
import { useAuth } from "@features/auth";
import { AppHeader } from "@widgets/header";
import { useTaskList, useTaskTags, useTaskLanguages } from "@entities/task";
import { useTaskFilters, TaskFiltersSidebar, TaskToolbar } from "@features/task-filters";
import { TaskBoard } from "@features/task-tile-layout";
import { TaskCard } from "@features/task-card";
import styles from "./TasksListPage.module.css";

const PAGE_LIMIT = 12;

export function TasksListPage() {
  const { user } = useAuth();
  const [filters, setFilters, clearFilters] = useTaskFilters();
  const [page, setPage] = useState(1);
  const [cursorHistory, setCursorHistory] = useState<(string | undefined)[]>([undefined]);
  const mainRef = useRef<HTMLElement>(null);

  const hasActive =
    filters.type !== "all" ||
    filters.level !== "all" ||
    filters.tags.length > 0 ||
    filters.languages.length > 0 ||
    filters.search.length > 0;

  const currentCursor = cursorHistory[page - 1];

  const queryParams = useMemo(
    () => ({
      limit: PAGE_LIMIT,
      ...(currentCursor ? { cursor: currentCursor } : {}),
      ...(filters.type !== "all" ? { task_type: filters.type as "backend" | "frontend" } : {}),
      ...(filters.level !== "all"
        ? { level: filters.level as "junior" | "middle" | "senior" }
        : {}),
      ...(filters.tags.length > 0 ? { tags: filters.tags } : {}),
      ...(filters.languages.length > 0 ? { languages: filters.languages } : {}),
      ...(filters.search ? { search: filters.search } : {}),
    }),
    [filters, currentCursor],
  );

  const { data: taskData, isFetching } = useTaskList(queryParams, !!user);

  const langCountParams = useMemo(
    () => ({
      limit: 1000,
      ...(filters.type !== "all" ? { task_type: filters.type as "backend" | "frontend" } : {}),
      ...(filters.level !== "all"
        ? { level: filters.level as "junior" | "middle" | "senior" }
        : {}),
      ...(filters.tags.length > 0 ? { tags: filters.tags } : {}),
      ...(filters.search ? { search: filters.search } : {}),
    }),
    [filters.type, filters.level, filters.tags, filters.search],
  );

  const { data: langCountData } = useTaskList(langCountParams, !!user);
  const { data: allTags } = useTaskTags();
  const { data: allLanguages } = useTaskLanguages();

  const resetPagination = useCallback(() => {
    setPage(1);
    setCursorHistory([undefined]);
  }, []);

  const tasks = taskData?.items || [];
  const tags = allTags || [];
  const languages = allLanguages || [];
  const hasNextPage = taskData?.page_info?.has_next_page ?? false;

  const toggleTag = useCallback(
    (tagId: string) => {
      const next = filters.tags.includes(tagId)
        ? filters.tags.filter((t) => t !== tagId)
        : [...filters.tags, tagId];
      setFilters({ tags: next });
      resetPagination();
    },
    [filters.tags, setFilters, resetPagination],
  );

  const toggleLanguage = useCallback(
    (langId: string) => {
      const next = filters.languages.includes(langId)
        ? filters.languages.filter((l) => l !== langId)
        : [...filters.languages, langId];
      setFilters({ languages: next });
      resetPagination();
    },
    [filters.languages, setFilters, resetPagination],
  );

  const langCounts = useMemo(() => {
    const allTasks = langCountData?.items || [];
    const counts: Record<string, number> = {};
    for (const task of allTasks) {
      for (const lang of task.languages) {
        counts[lang.id] = (counts[lang.id] || 0) + 1;
      }
    }
    return counts;
  }, [langCountData]);

  const goToNextPage = useCallback(() => {
    if (!hasNextPage || !taskData?.page_info?.next_cursor) return;
    const nextCursor = taskData.page_info.next_cursor;
    setCursorHistory((prev) => [...prev.slice(0, page), nextCursor]);
    setPage((p) => p + 1);
    if (mainRef.current) mainRef.current.scrollTop = 0;
  }, [hasNextPage, taskData, page]);

  const goToPrevPage = useCallback(() => {
    if (page <= 1) return;
    setPage((p) => p - 1);
    if (mainRef.current) mainRef.current.scrollTop = 0;
  }, [page]);

  return (
    <>
      <AppHeader />
      <div className={`${styles.pageLayoutOuter} page-layout page-layout--tasks`}>
        <div className={styles.tasksLayout}>
          <TaskFiltersSidebar
            filters={filters}
            tags={tags}
            languages={languages}
            taskCounts={{}}
            langCounts={langCounts}
            onToggleTag={toggleTag}
            onToggleLanguage={toggleLanguage}
            onClear={() => {
              clearFilters();
              resetPagination();
            }}
            hasActive={hasActive}
          />
          <main ref={mainRef} className={styles.main}>
            <TaskToolbar
              filters={filters}
              onSetFilter={(partial) => {
                setFilters(partial);
                resetPagination();
              }}
            />
            <TaskBoard
              items={tasks}
              getKey={(task) => task.id}
              renderItem={(task, tile) => (
                <TaskCard task={task} compact={tile.compact} wide={tile.wide} tall={tile.tall} />
              )}
            />
            <div className={styles.pagination}>
              <button
                type="button"
                className={styles.pageBtn}
                disabled={page <= 1 || isFetching}
                onClick={goToPrevPage}
                aria-label="Назад"
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polyline points="15 18 9 12 15 6" />
                </svg>
              </button>
              <span className={styles.pageNumber}>{page}</span>
              <button
                type="button"
                className={styles.pageBtn}
                disabled={!hasNextPage || isFetching}
                onClick={goToNextPage}
                aria-label="Далее"
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polyline points="9 18 15 12 9 6" />
                </svg>
              </button>
            </div>
          </main>
        </div>
      </div>
    </>
  );
}
