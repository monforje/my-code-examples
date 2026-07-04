import styles from "./TaskLevelBadge.module.css";
import { LEVEL_CONFIG, type LevelDisplay } from "../model/types";

export function TaskLevelBadge({ level }: { level: string }) {
  const lv = (level as LevelDisplay) || "middle";
  const conf = LEVEL_CONFIG[lv] || LEVEL_CONFIG.middle;
  return (
    <span
      className={styles.badge}
      style={{ color: conf.color, background: conf.bg, borderColor: conf.border }}
    >
      {level}
    </span>
  );
}
