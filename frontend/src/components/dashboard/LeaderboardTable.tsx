import React, { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import Pagination from "../common/Pagination";
import { Avatar } from "../common/Avatar";
import type { LeaderboardEntry } from "../../types";
import { formatNumber } from "../../utils/format";

const PAGE_SIZE = 5;

interface LeaderboardTableProps {
  entries: LeaderboardEntry[];
  isLoading?: boolean;
}

export const LeaderboardTable: React.FC<LeaderboardTableProps> = ({
  entries,
  isLoading,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(1);

  const totalPages = Math.ceil((entries?.length || 0) / PAGE_SIZE);
  const paginatedEntries = useMemo(() => {
    const start = (page - 1) * PAGE_SIZE;
    return (entries || []).slice(start, start + PAGE_SIZE);
  }, [entries, page]);

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-12 bg-gray-200 dark:bg-gray-700 rounded" />
        ))}
      </div>
    );
  }

  if (!entries || entries.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        {t("analytics.noData")}
      </div>
    );
  }

  return (
    <div>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                {t("analytics.rank")}
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                {t("analytics.contributor")}
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                {t("analytics.commits")}
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                {t("analytics.pullRequests")}
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                {t("analytics.codeReviews")}
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
            {paginatedEntries.map((entry) => (
              <tr
                key={entry.user_id}
                className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer transition-colors group"
                onClick={() => navigate(`/users/${entry.user_id}`)}
                title={`View ${entry.user_name}'s profile`}
              >
                <td className="px-4 py-3 whitespace-nowrap">
                  <span
                    className={`inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold
                    ${entry.rank === 1 ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200" : ""}
                    ${entry.rank === 2 ? "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200" : ""}
                    ${entry.rank === 3 ? "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200" : ""}
                    ${entry.rank > 3 ? "text-gray-600 dark:text-gray-400" : ""}
                  `}
                  >
                    {entry.rank}
                  </span>
                </td>
                <td className="px-4 py-3 whitespace-nowrap">
                  <div className="flex items-center">
                    <Avatar src={entry.avatar_url} name={entry.user_name || "?"} />
                    <div className="ml-3">
                      <p className="text-sm font-medium text-gray-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                        {entry.user_name}
                      </p>
                      {entry.github_username && (
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          @{entry.github_username}
                        </p>
                      )}
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-right text-sm text-gray-900 dark:text-white">
                  {formatNumber(entry.commits)}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-right text-sm text-gray-900 dark:text-white">
                  {formatNumber(entry.prs)}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-right text-sm text-gray-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                  {formatNumber(entry.reviews)}{" "}
                  <span className="ml-2 text-gray-400 dark:text-gray-500 group-hover:text-primary-500 dark:group-hover:text-primary-400 transition-colors">
                    →
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {totalPages > 1 && (
        <div className="pt-3">
          <Pagination
            currentPage={page}
            totalPages={totalPages}
            totalItems={entries.length}
            itemsPerPage={PAGE_SIZE}
            onPageChange={setPage}
          />
        </div>
      )}
    </div>
  );
};
