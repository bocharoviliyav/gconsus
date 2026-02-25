/**
 * Git configuration section for Settings page
 */
import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../common/Button';
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';

interface GitConfigSectionProps {
  title: string;
  enabled: boolean;
  url: string;
  tokenSet: boolean;
  onUpdate: (data: { enabled: boolean; url: string; token?: string }) => void;
  onTest: (url: string, token: string) => Promise<boolean>;
  isLoading?: boolean;
  exclusiveHint?: string;
}

export const GitConfigSection: React.FC<GitConfigSectionProps> = ({
  title,
  enabled,
  url,
  tokenSet,
  onUpdate,
  onTest,
  isLoading = false,
  exclusiveHint,
}) => {
  const { t } = useTranslation();
  const [localEnabled, setLocalEnabled] = useState(enabled);
  const [localUrl, setLocalUrl] = useState(url);
  const [localToken, setLocalToken] = useState('');
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null);
  const [isTesting, setIsTesting] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  // Sync local state when props change (e.g. after mutual exclusion toggle)
  React.useEffect(() => {
    setLocalEnabled(enabled);
    setLocalUrl(url);
    setHasChanges(false);
  }, [enabled, url]);

  const handleEnabledChange = (checked: boolean) => {
    setLocalEnabled(checked);
    setHasChanges(true);
  };

  const handleUrlChange = (value: string) => {
    setLocalUrl(value);
    setHasChanges(true);
    setTestResult(null);
  };

  const handleTokenChange = (value: string) => {
    setLocalToken(value);
    setHasChanges(true);
    setTestResult(null);
  };

  const handleTest = async () => {
    if (!localUrl || !localToken) return;
    
    setIsTesting(true);
    setTestResult(null);
    
    try {
      const success = await onTest(localUrl, localToken);
      setTestResult(success ? 'success' : 'error');
    } catch {
      setTestResult('error');
    } finally {
      setIsTesting(false);
    }
  };

  const handleSave = () => {
    onUpdate({
      enabled: localEnabled,
      url: localUrl,
      token: localToken || undefined,
    });
    setHasChanges(false);
    setLocalToken('');
  };

  const handleReset = () => {
    setLocalEnabled(enabled);
    setLocalUrl(url);
    setLocalToken('');
    setHasChanges(false);
    setTestResult(null);
  };

  return (
    <div className="space-y-4">
      {/* Enable/Disable Toggle */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">{title}</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            {localEnabled ? 'Enabled' : 'Disabled'}
          </p>
          {exclusiveHint && !localEnabled && (
            <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
              {exclusiveHint}
            </p>
          )}
        </div>
        <label className="relative inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            checked={localEnabled}
            onChange={(e) => handleEnabledChange(e.target.checked)}
            className="sr-only peer"
            disabled={isLoading}
          />
          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary-600"></div>
        </label>
      </div>

      {localEnabled && (
        <>
          {/* URL Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              API URL
            </label>
            <input
              type="url"
              value={localUrl}
              onChange={(e) => handleUrlChange(e.target.value)}
              placeholder="https://github.com/api/graphql"
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                placeholder-gray-400 dark:placeholder-gray-500
                focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              disabled={isLoading}
            />
          </div>

          {/* Token Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Access Token
            </label>
            <input
              type="password"
              value={localToken}
              onChange={(e) => handleTokenChange(e.target.value)}
              placeholder={tokenSet ? '••••••••••••••••' : 'Enter access token'}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                placeholder-gray-400 dark:placeholder-gray-500
                focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              disabled={isLoading}
            />
            {tokenSet && !localToken && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Token is already set. Enter a new token to update.
              </p>
            )}
          </div>

          {/* Test Connection */}
          <div className="flex items-center gap-3">
            <Button
              variant="secondary"
              onClick={handleTest}
              isLoading={isTesting}
              disabled={!localUrl || !localToken || isLoading}
            >
              Test Connection
            </Button>
            {testResult === 'success' && (
              <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
                <CheckCircleIcon className="w-5 h-5" />
                <span className="text-sm">Connection successful</span>
              </div>
            )}
            {testResult === 'error' && (
              <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
                <XCircleIcon className="w-5 h-5" />
                <span className="text-sm">Connection failed</span>
              </div>
            )}
          </div>
        </>
      )}

      {/* Action Buttons */}
      {hasChanges && (
        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
          <Button variant="secondary" onClick={handleReset} disabled={isLoading}>
            {t('common.cancel')}
          </Button>
          <Button variant="primary" onClick={handleSave} isLoading={isLoading}>
            {t('common.save')}
          </Button>
        </div>
      )}
    </div>
  );
};
