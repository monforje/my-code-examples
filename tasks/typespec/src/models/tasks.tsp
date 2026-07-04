namespace Tasks;

// ─── Task ─────────────────────────────────────────────────────────────────────

model Task {
  id: string;
  task_name: string;
  title: string;
  description: string;
  specification_md_text: string;
  task_type: TaskType;
  level: Level;
  tags: Tag[];
  languages: Language[];
  created_at: utcDateTime;
}

model TaskListItem {
  id: string;
  task_name: string;
  title: string;
  description: string;
  task_type: TaskType;
  level: Level;
  tags: Tag[];
  languages: Language[];
  created_at: utcDateTime;
}

// ─── Request Models ───────────────────────────────────────────────────────────

model CreateTaskRequest {
  task_name: string;
  title: string;
  description: string;
  specification_md_text: string;
  task_type: TaskType;
  level: Level;
  tag_ids: string[];
  language_ids: string[];
}

model UpdateTaskRequest {
  task_name?: string;
  title?: string;
  description?: string;
  specification_md_text?: string;
  task_type?: TaskType;
  level?: Level;
  tag_ids?: string[];
  language_ids?: string[];
}

// ─── Response Models ──────────────────────────────────────────────────────────

model TaskListResponse {
  items: TaskListItem[];
  page_info: PageInfo;
}
