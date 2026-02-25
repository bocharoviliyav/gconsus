import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card } from '../../components/common/Card';
import { Modal } from '../../components/common/Modal';
import { Button } from '../../components/common/Button';
import { TeamsTable } from '../../components/teams/TeamsTable';
import { TeamForm } from '../../components/teams/TeamForm';
import { PlusIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { useTeams, useCreateTeam, useUpdateTeam, useDeleteTeam } from '../../hooks/useTeams';
import type { Team } from '../../services/api/teams';

const Teams: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);

  const { data, isLoading } = useTeams({ page, page_size: 20, search });
  const createTeamMutation = useCreateTeam();
  const updateTeamMutation = useUpdateTeam();
  const deleteTeamMutation = useDeleteTeam();

  const handleCreateTeam = async (data: { name: string; description?: string; lead_id?: string }) => {
    try {
      await createTeamMutation.mutateAsync(data);
      setIsCreateModalOpen(false);
    } catch (error) {
      console.error('Failed to create team:', error);
    }
  };

  const handleUpdateTeam = async (data: { name: string; description?: string; lead_id?: string }) => {
    if (!selectedTeam) return;
    try {
      await updateTeamMutation.mutateAsync({ id: selectedTeam.id, data });
      setIsEditModalOpen(false);
      setSelectedTeam(null);
    } catch (error) {
      console.error('Failed to update team:', error);
    }
  };

  const handleDeleteTeam = async () => {
    if (!selectedTeam) return;
    try {
      await deleteTeamMutation.mutateAsync(selectedTeam.id);
      setIsDeleteModalOpen(false);
      setSelectedTeam(null);
    } catch (error) {
      console.error('Failed to delete team:', error);
    }
  };

  const openEditModal = (team: Team) => {
    setSelectedTeam(team);
    setIsEditModalOpen(true);
  };

  const openDeleteModal = (team: Team) => {
    setSelectedTeam(team);
    setIsDeleteModalOpen(true);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {t('teams.title')}
          </h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            Manage your teams and members
          </p>
        </div>
        <Button
          variant="primary"
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center gap-2"
        >
          <PlusIcon className="w-5 h-5" />
          {t('teams.createTeam')}
        </Button>
      </div>

      {/* Search */}
      <Card padding="md">
        <div className="relative">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder={t('common.search')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
              bg-white dark:bg-gray-700 text-gray-900 dark:text-white
              placeholder-gray-400 dark:placeholder-gray-500
              focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </Card>

      {/* Teams Table */}
      <Card padding="none">
        <TeamsTable
          teams={data?.teams || []}
          isLoading={isLoading}
          onEdit={openEditModal}
          onDelete={openDeleteModal}
        />
      </Card>

      {/* Pagination */}
      {data && data.total > data.page_size && (
        <div className="flex justify-center gap-2">
          <Button
            variant="secondary"
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            {t('common.previous')}
          </Button>
          <span className="flex items-center px-4 text-sm text-gray-600 dark:text-gray-400">
            Page {page} of {Math.ceil(data.total / data.page_size)}
          </span>
          <Button
            variant="secondary"
            onClick={() => setPage(p => p + 1)}
            disabled={page >= Math.ceil(data.total / data.page_size)}
          >
            {t('common.next')}
          </Button>
        </div>
      )}

      {/* Create Team Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title={t('teams.createTeam')}
      >
        <TeamForm
          onSubmit={handleCreateTeam}
          onCancel={() => setIsCreateModalOpen(false)}
          isLoading={createTeamMutation.isPending}
        />
      </Modal>

      {/* Edit Team Modal */}
      <Modal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedTeam(null);
        }}
        title={t('teams.editTeam')}
      >
        <TeamForm
          team={selectedTeam || undefined}
          onSubmit={handleUpdateTeam}
          onCancel={() => {
            setIsEditModalOpen(false);
            setSelectedTeam(null);
          }}
          isLoading={updateTeamMutation.isPending}
        />
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedTeam(null);
        }}
        title={t('teams.deleteTeam')}
      >
        <div className="space-y-4">
          <p className="text-gray-600 dark:text-gray-400">
            {t('teams.confirmDelete')}
          </p>
          {selectedTeam && (
            <div className="p-4 bg-gray-100 dark:bg-gray-800 rounded-lg">
              <p className="font-medium text-gray-900 dark:text-white">
                {selectedTeam.name}
              </p>
              {selectedTeam.description && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {selectedTeam.description}
                </p>
              )}
            </div>
          )}
          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
            <Button
              variant="secondary"
              onClick={() => {
                setIsDeleteModalOpen(false);
                setSelectedTeam(null);
              }}
              disabled={deleteTeamMutation.isPending}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="danger"
              onClick={handleDeleteTeam}
              isLoading={deleteTeamMutation.isPending}
            >
              {t('common.delete')}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Teams;
