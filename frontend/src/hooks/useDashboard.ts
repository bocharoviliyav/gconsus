import { useQuery } from "@tanstack/react-query";
import { getDashboard } from "../services/api/dashboard";
import type { DashboardStats } from "../types";

export const useDashboard = (start: string, end: string) => {
  return useQuery<DashboardStats>({
    queryKey: ["dashboard", start, end],
    queryFn: () => getDashboard(start, end),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!start && !!end,
  });
};
