/**
 * Users API service
 */
import { apiClient } from './client';

export interface User {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  patronymic?: string;
  email?: string;
  photoUrl?: string;
  position?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface UsersListResponse {
  users: User[];
  limit: number;
  offset: number;
  count: number;
}

/**
 * Get list of users with optional search
 */
export const getUsers = async (params?: {
  search?: string;
  active?: boolean;
  limit?: number;
  offset?: number;
}): Promise<UsersListResponse> => {
  const response = await apiClient.get<UsersListResponse>('/users', { params });
  return response.data;
};
