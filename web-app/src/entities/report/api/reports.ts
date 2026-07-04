import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { reportsGet, reportsList, type ReportsListParams } from "@shared/api";
import { getApiErrorMessage } from "@shared/lib/api-error";

// Интервал автообновления для pending-отчётов (real-time переход pending → финал).
const PENDING_REFETCH_MS = 5000;

export const reportKeys = {
  all: ["reports"] as const,
  list: (params: Omit<ReportsListParams, "cursor">) => [...reportKeys.all, "list", params] as const,
  detail: (id: string) => [...reportKeys.all, "detail", id] as const,
};

export function useReportList(params: Omit<ReportsListParams, "cursor">) {
  return useInfiniteQuery({
    queryKey: reportKeys.list(params),
    retry: false,
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }) => {
      const res = await reportsList({ ...params, cursor: pageParam });
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить отчёты."));
    },
    getNextPageParam: (lastPage) =>
      lastPage.page_info.has_next_page ? lastPage.page_info.next_cursor : undefined,
    // Автообновление, пока в списке есть pending-отчёты (live-статус CI-прогона).
    refetchInterval: (query) => {
      const pages = query.state.data?.pages;
      if (!pages) return false;
      return pages.some((page) => page.items.some((r) => r.status === "pending"))
        ? PENDING_REFETCH_MS
        : false;
    },
  });
}

// useReportDetail — один отчёт по id через GET /reports/{id} (без выборки всего списка).
// Пока статус pending — опрашивает сервер для real-time перехода в финальный статус.
export function useReportDetail(id: string, enabled = true) {
  return useQuery({
    queryKey: reportKeys.detail(id),
    enabled: enabled && !!id,
    retry: false,
    queryFn: async () => {
      const res = await reportsGet(id);
      if (res.status === 200) return res.data;
      throw new Error(getApiErrorMessage(res, "Не удалось загрузить CI-отчёт."));
    },
    refetchInterval: (query) =>
      query.state.data?.status === "pending" ? PENDING_REFETCH_MS : false,
  });
}
