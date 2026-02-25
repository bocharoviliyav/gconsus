/**
 * Settings API service
 */
import { apiClient } from './client';

export interface SystemSettings {
  id: string;
  github_enabled: boolean;
  github_url?: string;
  github_token_set: boolean;
  gitlab_enabled: boolean;
  gitlab_url?: string;
  gitlab_token_set: boolean;
  sync_schedule_cron: string;
  sync_enabled: boolean;
  last_sync_at?: string;
  updated_at: string;
}

export interface UpdateSystemSettingsRequest {
  github_enabled?: boolean;
  github_url?: string;
  github_token?: string;
  gitlab_enabled?: boolean;
  gitlab_url?: string;
  gitlab_token?: string;
  sync_schedule_cron?: string;
  sync_enabled?: boolean;
}

export interface SyncStatus {
  is_running: boolean;
  started_at?: string;
  last_completed_at?: string;
  last_error?: string;
  users_synced: number;
  teams_synced: number;
  activities_synced: number;
}

export interface SyncHistoryItem {
  id: string;
  started_at: string;
  completed_at?: string;
  status: 'running' | 'completed' | 'failed';
  users_synced: number;
  teams_synced: number;
  activities_synced: number;
  error_message?: string;
}

export interface SyncHistoryResponse {
  items: SyncHistoryItem[];
  total: number;
  page: number;
  page_size: number;
}

/**
 * Get system settings
 */
export const getSystemSettings = async (): Promise<SystemSettings> => {
  const response = await apiClient.get<SystemSettings>('/settings/system');
  return response.data;
};

/**
 * Update system settings
 */
export const updateSystemSettings = async (
  data: UpdateSystemSettingsRequest
): Promise<SystemSettings> => {
  const response = await apiClient.put<SystemSettings>('/settings/system', data);
  return response.data;
};

/**
 * Get sync status
 */
export const getSyncStatus = async (): Promise<SyncStatus> => {
  const response = await apiClient.get<SyncStatus>('/sync/status');
  return response.data;
};

/**
 * Trigger manual sync
 */
export const triggerSync = async (): Promise<void> => {
  await apiClient.post('/sync/trigger');
};

/**
 * Get sync history
 */
export const getSyncHistory = async (params?: {
  page?: number;
  page_size?: number;
}): Promise<SyncHistoryResponse> => {
  const response = await apiClient.get<SyncHistoryResponse>('/sync/history', {
    params,
  });
  return response.data;
};

/**
 * Test GitHub connection
 */
export const testGitHubConnection = async (url: string, token: string): Promise<boolean> => {
  try {
    await apiClient.post('/settings/test-github', { url, token });
    return true;
  } catch {
    return false;
  }
};

/**
 * Test GitLab connection
 */
export const testGitLabConnection = async (url: string, token: string): Promise<boolean> => {
  try {
    await apiClient.post('/settings/test-gitlab', { url, token });
    return true;
  } catch {
    return false;
  }
};
