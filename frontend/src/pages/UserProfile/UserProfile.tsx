/**
 * User Profile page - detailed profile and analytics for a user
 */
import React, { useState, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  ArrowLeftIcon,
  EnvelopeIcon,
  CodeBracketIcon,
  ArrowsRightLeftIcon,
  EyeIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { Avatar } from "../../components/common/Avatar";
import { Card } from "../../components/common/Card";
import { Loading } from "../../components/common/Loading";
import { Button } from "../../components/common/Button";
import { StatsCard } from "../../components/dashboard/StatsCard";
import { RecentActivitiesList } from "../../components/profile/RecentActivitiesList";
import { UserContributionsCard } from "../../components/profile/UserContributionsCard";
import ActivityTimeline from "../../components/charts/ActivityTimeline";
import Pagination from "../../components/common/Pagination";
import { useUserAnalytics } from "../../hooks/useAnalytics";
import { getDateRange, formatDate } from "../../utils/date";

type Period = "week" | "month" | "quarter";

const UserProfile: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [period, setPeriod] = useState<Period>("month");
  const [repoPage, setRepoPage] = useState(1);
  const repoPageSize = 25;

  const dateRange = useMemo(() => getDateRange(period), [period]);

  // Reset repo page when period changes
  useMemo(() => {
    setRepoPage(1);
  }, [period]);

  const {
    data: analytics,
    isLoading,
    error,
  } = useUserAnalytics(
    userId || "",
    dateRange.start,
    dateRange.end,
    repoPage,
    repoPageSize,
  );

  if (isLoading) {
    return <Loading text={t("common.loading")} />;
  }

  if (error || !analytics) {
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
      {/* Header with back button */}
      <div className="flex items-center gap-4">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => navigate(-1)}
          className="flex items-center gap-2"
        >
          <ArrowLeftIcon className="w-4 h-4" />
          {t("common.back")}
        </Button>
      </div>

      {/* Profile Header */}
      <Card padding="md">
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
          {/* Avatar */}
          <Avatar src={analytics.avatar_url} name={analytics.user_name} size="lg" />

          {/* User Info */}
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              {analytics.user_name}
            </h1>
            <div className="mt-2 space-y-1">
              <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                <EnvelopeIcon className="w-4 h-4" />
                {analytics.user_email}
              </div>
            </div>

            {/* Teams */}
            {analytics.teams && analytics.teams.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {analytics.teams.map((team) => (
                  <span
                    key={team}
                    className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary-100 dark:bg-primary-900 text-primary-800 dark:text-primary-200"
                  >
                    {team}
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Period Selector */}
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
      </Card>

      {/* Period info */}
      <div className="text-sm text-gray-600 dark:text-gray-400">
        {t("analytics.period")}: {formatDate(dateRange.start)} -{" "}
        {formatDate(dateRange.end)}
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title={t("analytics.commits")}
          value={analytics.total_commits}
          icon={<CodeBracketIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.pullRequests")}
          value={analytics.total_prs}
          icon={<ArrowsRightLeftIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.codeReviews")}
          value={analytics.total_reviews}
          icon={<EyeIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.issues")}
          value={analytics.total_issues}
          icon={<ExclamationTriangleIcon className="w-10 h-10" />}
        />
      </div>

      {/* Code Changes Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card padding="md">
          <div className="text-center">
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              {t("analytics.linesAdded")}
            </div>
            <div className="text-3xl font-bold text-green-600 dark:text-green-400">
              +{analytics.lines_added.toLocaleString()}
            </div>
          </div>
        </Card>
        <Card padding="md">
          <div className="text-center">
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              {t("analytics.linesDeleted")}
            </div>
            <div className="text-3xl font-bold text-red-600 dark:text-red-400">
              -{analytics.lines_deleted.toLocaleString()}
            </div>
          </div>
        </Card>
      </div>

      {/* Repository Contributions and Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Repository Contributions */}
        <Card
          title={t("profile.myRepositories")}
          subtitle={
            analytics.repository_contributions_pagination
              ? `Page ${analytics.repository_contributions_pagination.page} of ${analytics.repository_contributions_pagination.total_pages} (${analytics.repository_contributions_pagination.total} total)`
              : `${analytics.repository_contributions.length} repositories`
          }
          padding="md"
        >
          <UserContributionsCard
            contributions={analytics.repository_contributions}
            isLoading={isLoading}
          />
          {analytics.repository_contributions_pagination &&
            analytics.repository_contributions_pagination.total_pages > 1 && (
              <div className="mt-6 border-t border-gray-200 dark:border-gray-700 pt-4">
                <Pagination
                  currentPage={repoPage}
                  totalPages={
                    analytics.repository_contributions_pagination.total_pages
                  }
                  totalItems={
                    analytics.repository_contributions_pagination.total
                  }
                  itemsPerPage={repoPageSize}
                  onPageChange={setRepoPage}
                />
              </div>
            )}
        </Card>

        {/* Recent Activity */}
        <Card
          title={t("profile.recentActivity")}
          subtitle={`${analytics.recent_activities?.length || 0} activities`}
          padding="md"
        >
          <RecentActivitiesList
            activities={analytics.recent_activities || []}
            isLoading={isLoading}
          />
        </Card>
      </div>

      {/* Activity Timeline */}
      <Card title={t("analytics.activityTimeline")} padding="md">
        <ActivityTimeline
          data={analytics.activity_timeline || []}
          isLoading={isLoading}
          height={400}
        />
      </Card>
    </div>
  );
};

export default UserProfile;
