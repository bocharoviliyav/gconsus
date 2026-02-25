import React from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { useTranslation } from "react-i18next";
import type { ActivityTimelineEntry } from "../../types";

interface ActivityTimelineProps {
  data: ActivityTimelineEntry[];
  isLoading?: boolean;
  height?: number;
}

const ActivityTimeline: React.FC<ActivityTimelineProps> = ({
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
          <div className="text-lg mb-2">📈</div>
          <div>{t("analytics.noActivityData")}</div>
        </div>
      </div>
    );
  }

  // Determine if commits are on a very different scale from other metrics.
  // If so, use dual Y-axes so PRs/reviews/issues don't get squished.
  const maxCommits = Math.max(...data.map((d) => d.commits), 0);
  const maxOther = Math.max(
    ...data.map((d) => Math.max(d.prs, d.reviews, d.issues)),
    0,
  );
  const useDualAxis = maxCommits > 0 && maxOther > 0 && maxCommits / maxOther > 3;

  // Format data for the chart
  const chartData = data.map((entry) => ({
    ...entry,
    // Format date for display (show only day-month for better readability)
    displayDate: new Date(entry.date).toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
    }),
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
      case "reviews":
        label = t("analytics.codeReviews");
        break;
      case "issues":
        label = t("analytics.issues");
        break;
      default:
        label = name;
    }
    return [value, label];
  };

  const formatTooltipLabel = (label: string) => {
    const entry = data.find((d) => d.date.includes(label.replace(/\./g, "-")));
    if (entry) {
      return new Date(entry.date).toLocaleDateString("ru-RU", {
        weekday: "long",
        day: "2-digit",
        month: "long",
        year: "numeric",
      });
    }
    return label;
  };

  return (
    <div className="w-full">
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={chartData}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: 5,
          }}
        >
          <CartesianGrid
            strokeDasharray="3 3"
            className="stroke-gray-200 dark:stroke-gray-700"
          />
          <XAxis
            dataKey="displayDate"
            tick={{ fontSize: 12 }}
            className="fill-gray-600 dark:fill-gray-400"
            interval="preserveStartEnd"
          />
          <YAxis
            yAxisId="left"
            tick={{ fontSize: 12 }}
            className="fill-gray-600 dark:fill-gray-400"
          />
          {useDualAxis && (
            <YAxis
              yAxisId="right"
              orientation="right"
              tick={{ fontSize: 12 }}
              className="fill-gray-600 dark:fill-gray-400"
            />
          )}
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
          <Line
            type="monotone"
            dataKey="commits"
            name="commits"
            yAxisId="left"
            stroke="#10b981"
            strokeWidth={2}
            dot={{ fill: "#10b981", strokeWidth: 2, r: 3 }}
            activeDot={{ r: 5, stroke: "#10b981", strokeWidth: 2 }}
          />
          <Line
            type="monotone"
            dataKey="prs"
            name="prs"
            yAxisId={useDualAxis ? "right" : "left"}
            stroke="#3b82f6"
            strokeWidth={2}
            dot={{ fill: "#3b82f6", strokeWidth: 2, r: 3 }}
            activeDot={{ r: 5, stroke: "#3b82f6", strokeWidth: 2 }}
          />
          <Line
            type="monotone"
            dataKey="reviews"
            name="reviews"
            yAxisId={useDualAxis ? "right" : "left"}
            stroke="#f59e0b"
            strokeWidth={2}
            dot={{ fill: "#f59e0b", strokeWidth: 2, r: 3 }}
            activeDot={{ r: 5, stroke: "#f59e0b", strokeWidth: 2 }}
          />
          <Line
            type="monotone"
            dataKey="issues"
            name="issues"
            yAxisId={useDualAxis ? "right" : "left"}
            stroke="#ef4444"
            strokeWidth={2}
            dot={{ fill: "#ef4444", strokeWidth: 2, r: 3 }}
            activeDot={{ r: 5, stroke: "#ef4444", strokeWidth: 2 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

export default ActivityTimeline;
