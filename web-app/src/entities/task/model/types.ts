import type { Task, TaskListItem, Tag, Language } from "@shared/api";

export type { Task, TaskListItem, Tag, Language };

export type LevelDisplay = "junior" | "middle" | "senior";

export const LEVEL_CONFIG: Record<
  LevelDisplay,
  { color: string; bg: string; border: string; hint: string }
> = {
  junior: { color: "#16a34a", bg: "#dcfce7", border: "#16a34a", hint: "База" },
  middle: { color: "#eab308", bg: "#fef9c3", border: "#eab308", hint: "Практика" },
  senior: { color: "#dc2626", bg: "#fee2e2", border: "#dc2626", hint: "Сложно" },
};

export const TYPE_COLOR: Record<string, string> = {
  backend: "var(--cd-primary)",
  frontend: "var(--cd-success, #22c55e)",
};

export function levelDisplay(level: string): LevelDisplay {
  return (level as LevelDisplay) || "middle";
}
