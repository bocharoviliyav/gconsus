// User types
export interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
  department?: string;
  position?: string;
  github_username?: string;
  gitlab_username?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  full_name: string;
  department?: string;
  position?: string;
  github_username?: string;
  gitlab_username?: string;
}

export interface UpdateUserRequest {
  email?: string;
  full_name?: string;
  department?: string;
  position?: string;
  github_username?: string;
  gitlab_username?: string;
  is_active?: boolean;
}

// Team types
export interface Team {
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  member_count?: number;
}

export interface TeamMember {
  user_id: string;
  team_id: string;
  role: "member" | "lead";
  joined_at: string;
  user?: User;
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
}

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
}

export interface AddTeamMemberRequest {
  user_id: string;
  role: "member" | "lead";
}

// Analytics types
// export interface LeaderboardEntry {
//   user_id: string;
//   user_name: string;
//   github_username?: string;
//   total_commits: number;
//   total_prs: number;
//   total_reviews: number;
//   total_issues: number;
//   lines_added: number;
//   lines_deleted: number;
//   rank: number;
// }
export interface LeaderboardEntry {
  user_id: string;
  user_name: string;
  github_username?: string;
  avatar_url?: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
  lines_added: number;
  lines_deleted: number;
  rank: number;
}

export interface TeamLeaderboardEntry {
  team_id: string;
  team_name: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  member_count: number;
  rank: number;
}

export interface UserAnalytics {
  user_id: string;
  user_name: string;
  period_start: string;
  period_end: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  repositories: RepositoryContribution[];
  activity_timeline: ActivityTimelineEntry[];
}

export interface TeamAnalytics {
  team_id: string;
  team_name: string;
  period_start: string;
  period_end: string;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  lines_added: number;
  lines_deleted: number;
  members: MemberAnalytics[];
  top_repositories: RepositoryContribution[];
}

export interface RepositoryContribution {
  repository_owner: string;
  repository_name: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
  lines_added: number;
  lines_deleted: number;
}

export interface ActivityTimelineEntry {
  date: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
}

export interface MemberAnalytics {
  user_id: string;
  user_name: string;
  commits: number;
  prs: number;
  reviews: number;
  issues: number;
  lines_added: number;
  lines_deleted: number;
}

export interface DashboardStats {
  period_start: string;
  period_end: string;
  total_users: number;
  total_teams: number;
  total_commits: number;
  total_prs: number;
  total_reviews: number;
  total_issues: number;
  top_contributors: LeaderboardEntry[];
  top_teams: TeamLeaderboardEntry[];
  activity_trend: ActivityTimelineEntry[];
  repository_distribution: RepositoryStats[];
}

export interface RepositoryStats {
  repository_owner: string;
  repository_name: string;
  contributor_count: number;
  total_commits: number;
  total_prs: number;
}

// Auth types
export interface UserClaims {
  sub: string;
  email: string;
  preferred_username: string;
  name: string;
  given_name?: string;
  family_name?: string;
  roles: string[];
}

export interface AuthState {
  isAuthenticated: boolean;
  user: UserClaims | null;
  token: string | null;
  roles: string[];
}

// API Response types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ListUsersResponse {
  users: User[];
  count: number;
  limit: number;
  offset: number;
}

export interface ListTeamsResponse {
  teams: Team[];
  count: number;
  limit: number;
  offset: number;
}

export interface LeaderboardResponse {
  leaderboard: LeaderboardEntry[];
  period: {
    start: string;
    end: string;
  };
  count: number;
}

export interface SyncStatsResponse {
  message: string;
  users_created: number;
  users_updated: number;
  users_disabled: number;
  total: number;
}

// Theme and settings
export type Theme = "light" | "dark" | "system";
export type Locale = "en" | "ru";

export interface AppSettings {
  theme: Theme;
  locale: Locale;
  sidebarCollapsed: boolean;
}
