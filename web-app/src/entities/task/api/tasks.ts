import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  tasksList,
  tasksGet,
  tasksTagsList,
  tasksLanguagesList,
  type TasksListParams,
  type Task,
  type Tag,
  type Language,
} from "@shared/api";
import { getApiErrorMessage } from "@shared/lib/api-error";

export const taskKeys = {
  all: ["tasks"] as const,
  list: (params: TasksListParams) => [...taskKeys.all, "list", params] as const,
  detail: (id: string) => [...taskKeys.all, "detail", id] as const,
  tags: () => [...taskKeys.all, "tags"] as const,
  languages: () => [...taskKeys.all, "languages"] as const,
};

export function useTaskList(params: TasksListParams, enabled = true) {
  return useQuery({
    queryKey: taskKeys.list(params),
    enabled,
    retry: false,
    placeholderData: keepPreviousData,
    queryFn: async () => {
      const res = await tasksList(params);
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить задачи."));
    },
  });
}

export function useTaskDetail(id: string, enabled = true) {
  return useQuery({
    queryKey: taskKeys.detail(id),
    enabled: enabled && !!id,
    retry: false,
    staleTime: 10 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
    queryFn: async () => {
      const res = await tasksGet(id);
      if (res.status === 200) return res.data as Task;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить задачу."));
    },
  });
}

export function useTaskTags() {
  return useQuery({
    queryKey: taskKeys.tags(),
    retry: false,
    staleTime: 60 * 60 * 1000,
    gcTime: 2 * 60 * 60 * 1000,
    queryFn: async () => {
      const res = await tasksTagsList();
      if (res.status === 200) return res.data as Tag[];
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить теги."));
    },
  });
}

export function useTaskLanguages() {
  return useQuery({
    queryKey: taskKeys.languages(),
    retry: false,
    staleTime: 60 * 60 * 1000,
    gcTime: 2 * 60 * 60 * 1000,
    queryFn: async () => {
      const res = await tasksLanguagesList();
      if (res.status === 200) return res.data as Language[];
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить языки."));
    },
  });
}
