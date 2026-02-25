import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card } from "../../components/common/Card";
import { Loading } from "../../components/common/Loading";
import { GitConfigSection } from "../../components/settings/GitConfigSection";
import { SyncStatusCard } from "../../components/settings/SyncStatusCard";
import {
  useSystemSettings,
  useUpdateSystemSettings,
  useSyncStatus,
  useTriggerSync,
} from "../../hooks/useSettings";
import {
  testGitHubConnection,
  testGitLabConnection,
} from "../../services/api/settings";

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const [syncSchedule, setSyncSchedule] = useState("");
  const [syncEnabled, setSyncEnabled] = useState(true);
  const [hasScheduleChanges, setHasScheduleChanges] = useState(false);

  const { data: settings, isLoading: isLoadingSettings } = useSystemSettings();
  const { data: syncStatus } = useSyncStatus();
  const updateSettingsMutation = useUpdateSystemSettings();
  const triggerSyncMutation = useTriggerSync();

  React.useEffect(() => {
    if (settings) {
      setSyncSchedule(settings.sync_schedule_cron);
      setSyncEnabled(settings.sync_enabled);
    }
  }, [settings]);

  const handleGitHubUpdate = async (data: {
    enabled: boolean;
    url: string;
    token?: string;
  }) => {
    await updateSettingsMutation.mutateAsync({
      github_enabled: data.enabled,
      github_url: data.url,
      github_token: data.token,
      // Mutual exclusion: disable GitLab when enabling GitHub
      ...(data.enabled && { gitlab_enabled: false }),
    });
  };

  const handleGitLabUpdate = async (data: {
    enabled: boolean;
    url: string;
    token?: string;
  }) => {
    await updateSettingsMutation.mutateAsync({
      gitlab_enabled: data.enabled,
      gitlab_url: data.url,
      gitlab_token: data.token,
      // Mutual exclusion: disable GitHub when enabling GitLab
      ...(data.enabled && { github_enabled: false }),
    });
  };

  const handleScheduleUpdate = async () => {
    await updateSettingsMutation.mutateAsync({
      sync_schedule_cron: syncSchedule,
      sync_enabled: syncEnabled,
    });
    setHasScheduleChanges(false);
  };

  const handleScheduleChange = (value: string) => {
    setSyncSchedule(value);
    setHasScheduleChanges(true);
  };

  const handleSyncEnabledChange = (checked: boolean) => {
    setSyncEnabled(checked);
    setHasScheduleChanges(true);
  };

  const handleTriggerSync = async () => {
    await triggerSyncMutation.mutateAsync();
  };

  if (isLoadingSettings) {
    return <Loading text={t("common.loading")} />;
  }

  if (!settings) {
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
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          {t("settings.title")}
        </h1>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Configure application settings and integrations
        </p>
      </div>

      {/* Sync Status */}
      {syncStatus && (
        <Card title="Data Synchronization" padding="md">
          <SyncStatusCard
            status={syncStatus}
            onTriggerSync={handleTriggerSync}
            isTriggering={triggerSyncMutation.isPending}
          />
        </Card>
      )}

      {/* Sync Schedule */}
      <Card title={t("settings.schedule")} padding="md">
        <div className="space-y-4">
          {/* Enable/Disable Sync */}
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-medium text-gray-900 dark:text-white">
                Automatic Sync
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {syncEnabled ? "Enabled" : "Disabled"}
              </p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={syncEnabled}
                onChange={(e) => handleSyncEnabledChange(e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary-600"></div>
            </label>
          </div>

          {/* Cron Schedule */}
          {syncEnabled && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Cron Schedule
              </label>
              <input
                type="text"
                value={syncSchedule}
                onChange={(e) => handleScheduleChange(e.target.value)}
                placeholder="0 */6 * * *"
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                  placeholder-gray-400 dark:placeholder-gray-500
                  focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Example: "0 */6 * * *" runs every 6 hours
              </p>
            </div>
          )}

          {hasScheduleChanges && (
            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                onClick={() => {
                  setSyncSchedule(settings.sync_schedule_cron);
                  setSyncEnabled(settings.sync_enabled);
                  setHasScheduleChanges(false);
                }}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-200 dark:bg-gray-700 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                {t("common.cancel")}
              </button>
              <button
                onClick={handleScheduleUpdate}
                disabled={updateSettingsMutation.isPending}
                className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 transition-colors disabled:opacity-50"
              >
                {updateSettingsMutation.isPending
                  ? t("common.loading")
                  : t("common.save")}
              </button>
            </div>
          )}
        </div>
      </Card>

      {/* GitHub Configuration */}
      <Card title={t("settings.github")} padding="md">
        <GitConfigSection
          title="GitHub Integration"
          enabled={settings.github_enabled}
          url={settings.github_url || ""}
          tokenSet={settings.github_token_set}
          onUpdate={handleGitHubUpdate}
          onTest={testGitHubConnection}
          isLoading={updateSettingsMutation.isPending}
          exclusiveHint={
            settings.gitlab_enabled
              ? t("settings.providerExclusiveHint", { other: "GitLab" })
              : undefined
          }
        />
      </Card>

      {/* GitLab Configuration */}
      <Card title={t("settings.gitlab")} padding="md">
        <GitConfigSection
          title="GitLab Integration"
          enabled={settings.gitlab_enabled}
          url={settings.gitlab_url || ""}
          tokenSet={settings.gitlab_token_set}
          onUpdate={handleGitLabUpdate}
          onTest={testGitLabConnection}
          isLoading={updateSettingsMutation.isPending}
          exclusiveHint={
            settings.github_enabled
              ? t("settings.providerExclusiveHint", { other: "GitHub" })
              : undefined
          }
        />
      </Card>
    </div>
  );
};

export default Settings;
