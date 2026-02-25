/**
 * Recent activities list component for user profile
 */
import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  CodeBracketIcon,
  ArrowsRightLeftIcon,
  EyeIcon,
  ExclamationTriangleIcon,
  InboxIcon,
} from '@heroicons/react/24/outline';
import { getRelativeTime } from '../../utils/date';
import type { RecentActivity } from '../../services/api/analytics';

interface RecentActivitiesListProps {
  activities: RecentActivity[];
  isLoading?: boolean;
}

const activityIcons: Record<RecentActivity['type'], React.FC<React.SVGProps<SVGSVGElement>>> = {
  commit: CodeBracketIcon,
  pr: ArrowsRightLeftIcon,
  review: EyeIcon,
  issue: ExclamationTriangleIcon,
};

const activityColors: Record<RecentActivity['type'], string> = {
  commit: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
  pr: 'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200',
  review: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
  issue: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
};

export const RecentActivitiesList: React.FC<RecentActivitiesListProps> = ({
  activities,
  isLoading = false,
}) => {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-16 bg-gray-200 dark:bg-gray-700 rounded-lg" />
        ))}
      </div>
    );
  }

  if (activities.length === 0) {
    return (
      <div className="text-center py-12">
        <InboxIcon className="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500 mb-4" />
        <p className="text-gray-600 dark:text-gray-400">{t('profile.noRecentActivity')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {activities.map((activity) => (
        <div
          key={activity.id}
          className="flex items-start gap-3 p-4 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div
            className={`flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center ${
              activityColors[activity.type]
            }`}
          >
            {React.createElement(activityIcons[activity.type], { className: 'w-5 h-5' })}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {activity.title}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {activity.repository_name}
                </p>
              </div>
              <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                {getRelativeTime(activity.created_at)}
              </span>
            </div>
            {activity.url && (
              <a
                href={activity.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-primary-600 dark:text-primary-400 hover:underline mt-1 inline-block"
              >
                View details
              </a>
            )}
          </div>
        </div>
      ))}
    </div>
  );
};
