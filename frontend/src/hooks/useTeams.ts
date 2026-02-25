/**
 * React Query hooks for teams management
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getTeams,
  getTeam,
  getTeamMembers,
  createTeam,
  updateTeam,
  deleteTeam,
  addTeamMember,
  removeTeamMember,
  type CreateTeamRequest,
  type UpdateTeamRequest,
  type AddTeamMemberRequest,
} from '../services/api/teams';

/**
 * Hook to fetch teams list
 */
export const useTeams = (params?: {
  page?: number;
  page_size?: number;
  search?: string;
}) => {
  return useQuery({
    queryKey: ['teams', params],
    queryFn: () => getTeams(params),
  });
};

/**
 * Hook to fetch single team
 */
export const useTeam = (id: string) => {
  return useQuery({
    queryKey: ['teams', id],
    queryFn: () => getTeam(id),
    enabled: !!id,
  });
};

/**
 * Hook to fetch team members
 */
export const useTeamMembers = (teamId: string) => {
  return useQuery({
    queryKey: ['teamMembers', teamId],
    queryFn: () => getTeamMembers(teamId),
    enabled: !!teamId,
  });
};

/**
 * Hook to create team
 */
export const useCreateTeam = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTeamRequest) => createTeam(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
    },
  });
};

/**
 * Hook to update team
 */
export const useUpdateTeam = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTeamRequest }) =>
      updateTeam(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['teams', variables.id] });
    },
  });
};

/**
 * Hook to delete team
 */
export const useDeleteTeam = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteTeam(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
    },
  });
};

/**
 * Hook to add team member
 */
export const useAddTeamMember = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, data }: { teamId: string; data: AddTeamMemberRequest }) =>
      addTeamMember(teamId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['teams', variables.teamId] });
      queryClient.invalidateQueries({ queryKey: ['teamMembers', variables.teamId] });
    },
  });
};

/**
 * Hook to remove team member
 */
export const useRemoveTeamMember = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, userId }: { teamId: string; userId: string }) =>
      removeTeamMember(teamId, userId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['teams', variables.teamId] });
      queryClient.invalidateQueries({ queryKey: ['teamMembers', variables.teamId] });
    },
  });
};
