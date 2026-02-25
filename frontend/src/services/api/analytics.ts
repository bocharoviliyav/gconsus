/**
 * Analytics API service
 */
import { apiClient } from "./client";

export interface TeamAnalytics {
  team_id: string;
  team_name: string;
  lead_id?: string;
  lead_name?: string;
  period_start: string;
  period_end: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  member_stats: MemberStats[];
  repository_stats: RepositoryStats[];
  repository_stats_pagination: {
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
  };
  activity_timeline: ActivityTimelineItem[];
}

export interface MemberStats {
  user_id: string;
  user_name: string;
  user_email: string;
  avatar_url?: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
  lines_added: number;
  lines_deleted: number;
  score: number;
  rank: number;
}

export interface RepositoryStats {
  repository_name: string;
  repository_url?: string;
  commits: number;
  prs: number;
  contributors: number;
  lines_added: number;
  lines_deleted: number;
}

export interface ActivityTimelineItem {
  date: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
}

export interface UserAnalytics {
  user_id: string;
  user_name: string;
  user_email: string;
  avatar_url?: string;
  period_start: string;
  period_end: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  teams: string[];
  repository_contributions: RepositoryContribution[];
  repository_contributions_pagination: {
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
  };
  activity_timeline: ActivityTimelineItem[];
  recent_activities: RecentActivity[];
}

export interface RepositoryContribution {
  repository_name: string;
  repository_url?: string;
  commits: number;
  prs: number;
  lines_added: number;
  lines_deleted: number;
}

export interface RecentActivity {
  id: string;
  type: "commit" | "pr" | "review" | "issue";
  title: string;
  repository_name: string;
  url?: string;
  created_at: string;
}

/**
 * Get team analytics
 */
export const getTeamAnalytics = async (
  teamId: string,
  params: {
    start_date: string;
    end_date: string;
    repo_page?: number;
    repo_page_size?: number;
  },
): Promise<TeamAnalytics> => {
  const response = await apiClient.get<TeamAnalytics>(
    `/analytics/teams/${teamId}`,
    {
      params,
    },
  );
  return response.data;
};

/**
 * Get user analytics
 */
export const getUserAnalytics = async (
  userId: string,
  params: {
    start_date: string;
    end_date: string;
    repo_page?: number;
    repo_page_size?: number;
  },
): Promise<UserAnalytics> => {
  const response = await apiClient.get<UserAnalytics>(
    `/analytics/users/${userId}`,
    {
      params,
    },
  );
  return response.data;
};
