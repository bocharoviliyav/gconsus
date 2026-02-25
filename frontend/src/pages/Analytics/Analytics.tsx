import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card } from '../../components/common/Card';

const Analytics: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          {t('analytics.title')}
        </h1>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          View detailed analytics and metrics
        </p>
      </div>

      <Card>
        <div className="text-center py-16 text-gray-500 dark:text-gray-400">
          Analytics coming soon...
        </div>
      </Card>
    </div>
  );
};

export default Analytics;
