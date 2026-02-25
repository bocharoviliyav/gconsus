import React from "react";
import { Card } from "../common/Card";
import { formatCompactNumber } from "../../utils/format";

interface StatsCardProps {
  title: string;
  value: number;
  icon?: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
}

export const StatsCard: React.FC<StatsCardProps> = ({
  title,
  value,
  icon,
  trend,
}) => {
  return (
    <Card padding="md" className="hover:shadow-lg transition-shadow">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
            {title}
          </p>
          <p className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">
            {formatCompactNumber(value)}
          </p>
          {trend && (
            <p
              className={`mt-2 text-sm ${trend.isPositive ? "text-green-600" : "text-red-600"}`}
            >
              {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
            </p>
          )}
        </div>
        {icon && <div className="text-primary-500 dark:text-primary-400 opacity-80">{icon}</div>}
      </div>
    </Card>
  );
};
