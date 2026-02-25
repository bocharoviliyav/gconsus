/**
 * Repository leaderboard table component
 */
import React from "react";
import { useTranslation } from "react-i18next";
// import { useNavigate } from "react-router-dom";
import type { RepositoryLeaderboardItem } from "../../services/api/repositories";

interface RepositoryLeaderboardTableProps {
  repositories: RepositoryLeaderboardItem[];
  isLoading?: boolean;
}

export const RepositoryLeaderboardTable: React.FC<
  RepositoryLeaderboardTableProps
> = ({ repositories, isLoading = false }) => {
  const { t } = useTranslation();
  // const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-4">
        {[...Array(10)].map((_, i) => (
          <div
            key={i}
            className="h-16 bg-gray-200 dark:bg-gray-700 rounded-lg"
          />
        ))}
      </div>
    );
  }

  if (repositories.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 dark:text-gray-500 text-6xl mb-4">📦</div>
        <p className="text-gray-600 dark:text-gray-400 text-lg">
          {t("analytics.noData")}
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
              {t("analytics.rank")}
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.repository")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.commits")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.pullRequests")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.contributors")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.linesAdded")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {t("analytics.linesDeleted")}
            </th>
            <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Score
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          {repositories.map((repo) => (
            <tr
              key={repo.repository_name}
              className="hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors cursor-pointer"
              // onClick={() => navigate(`/repositories/${repo.repository_name}`)}
            >
              <td className="px-6 py-4 whitespace-nowrap">
                <span
                  className={`inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold ${
                    repo.rank === 1
                      ? "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200"
                      : repo.rank === 2
                        ? "bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200"
                        : repo.rank === 3
                          ? "bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200"
                          : "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400"
                  }`}
                >
                  {repo.rank}
                </span>
              </td>
              <td className="px-6 py-4">
                <div>
                  <div className="text-sm font-medium text-gray-900 dark:text-white">
                    {repo.repository_name}
                  </div>
                  {repo.repository_url && (
                    <a
                      href={repo.repository_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-xs text-primary-600 dark:text-primary-400 hover:underline"
                      onClick={(e) => e.stopPropagation()}
                    >
                      View repository
                    </a>
                  )}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {repo.total_commits.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {repo.total_prs.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {repo.contributors_count}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm text-green-600 dark:text-green-400 font-medium">
                  +{repo.lines_added.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="text-sm text-red-600 dark:text-red-400 font-medium">
                  -{repo.lines_deleted.toLocaleString()}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-center">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 dark:bg-primary-900 text-primary-800 dark:text-primary-200">
                  {repo.activity_score.toFixed(1)}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
