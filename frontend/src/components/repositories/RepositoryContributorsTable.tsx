/**
 * Repository contributors table component
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { UsersIcon } from '@heroicons/react/24/outline';
import { Avatar } from '../common/Avatar';
import type { RepositoryContributor } from '../../services/api/repositories';

interface RepositoryContributorsTableProps {
  contributors: RepositoryContributor[];
  isLoading?: boolean;
}

export const RepositoryContributorsTable: React.FC<RepositoryContributorsTableProps> = ({
  contributors,
  isLoading = false,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-4">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-16 bg-gray-200 dark:bg-gray-700 rounded-lg" />
        ))}
      </div>
    );
  }

  if (contributors.length === 0) {
    return (
      <div className="text-center py-12">
        <UsersIcon className="mx-auto h-16 w-16 text-gray-400 dark:text-gray-500 mb-4" />
        <p className="text-gray-600 dark:text-gray-400 text-lg">
          {t('analytics.noData')}
        </p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead className="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.contributor')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.commits')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.pullRequests')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.codeReviews')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.linesAdded')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t('analytics.linesDeleted')}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Contribution %
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          {contributors.map((contributor) => (
            <tr
              key={contributor.user_id}
              className="hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors cursor-pointer"
              onClick={() => navigate(`/users/${contributor.user_id}`)}
            >
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="flex items-center">
                  <Avatar src={contributor.avatar_url} name={contributor.user_name} className="mr-3" />
                  <div>
                    <div className="text-sm font-medium text-gray-900 dark:text-white">
                      {contributor.user_name}
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      {contributor.user_email}
                    </div>
                  </div>
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {contributor.commits.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {contributor.prs.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {contributor.reviews.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm text-green-600 dark:text-green-400 font-medium">
                  +{contributor.lines_added.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm text-red-600 dark:text-red-400 font-medium">
                  -{contributor.lines_deleted.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <div className="flex items-center justify-center gap-2">
                  <div className="flex-1 max-w-[100px]">
                    <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-primary-600 dark:bg-primary-400 rounded-full"
                        style={{ width: `${Math.min(contributor.percentage, 100)}%` }}
                      />
                    </div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                    {contributor.percentage.toFixed(1)}%
                  </span>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
