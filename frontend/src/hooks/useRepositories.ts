/**
 * React Query hooks for repositories
 */
import { useQuery } from "@tanstack/react-query";
import {
  getRepositories,
  getRepositoryAnalytics,
  getRepositoriesLeaderboard,
} from "../services/api/repositories";

/**
 * Hook to fetch repositories list
 */
export const useRepositories = (params?: {
  page?: number;
  page_size?: number;
  search?: string;
  is_active?: boolean;
}) => {
  return useQuery({
    queryKey: ["repositories", params],
    queryFn: () => getRepositories(params),
  });
};

/**
 * Hook to fetch repository analytics
 */
export const useRepositoryAnalytics = (
  repositoryId: string,
  startDate: string,
  endDate: string,
) => {
  return useQuery({
    queryKey: ["analytics", "repositories", repositoryId, startDate, endDate],
    queryFn: () =>
      getRepositoryAnalytics(repositoryId, {
        start_date: startDate,
        end_date: endDate,
      }),
    enabled: !!repositoryId && !!startDate && !!endDate,
  });
};

/**
 * Hook to fetch repositories leaderboard
 */
export const useRepositoriesLeaderboard = (
  startDate: string,
  endDate: string,
  page?: number,
  pageSize?: number,
) => {
  return useQuery({
    queryKey: [
      "analytics",
      "repositories",
      "leaderboard",
      startDate,
      endDate,
      page,
      pageSize,
    ],
    queryFn: () =>
      getRepositoriesLeaderboard({
        start_date: startDate,
        end_date: endDate,
        page,
        page_size: pageSize,
      }),
    enabled: !!startDate && !!endDate,
  });
};
