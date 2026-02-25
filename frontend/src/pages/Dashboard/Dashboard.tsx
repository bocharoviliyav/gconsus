import React, { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import {
  CodeBracketIcon,
  ArrowsRightLeftIcon,
  EyeIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { Card } from "../../components/common/Card";
import { Loading } from "../../components/common/Loading";
import { StatsCard } from "../../components/dashboard/StatsCard";
import { LeaderboardTable } from "../../components/dashboard/LeaderboardTable";
import Pagination from "../../components/common/Pagination";
import ActivityTimeline from "../../components/charts/ActivityTimeline";
import { useDashboard } from "../../hooks/useDashboard";
import { getDateRange } from "../../utils/date";

type Period = "week" | "month" | "quarter";
const TEAMS_PAGE_SIZE = 5;

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [period, setPeriod] = useState<Period>("month");
  const [teamsPage, setTeamsPage] = useState(1);

  const dateRange = useMemo(() => getDateRange(period), [period]);
  const { data, isLoading, error } = useDashboard(
    dateRange.start,
    dateRange.end,
  );

  const allTeams = data?.top_teams || [];
  const teamsTotalPages = Math.ceil(allTeams.length / TEAMS_PAGE_SIZE);
  const paginatedTeams = useMemo(() => {
    const start = (teamsPage - 1) * TEAMS_PAGE_SIZE;
    return allTeams.slice(start, start + TEAMS_PAGE_SIZE);
  }, [allTeams, teamsPage]);

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
      {/* Header with period selector */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {t("dashboard.title")}
          </h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            {t("dashboard.overview")}
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

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title={t("dashboard.totalCommits")}
          value={data?.total_commits || 0}
          icon={<CodeBracketIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("dashboard.totalPRs")}
          value={data?.total_prs || 0}
          icon={<ArrowsRightLeftIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("dashboard.totalReviews")}
          value={data?.total_reviews || 0}
          icon={<EyeIcon className="w-10 h-10" />}
        />
        <StatsCard
          title={t("dashboard.totalIssues")}
          value={data?.total_issues || 0}
          icon={<ExclamationTriangleIcon className="w-10 h-10" />}
        />
      </div>

      {/* Leaderboards */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card
          title={t("dashboard.topContributors")}
          subtitle={`${t("analytics.period")}: ${period === "week" ? t("analytics.last7Days") : period === "month" ? t("analytics.last30Days") : t("analytics.last90Days")}`}
          padding="none"
        >
          <div className="p-4">
            <LeaderboardTable
              entries={data?.top_contributors || []}
              isLoading={isLoading}
            />
          </div>
        </Card>

        <Card
          title={t("dashboard.topTeams")}
          subtitle={`${t("analytics.period")}: ${period === "week" ? t("analytics.last7Days") : period === "month" ? t("analytics.last30Days") : t("analytics.last90Days")}`}
          padding="md"
        >
          {allTeams.length > 0 ? (
            <div>
              <div className="space-y-3">
                {paginatedTeams.map((team) => (
                  <div
                    key={team.team_id}
                    className="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors group"
                    onClick={() => navigate(`/teams/${team.team_id}`)}
                    title={`View ${team.team_name} team profile`}
                  >
                    <div>
                      <p className="font-medium text-gray-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                        {team.team_name}
                      </p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        {team.member_count} {t("teams.members")}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="text-right">
                        <p className="text-lg font-bold text-primary-600 dark:text-primary-400">
                          {team.total_commits}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {t("analytics.commits")}
                        </p>
                      </div>
                      <div className="text-gray-400 dark:text-gray-500 group-hover:text-primary-500 dark:group-hover:text-primary-400 transition-colors">
                        →
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              {teamsTotalPages > 1 && (
                <div className="pt-3">
                  <Pagination
                    currentPage={teamsPage}
                    totalPages={teamsTotalPages}
                    totalItems={allTeams.length}
                    itemsPerPage={TEAMS_PAGE_SIZE}
                    onPageChange={setTeamsPage}
                  />
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              {t("analytics.noData")}
            </div>
          )}
        </Card>
      </div>

      {/* Activity Trend */}
      <Card title={t("dashboard.activityTrend")} padding="md">
        <ActivityTimeline
          data={data?.activity_trend || []}
          isLoading={isLoading}
          height={400}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
