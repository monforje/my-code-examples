import styles from "./TaskTag.module.css";

export function TaskTag({ name }: { name: string }) {
  return <span className={styles.tag}>{name}</span>;
}
