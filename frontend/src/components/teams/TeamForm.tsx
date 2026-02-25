/**
 * Team form component for create/edit operations
 */
import React, { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../common/Button';
import { Avatar } from '../common/Avatar';
import { XMarkIcon, UserPlusIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import type { Team, TeamMember } from '../../services/api/teams';
import { useTeamMembers, useAddTeamMember, useRemoveTeamMember } from '../../hooks/useTeams';
import { useUsers } from '../../hooks/useUsers';

interface TeamFormProps {
  team?: Team;
  onSubmit: (data: { name: string; description?: string; lead_id?: string }) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

const ROLES: Array<TeamMember['role']> = ['developer', 'lead', 'architect', 'qa', 'analyst', 'devops', 'sre'];

export const TeamForm: React.FC<TeamFormProps> = ({
  team,
  onSubmit,
  onCancel,
  isLoading = false,
}) => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    name: team?.name || '',
    description: team?.description || '',
    lead_id: team?.lead_id || '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Member management state (only used when editing)
  const [userSearch, setUserSearch] = useState('');
  const [showDropdown, setShowDropdown] = useState(false);
  const [selectedRole, setSelectedRole] = useState<TeamMember['role']>('developer');
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Hooks for member management
  const { data: members = [], isLoading: membersLoading } = useTeamMembers(team?.id || '');
  const addMemberMutation = useAddTeamMember();
  const removeMemberMutation = useRemoveTeamMember();
  const { data: usersData } = useUsers(
    team ? { search: userSearch, active: true, limit: 20 } : undefined
  );

  // Filter out users who are already members
  const memberUserIds = new Set(members.map((m) => m.userId));
  const availableUsers = (usersData?.users || []).filter(
    (u) => !memberUserIds.has(u.id)
  );

  useEffect(() => {
    if (team) {
      setFormData({
        name: team.name,
        description: team.description || '',
        lead_id: team.lead_id || '',
      });
    }
  }, [team]);

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t('validation.required');
    } else if (formData.name.length < 3) {
      newErrors.name = t('validation.minLength', { min: 3 });
    } else if (formData.name.length > 100) {
      newErrors.name = t('validation.maxLength', { max: 100 });
    }

    if (formData.description && formData.description.length > 500) {
      newErrors.description = t('validation.maxLength', { max: 500 });
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validate()) {
      onSubmit({
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
        lead_id: formData.lead_id || undefined,
      });
    }
  };

  const handleAddMember = async (userId: string) => {
    if (!team) return;
    try {
      await addMemberMutation.mutateAsync({
        teamId: team.id,
        data: { userId, role: selectedRole },
      });
      setUserSearch('');
      setShowDropdown(false);
    } catch (error) {
      console.error('Failed to add member:', error);
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!team) return;
    try {
      await removeMemberMutation.mutateAsync({ teamId: team.id, userId });
    } catch (error) {
      console.error('Failed to remove member:', error);
    }
  };

  const getUserDisplayName = (user: { firstName: string; lastName: string; username: string }) => {
    return `${user.lastName} ${user.firstName}`.trim() || user.username;
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label
          htmlFor="name"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
        >
          {t('teams.teamName')} <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          id="name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent
            ${
              errors.name
                ? 'border-red-500 dark:border-red-400'
                : 'border-gray-300 dark:border-gray-600'
            }
            bg-white dark:bg-gray-700 text-gray-900 dark:text-white
            placeholder-gray-400 dark:placeholder-gray-500`}
          placeholder={t('teams.teamName')}
          disabled={isLoading}
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.name}</p>
        )}
      </div>

      <div>
        <label
          htmlFor="description"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
        >
          {t('teams.description')}
        </label>
        <textarea
          id="description"
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          rows={3}
          className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent
            ${
              errors.description
                ? 'border-red-500 dark:border-red-400'
                : 'border-gray-300 dark:border-gray-600'
            }
            bg-white dark:bg-gray-700 text-gray-900 dark:text-white
            placeholder-gray-400 dark:placeholder-gray-500 resize-none`}
          placeholder={t('teams.description')}
          disabled={isLoading}
        />
        {errors.description && (
          <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.description}</p>
        )}
      </div>

      {/* Member Management — only when editing */}
      {team && (
        <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
          <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
            {t('teams.members')} ({members.length})
          </h3>

          {/* Add member section */}
          <div className="mb-4" ref={dropdownRef}>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  value={userSearch}
                  onChange={(e) => {
                    setUserSearch(e.target.value);
                    setShowDropdown(true);
                  }}
                  onFocus={() => setShowDropdown(true)}
                  placeholder={t('teams.searchUsers')}
                  className="w-full pl-9 pr-4 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg
                    bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                    placeholder-gray-400 dark:placeholder-gray-500
                    focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
                {/* Dropdown */}
                {showDropdown && userSearch.length >= 1 && (
                  <div className="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600
                    rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    {availableUsers.length === 0 ? (
                      <div className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                        {t('users.noUsers')}
                      </div>
                    ) : (
                      availableUsers.map((user) => (
                        <button
                          key={user.id}
                          type="button"
                          onClick={() => handleAddMember(user.id)}
                          className="w-full flex items-center gap-3 px-4 py-2 text-sm text-left
                            hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                        >
                          <Avatar src={user.photoUrl} name={getUserDisplayName(user)} size="sm" />
                          <div className="flex-1 min-w-0">
                            <div className="text-gray-900 dark:text-white truncate">
                              {getUserDisplayName(user)}
                            </div>
                            <div className="text-xs text-gray-500 dark:text-gray-400 truncate">
                              @{user.username}
                            </div>
                          </div>
                          <UserPlusIcon className="w-4 h-4 text-gray-400 flex-shrink-0" />
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
              <select
                value={selectedRole}
                onChange={(e) => setSelectedRole(e.target.value as TeamMember['role'])}
                className="px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                  focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              >
                {ROLES.map((role) => (
                  <option key={role} value={role}>
                    {t(`teams.roles.${role}`)}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Current members list */}
          <div className="space-y-1 max-h-48 overflow-y-auto">
            {membersLoading ? (
              <div className="text-sm text-gray-500 dark:text-gray-400 py-2">
                {t('common.loading')}
              </div>
            ) : members.length === 0 ? (
              <div className="text-sm text-gray-500 dark:text-gray-400 py-2">
                {t('teams.noMembers')}
              </div>
            ) : (
              members.map((member) => (
                <div
                  key={member.id}
                  className="flex items-center gap-3 px-3 py-2 rounded-lg
                    bg-gray-50 dark:bg-gray-800/50 group"
                >
                  <Avatar
                    src={member.user?.photoUrl}
                    name={member.user ? getUserDisplayName(member.user) : '?'}
                    size="sm"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-gray-900 dark:text-white truncate">
                      {member.user
                        ? getUserDisplayName(member.user)
                        : member.userId}
                    </div>
                    {member.user?.username && (
                      <div className="text-xs text-gray-500 dark:text-gray-400 truncate">
                        @{member.user.username}
                      </div>
                    )}
                  </div>
                  <span className="text-xs px-2 py-0.5 rounded-full bg-gray-200 dark:bg-gray-700
                    text-gray-600 dark:text-gray-300">
                    {t(`teams.roles.${member.role}`)}
                  </span>
                  <button
                    type="button"
                    onClick={() => handleRemoveMember(member.userId)}
                    disabled={removeMemberMutation.isPending}
                    className="opacity-0 group-hover:opacity-100 p-1 rounded-full
                      hover:bg-red-100 dark:hover:bg-red-900/30 transition-all"
                    title={t('teams.removeMember')}
                  >
                    <XMarkIcon className="w-4 h-4 text-red-500" />
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
        <Button
          type="button"
          variant="secondary"
          onClick={onCancel}
          disabled={isLoading}
        >
          {t('common.cancel')}
        </Button>
        <Button
          type="submit"
          variant="primary"
          isLoading={isLoading}
        >
          {team ? t('common.save') : t('common.create')}
        </Button>
      </div>
    </form>
  );
};
