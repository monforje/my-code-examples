import type { TaskFilters } from "../model/store";
import { TaskTypeSegmented } from "./TaskTypeSegmented";
import { TaskSearchInput } from "./TaskSearchInput";
import { TaskLevelDropdown } from "./TaskLevelDropdown";
import styles from "./TaskToolbar.module.css";

interface TaskToolbarProps {
  filters: TaskFilters;
  onSetFilter: (partial: Partial<TaskFilters>) => void;
}

export function TaskToolbar({ filters, onSetFilter }: TaskToolbarProps) {
  return (
    <div className={styles.toolbar}>
      <TaskTypeSegmented value={filters.type} onChange={(type) => onSetFilter({ type })} />
      <div className={styles.toolbarRow}>
        <TaskSearchInput value={filters.search} onChange={(search) => onSetFilter({ search })} />
        <TaskLevelDropdown value={filters.level} onChange={(level) => onSetFilter({ level })} />
      </div>
    </div>
  );
}
