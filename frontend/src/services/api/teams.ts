/**
 * Teams API service
 */
import { apiClient } from './client';

export interface Team {
  id: string;
  name: string;
  description?: string;
  lead_id?: string;
  lead_name?: string;
  member_count: number;
  created_at: string;
  updated_at: string;
}

export interface TeamMemberUser {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  patronymic?: string;
  email?: string;
  photoUrl?: string;
  position?: string;
  isActive: boolean;
}

export interface TeamMember {
  id: string;
  teamId: string;
  userId: string;
  role: 'developer' | 'lead' | 'architect' | 'qa' | 'analyst' | 'devops' | 'sre';
  joinedAt: string;
  leftAt?: string;
  user?: TeamMemberUser;
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
  lead_id?: string;
}

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
  lead_id?: string;
}

export interface AddTeamMemberRequest {
  userId: string;
  role: 'developer' | 'lead' | 'architect' | 'qa' | 'analyst' | 'devops' | 'sre';
}

export interface TeamsListResponse {
  teams: Team[];
  total: number;
  page: number;
  page_size: number;
}

/**
 * Get list of teams
 */
export const getTeams = async (params?: {
  page?: number;
  page_size?: number;
  search?: string;
}): Promise<TeamsListResponse> => {
  const response = await apiClient.get<TeamsListResponse>('/teams', { params });
  return response.data;
};

/**
 * Get team by ID
 */
export const getTeam = async (id: string): Promise<Team> => {
  const response = await apiClient.get<Team>(`/teams/${id}`);
  return response.data;
};

/**
 * Get team members
 */
export interface TeamMembersResponse {
  members: TeamMember[];
  count: number;
}

export const getTeamMembers = async (teamId: string): Promise<TeamMember[]> => {
  const response = await apiClient.get<TeamMembersResponse>(`/teams/${teamId}/members`);
  return response.data.members || [];
};

/**
 * Create new team
 */
export const createTeam = async (data: CreateTeamRequest): Promise<Team> => {
  const response = await apiClient.post<Team>('/teams', data);
  return response.data;
};

/**
 * Update team
 */
export const updateTeam = async (id: string, data: UpdateTeamRequest): Promise<Team> => {
  const response = await apiClient.put<Team>(`/teams/${id}`, data);
  return response.data;
};

/**
 * Delete team
 */
export const deleteTeam = async (id: string): Promise<void> => {
  await apiClient.delete(`/teams/${id}`);
};

/**
 * Add member to team
 */
export const addTeamMember = async (
  teamId: string,
  data: AddTeamMemberRequest
): Promise<void> => {
  await apiClient.post(`/teams/${teamId}/members`, data);
};

/**
 * Remove member from team
 */
export const removeTeamMember = async (teamId: string, userId: string): Promise<void> => {
  await apiClient.delete(`/teams/${teamId}/members/${userId}`);
};
