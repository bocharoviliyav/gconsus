/**
 * Sync status card component for Settings page
 */
import React from "react";
import { Button } from "../common/Button";
import {
  ArrowPathIcon,
  CheckCircleIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { getRelativeTime } from "../../utils/date";
import type { SyncStatus } from "../../services/api/settings";

interface SyncStatusCardProps {
  status: SyncStatus;
  onTriggerSync: () => void;
  isTriggering?: boolean;
}

export const SyncStatusCard: React.FC<SyncStatusCardProps> = ({
  status,
  onTriggerSync,
  isTriggering = false,
}) => {
  return (
    <div className="space-y-4">
      {/* Status Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            Sync Status
          </h3>
          {status.is_running ? (
            <div className="flex items-center gap-2 mt-1">
              <ArrowPathIcon className="w-4 h-4 animate-spin text-primary-600 dark:text-primary-400" />
              <span className="text-sm text-primary-600 dark:text-primary-400">
                Sync in progress...
              </span>
            </div>
          ) : status.last_error ? (
            <div className="flex items-center gap-2 mt-1">
              <XCircleIcon className="w-4 h-4 text-red-600 dark:text-red-400" />
              <span className="text-sm text-red-600 dark:text-red-400">
                Last sync failed
              </span>
            </div>
          ) : (
            <div className="flex items-center gap-2 mt-1">
              <CheckCircleIcon className="w-4 h-4 text-green-600 dark:text-green-400" />
              <span className="text-sm text-green-600 dark:text-green-400">
                All systems operational
              </span>
            </div>
          )}
        </div>
        <Button
          variant="primary"
          onClick={onTriggerSync}
          isLoading={isTriggering || status.is_running}
          disabled={status.is_running}
          className="flex items-center gap-2"
        >
          <ArrowPathIcon className="w-4 h-4" />
          Trigger Sync
        </Button>
      </div>

      {/* Sync Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">
            Users Synced
          </div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {status.users_synced.toLocaleString()}
          </div>
        </div>
        <div className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">
            Teams Synced
          </div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {status.teams_synced.toLocaleString()}
          </div>
        </div>
        <div className="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">
            Activities Synced
          </div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {status.activities_synced.toLocaleString()}
          </div>
        </div>
      </div>

      {/* Last Sync Info */}
      <div className="space-y-2">
        {status.started_at && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">Started:</span>
            <span className="text-gray-900 dark:text-white">
              {getRelativeTime(status.started_at)}
            </span>
          </div>
        )}
        {status.last_completed_at && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">
              Last Completed:
            </span>
            <span className="text-gray-900 dark:text-white">
              {getRelativeTime(status.last_completed_at)}
            </span>
          </div>
        )}
        {status.last_error && (
          <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
            <div className="text-sm font-medium text-red-800 dark:text-red-200 mb-1">
              Last Error:
            </div>
            <div className="text-sm text-red-600 dark:text-red-400">
              {status.last_error}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
