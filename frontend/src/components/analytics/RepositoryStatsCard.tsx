/**
 * Repository stats card component
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import type { RepositoryStats } from '../../services/api/analytics';

interface RepositoryStatsCardProps {
  repositories: RepositoryStats[];
  isLoading?: boolean;
}

export const RepositoryStatsCard: React.FC<RepositoryStatsCardProps> = ({
  repositories,
  isLoading = false,
}) => {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-20 bg-gray-200 dark:bg-gray-700 rounded-lg" />
        ))}
      </div>
    );
  }

  if (repositories.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 dark:text-gray-500 text-6xl mb-4">📦</div>
        <p className="text-gray-600 dark:text-gray-400">{t('analytics.noData')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {repositories.map((repo, index) => (
        <div
          key={repo.repository_name}
          className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div className="flex items-start justify-between mb-3">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                  #{index + 1}
                </span>
                <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                  {repo.repository_name}
                </h4>
              </div>
              {repo.repository_url && (
                <a
                  href={repo.repository_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-primary-600 dark:text-primary-400 hover:underline"
                  onClick={(e) => e.stopPropagation()}
                >
                  View on Git
                </a>
              )}
            </div>
            <div className="text-right">
              <div className="text-xs text-gray-500 dark:text-gray-400">
                {repo.contributors} {t('analytics.contributors')}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-4 gap-4">
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.commits')}
              </div>
              <div className="text-lg font-bold text-gray-900 dark:text-white">
                {repo.commits.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.pullRequests')}
              </div>
              <div className="text-lg font-bold text-gray-900 dark:text-white">
                {repo.prs.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.linesAdded')}
              </div>
              <div className="text-lg font-bold text-green-600 dark:text-green-400">
                +{repo.lines_added.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.linesDeleted')}
              </div>
              <div className="text-lg font-bold text-red-600 dark:text-red-400">
                -{repo.lines_deleted.toLocaleString()}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
