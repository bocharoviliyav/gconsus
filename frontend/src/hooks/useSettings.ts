/**
 * React Query hooks for settings
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getSystemSettings,
  updateSystemSettings,
  getSyncStatus,
  triggerSync,
  getSyncHistory,
  type UpdateSystemSettingsRequest,
} from '../services/api/settings';

/**
 * Hook to fetch system settings
 */
export const useSystemSettings = () => {
  return useQuery({
    queryKey: ['settings', 'system'],
    queryFn: getSystemSettings,
  });
};

/**
 * Hook to update system settings
 */
export const useUpdateSystemSettings = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateSystemSettingsRequest) => updateSystemSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings', 'system'] });
    },
  });
};

/**
 * Hook to fetch sync status
 */
export const useSyncStatus = () => {
  return useQuery({
    queryKey: ['sync', 'status'],
    queryFn: getSyncStatus,
    refetchInterval: 5000, // Refetch every 5 seconds
  });
};

/**
 * Hook to trigger sync
 */
export const useTriggerSync = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: triggerSync,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync', 'status'] });
      queryClient.invalidateQueries({ queryKey: ['sync', 'history'] });
    },
  });
};

/**
 * Hook to fetch sync history
 */
export const useSyncHistory = (params?: { page?: number; page_size?: number }) => {
  return useQuery({
    queryKey: ['sync', 'history', params],
    queryFn: () => getSyncHistory(params),
  });
};
