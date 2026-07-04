import { useSearchParams } from "react-router";

export interface TaskFilters {
  type: string;
  level: string;
  tags: string[];
  languages: string[];
  search: string;
}

export function useTaskFilters(): [
  TaskFilters,
  (filters: Partial<TaskFilters>) => void,
  () => void,
] {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters: TaskFilters = {
    type: searchParams.get("type") || "all",
    level: searchParams.get("level") || "all",
    tags: searchParams.getAll("tags"),
    languages: searchParams.getAll("languages"),
    search: searchParams.get("search") || "",
  };

  const setFilters = (partial: Partial<TaskFilters>) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);

      if (partial.type !== undefined) {
        if (partial.type === "all") next.delete("type");
        else next.set("type", partial.type);
      }
      if (partial.level !== undefined) {
        if (partial.level === "all") next.delete("level");
        else next.set("level", partial.level);
      }
      if (partial.tags !== undefined) {
        next.delete("tags");
        for (const tag of partial.tags) next.append("tags", tag);
      }
      if (partial.languages !== undefined) {
        next.delete("languages");
        for (const lang of partial.languages) next.append("languages", lang);
      }
      if (partial.search !== undefined) {
        if (!partial.search) next.delete("search");
        else next.set("search", partial.search);
      }

      return next;
    });
  };

  const clearFilters = () => {
    setSearchParams(new URLSearchParams());
  };

  return [filters, setFilters, clearFilters];
}
