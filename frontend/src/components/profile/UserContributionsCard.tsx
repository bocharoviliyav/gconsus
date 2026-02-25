/**
 * User contributions card component for profile page
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import type { RepositoryContribution } from '../../services/api/analytics';

interface UserContributionsCardProps {
  contributions: RepositoryContribution[];
  isLoading?: boolean;
}

export const UserContributionsCard: React.FC<UserContributionsCardProps> = ({
  contributions,
  isLoading = false,
}) => {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-24 bg-gray-200 dark:bg-gray-700 rounded-lg" />
        ))}
      </div>
    );
  }

  if (contributions.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 dark:text-gray-500 text-6xl mb-4">📦</div>
        <p className="text-gray-600 dark:text-gray-400">{t('profile.noContributions')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {contributions.map((contribution, index) => (
        <div
          key={contribution.repository_name}
          className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700"
        >
          <div className="flex items-start justify-between mb-3">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                  #{index + 1}
                </span>
                <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                  {contribution.repository_name}
                </h4>
              </div>
              {contribution.repository_url && (
                <a
                  href={contribution.repository_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-primary-600 dark:text-primary-400 hover:underline"
                >
                  View repository
                </a>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.commits')}
              </div>
              <div className="text-lg font-bold text-gray-900 dark:text-white">
                {contribution.commits.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.pullRequests')}
              </div>
              <div className="text-lg font-bold text-gray-900 dark:text-white">
                {contribution.prs.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.linesAdded')}
              </div>
              <div className="text-lg font-bold text-green-600 dark:text-green-400">
                +{contribution.lines_added.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                {t('analytics.linesDeleted')}
              </div>
              <div className="text-lg font-bold text-red-600 dark:text-red-400">
                -{contribution.lines_deleted.toLocaleString()}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
