/**
 * Repositories Analytics page - leaderboard and analytics for repositories
 */
import React, { useState, useMemo, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Card } from "../../components/common/Card";
import { Loading } from "../../components/common/Loading";
import { RepositoryLeaderboardTable } from "../../components/repositories/RepositoryLeaderboardTable";

import ActivityTimeline from "../../components/charts/ActivityTimeline";
import Pagination from "../../components/common/Pagination";
import { useRepositoriesLeaderboard } from "../../hooks/useRepositories";
import { getDateRange, formatDate } from "../../utils/date";

type Period = "week" | "month" | "quarter";

const RepositoriesAnalytics: React.FC = () => {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<Period>("month");
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 25;

  const dateRange = useMemo(() => getDateRange(period), [period]);

  const { data, isLoading, error } = useRepositoriesLeaderboard(
    dateRange.start,
    dateRange.end,
    currentPage,
    pageSize,
  );

  // Reset page when period changes
  useEffect(() => {
    setCurrentPage(1);
  }, [period]);

  if (isLoading) {
    return <Loading text={t("common.loading")} />;
  }

  if (error) {
    return (
      <Card>
        <div className="text-center py-8 text-red-600 dark:text-red-400">
          {t("errors.generic")}
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {t("analytics.repositoryAnalytics")}
          </h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            Repository activity leaderboard and analytics
          </p>
        </div>
        <div className="flex gap-2">
          {(["week", "month", "quarter"] as Period[]).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                period === p
                  ? "bg-primary-600 text-white"
                  : "bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600"
              }`}
            >
              {p === "week" && t("analytics.last7Days")}
              {p === "month" && t("analytics.last30Days")}
              {p === "quarter" && t("analytics.last90Days")}
            </button>
          ))}
        </div>
      </div>

      {/* Period info */}
      <div className="text-sm text-gray-600 dark:text-gray-400">
        {t("analytics.period")}: {formatDate(dateRange.start)} -{" "}
        {formatDate(dateRange.end)}
      </div>

      {/* Summary Stats */}
      {data && (
        <div className="grid grid-cols-4 md:grid-cols-2 gap-4">
          <Card padding="md">
            <div className="text-center">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                {t("analytics.linesAdded")}
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {(data.total_lines_added ?? 0).toLocaleString()}
              </div>
            </div>
          </Card>
          <Card padding="md">
            <div className="text-center">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                {t("analytics.linesDeleted")}
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {(data.total_lines_deleted ?? 0).toLocaleString()}
              </div>
            </div>
          </Card>
          <Card padding="md">
            <div className="text-center">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                {t("analytics.commits")}
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {(data.total_commits ?? 0).toLocaleString()}
              </div>
            </div>
          </Card>
          <Card padding="md">
            <div className="text-center">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                {t("analytics.contributors")}
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {(data.total_contributors ?? 0).toLocaleString()}
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Repository Leaderboard */}
      <Card
        title={t("analytics.topRepositories")}
        subtitle={`${t("common.page")} ${currentPage} / ${data?.total_pages || 1} (${data?.total || 0})`}
        padding="none"
      >
        <div className="p-4">
          <RepositoryLeaderboardTable
            repositories={data?.repositories || []}
            isLoading={isLoading}
          />
          {data && data.total_pages > 1 && (
            <div className="mt-6 border-t border-gray-200 dark:border-gray-700 pt-4">
              <Pagination
                currentPage={currentPage}
                totalPages={data.total_pages}
                totalItems={data.total}
                itemsPerPage={pageSize}
                onPageChange={setCurrentPage}
              />
            </div>
          )}
        </div>
      </Card>

      {/* Repository Activity Timeline */}
      <Card
        title={t("analytics.repositoryActivityTimeline")}
        subtitle={t("analytics.combinedRepositoryActivity")}
        padding="md"
      >
        <ActivityTimeline
          data={data?.repository_activity_timeline || []}
          isLoading={isLoading}
          height={400}
        />
      </Card>
    </div>
  );
};

export default RepositoriesAnalytics;
