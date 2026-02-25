/**
 * React Query hooks for users
 */
import { useQuery } from '@tanstack/react-query';
import { getUsers } from '../services/api/users';

/**
 * Hook to fetch users list with optional search
 */
export const useUsers = (params?: {
  search?: string;
  active?: boolean;
  limit?: number;
  offset?: number;
}) => {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () => getUsers(params),
  });
};
