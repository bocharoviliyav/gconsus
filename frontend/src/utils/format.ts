/**
 * Number and text formatting utilities
 */

/**
 * Format number with thousand separators
 */
export const formatNumber = (num: number, locale: string = "en"): string => {
  return new Intl.NumberFormat(locale).format(num);
};

/**
 * Format number as percentage
 */
export const formatPercent = (
  num: number,
  decimals: number = 1,
  locale: string = "en",
): string => {
  return new Intl.NumberFormat(locale, {
    style: "percent",
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num);
};

/**
 * Format large numbers with abbreviations (K, M, B)
 */
export const formatCompactNumber = (
  num: number,
  locale: string = "en",
): string => {
  if (num < 1000) return num.toString();
  const formatter = new Intl.NumberFormat(locale, {
    notation: "compact",
    compactDisplay: "short",
  });
  return formatter.format(num);
};

/**
 * Format bytes to human readable format
 */
export const formatBytes = (bytes: number, decimals: number = 2): string => {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
};

/**
 * Truncate text with ellipsis
 */
export const truncate = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + "...";
};

/**
 * Capitalize first letter
 */
export const capitalize = (text: string): string => {
  if (!text) return "";
  return text.charAt(0).toUpperCase() + text.slice(1);
};

/**
 * Convert camelCase to Title Case
 */
export const camelToTitle = (text: string): string => {
  return text
    .replace(/([A-Z])/g, " $1")
    .replace(/^./, (str) => str.toUpperCase())
    .trim();
};

/**
 * Format duration in seconds to human readable
 */
export const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  }
  return `${secs}s`;
};

/**
 * Get initials from name (e.g., "John Doe" -> "JD")
 */
export const getInitials = (name: string): string => {
  if (!name) return "";

  const parts = name.trim().split(" ");
  if (parts.length === 1) {
    return parts[0].charAt(0).toUpperCase();
  }

  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
};

/**
 * Format lines of code with +/- indicators
 */
export const formatLinesChanged = (added: number, deleted: number): string => {
  const addedStr = added > 0 ? `+${formatCompactNumber(added)}` : "";
  const deletedStr = deleted > 0 ? `-${formatCompactNumber(deleted)}` : "";

  if (addedStr && deletedStr) {
    return `${addedStr} ${deletedStr}`;
  }
  return addedStr || deletedStr || "0";
};
