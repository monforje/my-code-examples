import type { CSSProperties } from "react";
import { useNavigate } from "react-router";
import { Icon } from "@iconify/react";
import type { TaskListItem } from "@shared/api";
import { TaskLevelBadge, TaskTypeDot, TaskTag, levelDisplay, languageIcon } from "@entities/task";
import styles from "./TaskCard.module.css";

interface TaskCardProps {
  task: TaskListItem;
  className?: string;
  style?: CSSProperties;
  compact?: boolean;
  wide?: boolean;
  tall?: boolean;
}

export function TaskCard({
  task,
  className = "",
  style,
  compact = false,
  wide = false,
  tall = false,
}: TaskCardProps) {
  const navigate = useNavigate();
  const lv = levelDisplay(task.level);
  const languageName = task.languages[0]?.name || "";
  const langIcon = languageIcon(languageName);
  const tagNames = task.tags.map((t) => t.name);
  const classes = [
    styles.card,
    compact ? styles.compact : "",
    wide ? styles.wide : "",
    tall ? styles.tall : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <button
      type="button"
      className={classes}
      style={style}
      aria-label={`Открыть задачу: ${task.title}`}
      onClick={() => navigate(`/tasks/${task.id}`)}
    >
      <div className={styles.body}>
        <div className={styles.header}>
          <h3 className={styles.title}>{task.title}</h3>
          <TaskLevelBadge level={lv} />
        </div>
        {task.description && <p className={styles.desc}>{task.description}</p>}
        <div className={styles.footer}>
          <div className={styles.meta}>
            {languageName && (
              <span className={styles.lang} title={languageName} aria-label={languageName}>
                <Icon icon={langIcon} width={18} height={18} />
              </span>
            )}
            <TaskTypeDot type={task.task_type} />
          </div>
          <div className={styles.tags}>
            {tagNames.map((tag) => (
              <TaskTag key={tag} name={tag} />
            ))}
          </div>
        </div>
      </div>
    </button>
  );
}
