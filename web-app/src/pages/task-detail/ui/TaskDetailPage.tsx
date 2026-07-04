import { useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { Icon } from "@iconify/react";
import { useAuth } from "@features/auth";
import { AppHeader } from "@widgets/header";
import {
  useTaskDetail,
  typeContent,
  languageContent,
  languageIcon,
  tagSummary,
  levelDisplay,
  TaskTag,
  TaskLevelBadge,
  TaskTypeDot,
} from "@entities/task";
import { MarkdownRenderer } from "@shared/ui";
import styles from "./TaskDetailPage.module.css";

export function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { data: task, isLoading } = useTaskDetail(taskId || "", !!user && !!taskId);

  useEffect(() => {
    if (!isLoading && !task) navigate("/tasks", { replace: true });
  }, [isLoading, navigate, task]);

  if (isLoading) return null;
  if (!task) return null;

  const lv = levelDisplay(task.level);
  const type = typeContent(task.task_type);
  const language = languageContent(task.languages[0]?.name || "");
  const langIcon = languageIcon(task.languages[0]?.name || "");
  const tagNames = task.tags.map((t) => t.name);

  const generatedCards = [
    {
      title: type.title,
      text: type.copy,
    },
    {
      title: task.languages[0] ? `Решение на ${language.label}` : "Язык решения",
      text: task.languages[0]
        ? `Пиши решение на ${language.label} и сверяй реализацию с требованиями из ТЗ.`
        : "Язык решения не указан, поэтому ориентируйся на требования из ТЗ.",
    },
    {
      title: tagNames.length ? "Технический фокус" : "Фокус из условия",
      text: tagNames.length
        ? `Держи в фокусе темы: ${tagSummary(tagNames)}.`
        : "Дополнительные теги не переданы — основной источник требований находится в ТЗ.",
    },
  ];

  const checklistItems = [...type.checklist.slice(0, 2), ...language.checklist.slice(0, 2)]
    .filter((item, i, arr) => arr.indexOf(item) === i)
    .slice(0, 5);

  return (
    <>
      <AppHeader />
      <main className={`${styles.page} task-page`}>
        <div className={styles.gridBg} aria-hidden="true" />

        {/* Hero section */}
        <section className={styles.hero} id="task-overview">
          <div className={`${styles.heroLayout} cd-container`}>
            <div className={styles.heroCopy}>
              <div className={styles.chipRow}>
                <TaskTypeDot type={task.task_type} />
                {task.languages[0] && (
                  <span
                    className="task-card__lang"
                    title={task.languages[0].name}
                    aria-label={task.languages[0].name}
                  >
                    <Icon icon={langIcon} width={18} height={18} />
                  </span>
                )}
                <TaskLevelBadge level={lv} />
              </div>
              <h1 className={styles.heroTitle}>{task.title}</h1>
              <p className={styles.heroDescription}>
                {task.description || "Описание задачи появится здесь, когда оно будет передано."}
              </p>
              <div className={styles.heroTags}>
                {tagNames.map((tag) => (
                  <TaskTag key={tag} name={tag} />
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Generated content grid */}
        <section className={`${styles.sectionTight} cd-section task-section-tight`}>
          <div className={`${styles.sectionHeading} cd-container cd-stack-lg`}>
            <div>
              <h2>Короткая карта перед стартом</h2>
              <p>Сначала разберись с форматом, затем переходи к полному ТЗ и реализации.</p>
            </div>
            <div className={styles.generatedGrid}>
              {generatedCards.map((card, i) => (
                <article key={i} className={`${styles.generatedCard} cd-card task-generated-card`}>
                  <h3>{card.title}</h3>
                  <p>{card.text}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        {/* Spec + checklist */}
        <section className="cd-section" id="task-spec">
          <div className={`${styles.contentLayout} cd-container`}>
            <article className={`${styles.specCard} cd-card task-spec-card`}>
              <header className={styles.specHeader}>
                <div>
                  <h2>Техническое задание</h2>
                </div>
              </header>
              <div className={styles.specBody}>
                <MarkdownRenderer content={task.specification_md_text || ""} />
              </div>
            </article>
            <aside className="task-side" aria-label="Чеклист">
              <section className={`${styles.checklistCard} cd-card-soft task-checklist-card`}>
                <div className={styles.checklistHeader}>
                  <span className={styles.checklistIcon} aria-hidden="true">
                    ✓
                  </span>
                  <div>
                    <h3>Перед отправкой</h3>
                    <p>Проверь эти пункты перед отправкой решения.</p>
                  </div>
                </div>
                <ul className={styles.checklist}>
                  {checklistItems.map((item, i) => (
                    <li key={i}>{item}</li>
                  ))}
                </ul>
              </section>
            </aside>
          </div>
        </section>
      </main>
    </>
  );
}
