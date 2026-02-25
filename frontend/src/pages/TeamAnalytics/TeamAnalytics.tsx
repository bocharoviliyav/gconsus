/**
 * Team Analytics page - detailed analytics for a specific team
 */
import React, { useState, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  ArrowLeftIcon,
  CodeBracketIcon,
  ArrowsRightLeftIcon,
  EyeIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { Card } from "../../components/common/Card";
import { Loading } from "../../components/common/Loading";
import { Button } from "../../components/common/Button";
import { StatsCard } from "../../components/dashboard/StatsCard";
import { MembersStatsTable } from "../../components/analytics/MembersStatsTable";
import { RepositoryStatsCard } from "../../components/analytics/RepositoryStatsCard";
import ActivityTimeline from "../../components/charts/ActivityTimeline";
import Pagination from "../../components/common/Pagination";
import { useTeam } from "../../hooks/useTeams";
import { useTeamAnalytics } from "../../hooks/useAnalytics";
import { getDateRange, formatDate } from "../../utils/date";

type Period = "week" | "month" | "quarter";

const TeamAnalytics: React.FC = () => {
  const { teamId } = useParams<{ teamId: string }>();
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

  const { data: team, isLoading: isTeamLoading } = useTeam(teamId || "");
  const {
    data: analytics,
    isLoading: isAnalyticsLoading,
    error,
  } = useTeamAnalytics(
    teamId || "",
    dateRange.start,
    dateRange.end,
    repoPage,
    repoPageSize,
  );

  const isLoading = isTeamLoading || isAnalyticsLoading;

  if (isLoading) {
    return <Loading text={t("common.loading")} />;
  }

  if (error || !team) {
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
          onClick={() => navigate("/teams")}
          className="flex items-center gap-2"
        >
          <ArrowLeftIcon className="w-4 h-4" />
          {t("common.back")}
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {team.name}
          </h1>
          {team.description && (
            <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
              {team.description}
            </p>
          )}
          {analytics?.lead_name && (
            <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
              {t("teams.teamLead")}: {analytics.lead_name}
            </p>
          )}
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

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title={t("analytics.commits")}
          value={analytics?.total_commits || 0}
          icon={<CodeBracketIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.pullRequests")}
          value={analytics?.total_prs || 0}
          icon={<ArrowsRightLeftIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.codeReviews")}
          value={analytics?.total_reviews || 0}
          icon={<EyeIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("analytics.issues")}
          value={analytics?.total_issues || 0}
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
              +{analytics?.lines_added.toLocaleString() || 0}
            </div>
          </div>
        </Card>
        <Card padding="md">
          <div className="text-center">
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              {t("analytics.linesDeleted")}
            </div>
            <div className="text-3xl font-bold text-red-600 dark:text-red-400">
              -{analytics?.lines_deleted.toLocaleString() || 0}
            </div>
          </div>
        </Card>
      </div>

      {/* Members Stats */}
      <Card
        title={t("analytics.teamAnalytics")}
        subtitle={`${analytics?.member_stats.length || 0} ${t("teams.members")}`}
        padding="none"
      >
        <div className="p-4">
          <MembersStatsTable
            members={analytics?.member_stats || []}
            isLoading={isAnalyticsLoading}
          />
        </div>
      </Card>

      {/* Repository Stats */}
      <Card
        title={t("analytics.topRepositories")}
        subtitle={
          analytics?.repository_stats_pagination
            ? `Page ${analytics.repository_stats_pagination.page} of ${analytics.repository_stats_pagination.total_pages} (${analytics.repository_stats_pagination.total} total)`
            : `${analytics?.repository_stats.length || 0} repositories`
        }
        padding="md"
      >
        <RepositoryStatsCard
          repositories={analytics?.repository_stats || []}
          isLoading={isAnalyticsLoading}
        />
        {analytics?.repository_stats_pagination &&
          analytics.repository_stats_pagination.total_pages > 1 && (
            <div className="mt-6 border-t border-gray-200 dark:border-gray-700 pt-4">
              <Pagination
                currentPage={repoPage}
                totalPages={analytics.repository_stats_pagination.total_pages}
                totalItems={analytics.repository_stats_pagination.total}
                itemsPerPage={repoPageSize}
                onPageChange={setRepoPage}
              />
            </div>
          )}
      </Card>

      {/* Activity Timeline */}
      <Card title={t("analytics.activityTimeline")} padding="md">
        <ActivityTimeline
          data={analytics?.activity_timeline || []}
          isLoading={isAnalyticsLoading}
          height={400}
        />
      </Card>
    </div>
  );
};

export default TeamAnalytics;
