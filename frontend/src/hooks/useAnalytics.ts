/**
 * React Query hooks for analytics
 */
import { useQuery } from "@tanstack/react-query";
import { getTeamAnalytics, getUserAnalytics } from "../services/api/analytics";

/**
 * Hook to fetch team analytics
 */
export const useTeamAnalytics = (
  teamId: string,
  startDate: string,
  endDate: string,
  repoPage?: number,
  repoPageSize?: number,
) => {
  return useQuery({
    queryKey: [
      "analytics",
      "teams",
      teamId,
      startDate,
      endDate,
      repoPage,
      repoPageSize,
    ],
    queryFn: () =>
      getTeamAnalytics(teamId, {
        start_date: startDate,
        end_date: endDate,
        repo_page: repoPage,
        repo_page_size: repoPageSize,
      }),
    enabled: !!teamId && !!startDate && !!endDate,
  });
};

/**
 * Hook to fetch user analytics
 */
export const useUserAnalytics = (
  userId: string,
  startDate: string,
  endDate: string,
  repoPage?: number,
  repoPageSize?: number,
) => {
  return useQuery({
    queryKey: [
      "analytics",
      "users",
      userId,
      startDate,
      endDate,
      repoPage,
      repoPageSize,
    ],
    queryFn: () =>
      getUserAnalytics(userId, {
        start_date: startDate,
        end_date: endDate,
        repo_page: repoPage,
        repo_page_size: repoPageSize,
      }),
    enabled: !!userId && !!startDate && !!endDate,
  });
};
