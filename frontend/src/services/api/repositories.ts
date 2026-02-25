/**
 * Repositories API service
 */
import { apiClient } from "./client";

export interface Repository {
  id: string;
  name: string;
  url?: string;
  description?: string;
  language?: string;
  stars?: number;
  forks?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RepositoryAnalytics {
  repository_id: string;
  repository_name: string;
  repository_url?: string;
  period_start: string;
  period_end: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  contributors_count: number;
  top_contributors: RepositoryContributor[];
  activity_timeline: RepositoryActivityItem[];
  language_distribution: LanguageStats[];
}

export interface RepositoryContributor {
  user_id: string;
  user_name: string;
  user_email: string;
  avatar_url?: string;
  commits: number;
  prs: number;
  reviews: number;
  lines_added: number;
  lines_deleted: number;
  percentage: number;
}

export interface RepositoryActivityItem {
  date: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
}

export interface LanguageStats {
  language: string;
  lines: number;
  percentage: number;
  color?: string;
}

export interface RepositoriesListResponse {
  repositories: Repository[];
  total: number;
  page: number;
  page_size: number;
}

export interface RepositoryLeaderboardItem {
  repository_name: string;
  repository_url?: string;
  total_commits: number;
  total_prs: number;
  contributors_count: number;
  lines_added: number;
  lines_deleted: number;
  activity_score: number;
  rank: number;
}

export interface RepositoriesLeaderboardResponse {
  repositories: RepositoryLeaderboardItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  total_commits: number;
  total_prs: number;
  total_lines_added: number;
  total_lines_deleted: number;
  total_contributors: number;
  period_start: string;
  period_end: string;
  repository_activity_timeline: RepositoryActivityItem[];
}

/**
 * Get list of repositories
 */
export const getRepositories = async (params?: {
  page?: number;
  page_size?: number;
  search?: string;
  is_active?: boolean;
}): Promise<RepositoriesListResponse> => {
  const response = await apiClient.get<RepositoriesListResponse>(
    "/repositories",
    {
      params,
    },
  );
  return response.data;
};

/**
 * Get repository analytics
 */
export const getRepositoryAnalytics = async (
  repositoryId: string,
  params: {
    start_date: string;
    end_date: string;
  },
): Promise<RepositoryAnalytics> => {
  const response = await apiClient.get<RepositoryAnalytics>(
    `/analytics/repositories/${repositoryId}`,
    { params },
  );
  return response.data;
};

/**
 * Get repositories leaderboard
 */
export const getRepositoriesLeaderboard = async (params: {
  start_date: string;
  end_date: string;
  page?: number;
  page_size?: number;
}): Promise<RepositoriesLeaderboardResponse> => {
  const response = await apiClient.get<RepositoriesLeaderboardResponse>(
    "/analytics/repositories/leaderboard",
    { params },
  );
  return response.data;
};
