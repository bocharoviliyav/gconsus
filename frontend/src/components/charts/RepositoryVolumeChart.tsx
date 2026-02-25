/**
 * Repository Volume Chart component - shows repository activity volume over time
 */
import React from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { useTranslation } from "react-i18next";

interface RepositoryVolumeData {
  repository_name: string;
  total_commits: number;
  total_prs: number;
  contributors_count: number;
  lines_added: number;
  lines_deleted: number;
}

interface RepositoryVolumeChartProps {
  data: RepositoryVolumeData[];
  isLoading?: boolean;
  height?: number;
}

const RepositoryVolumeChart: React.FC<RepositoryVolumeChartProps> = ({
  data,
  isLoading = false,
  height = 400,
}) => {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div
        className="flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded-lg"
        style={{ height }}
      >
        <div className="text-gray-500 dark:text-gray-400">
          {t("common.loading")}...
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div
        className="flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded-lg"
        style={{ height }}
      >
        <div className="text-center text-gray-500 dark:text-gray-400">
          <div className="text-lg mb-2">📊</div>
          <div>{t("analytics.noData")}</div>
        </div>
      </div>
    );
  }

  // Format data for the chart - take top 10 repositories
  const chartData = data.slice(0, 10).map((repo) => ({
    name: repo.repository_name.length > 15
      ? repo.repository_name.substring(0, 15) + "..."
      : repo.repository_name,
    fullName: repo.repository_name,
    commits: repo.total_commits,
    prs: repo.total_prs,
    contributors: repo.contributors_count,
  }));

  const formatTooltipValue = (value: number, name: string) => {
    let label = "";
    switch (name) {
      case "commits":
        label = t("analytics.commits");
        break;
      case "prs":
        label = t("analytics.pullRequests");
        break;
      case "contributors":
        label = t("analytics.contributors");
        break;
      default:
        label = name;
    }
    return [value, label];
  };

  const formatTooltipLabel = (label: string) => {
    const entry = chartData.find((d) => d.name === label);
    return entry ? entry.fullName : label;
  };

  return (
    <div className="w-full">
      <ResponsiveContainer width="100%" height={height}>
        <BarChart
          data={chartData}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: 60,
          }}
        >
          <CartesianGrid
            strokeDasharray="3 3"
            className="stroke-gray-200 dark:stroke-gray-700"
          />
          <XAxis
            dataKey="name"
            tick={{ fontSize: 12 }}
            className="fill-gray-600 dark:fill-gray-400"
            angle={-45}
            textAnchor="end"
            height={80}
          />
          <YAxis
            tick={{ fontSize: 12 }}
            className="fill-gray-600 dark:fill-gray-400"
          />
          <Tooltip
            formatter={formatTooltipValue}
            labelFormatter={formatTooltipLabel}
            contentStyle={{
              backgroundColor: "var(--tooltip-bg, #fff)",
              border: "1px solid var(--tooltip-border, #e5e7eb)",
              borderRadius: "8px",
              boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
            }}
            labelStyle={{
              color: "var(--tooltip-label, #374151)",
              fontWeight: "bold",
              marginBottom: "4px",
            }}
          />
          <Legend
            wrapperStyle={{
              paddingTop: "20px",
            }}
          />
          <Bar
            dataKey="commits"
            name="commits"
            fill="#10b981"
            radius={[4, 4, 0, 0]}
          />
          <Bar
            dataKey="prs"
            name="prs"
            fill="#3b82f6"
            radius={[4, 4, 0, 0]}
          />
          <Bar
            dataKey="contributors"
            name="contributors"
            fill="#f59e0b"
            radius={[4, 4, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

export default RepositoryVolumeChart;
