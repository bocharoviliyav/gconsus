/**
 * Validation utility functions
 */

/**
 * Validate email format
 */
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * Validate URL format
 */
export const isValidUrl = (url: string): boolean => {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
};

/**
 * Validate GitHub username
 */
export const isValidGitHubUsername = (username: string): boolean => {
  // GitHub username rules: 1-39 chars, alphanumeric + hyphens, no consecutive hyphens
  const regex = /^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$/;
  return regex.test(username);
};

/**
 * Validate GitLab username
 */
export const isValidGitLabUsername = (username: string): boolean => {
  // GitLab username rules: similar to GitHub
  const regex = /^[a-zA-Z0-9_.-]+$/;
  return regex.test(username) && username.length >= 2 && username.length <= 255;
};

/**
 * Validate username format
 */
export const isValidUsername = (username: string): boolean => {
  return /^[a-zA-Z0-9_.-]{2,100}$/.test(username);
};

/**
 * Check if string is empty or whitespace
 */
export const isEmpty = (str: string): boolean => {
  return !str || str.trim().length === 0;
};

/**
 * Validate minimum length
 */
export const hasMinLength = (str: string, minLength: number): boolean => {
  return str.length >= minLength;
};

/**
 * Validate maximum length
 */
export const hasMaxLength = (str: string, maxLength: number): boolean => {
  return str.length <= maxLength;
};

/**
 * Validate date range
 */
export const isValidDateRange = (start: Date | string, end: Date | string): boolean => {
  const startDate = typeof start === 'string' ? new Date(start) : start;
  const endDate = typeof end === 'string' ? new Date(end) : end;
  return startDate <= endDate;
};

/**
 * Validate required field
 */
export const isRequired = (value: any): boolean => {
  if (value === null || value === undefined) return false;
  if (typeof value === 'string') return !isEmpty(value);
  if (Array.isArray(value)) return value.length > 0;
  return true;
};

/**
 * Form validation helper
 */
export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  email?: boolean;
  url?: boolean;
  pattern?: RegExp;
  custom?: (value: any) => boolean;
  message?: string;
}

export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

export const validate = (value: any, rules: ValidationRule): ValidationResult => {
  if (rules.required && !isRequired(value)) {
    return { isValid: false, error: rules.message || 'This field is required' };
  }

  if (typeof value === 'string') {
    if (rules.minLength && !hasMinLength(value, rules.minLength)) {
      return {
        isValid: false,
        error: rules.message || `Minimum length is ${rules.minLength} characters`,
      };
    }

    if (rules.maxLength && !hasMaxLength(value, rules.maxLength)) {
      return {
        isValid: false,
        error: rules.message || `Maximum length is ${rules.maxLength} characters`,
      };
    }

    if (rules.email && !isValidEmail(value)) {
      return { isValid: false, error: rules.message || 'Invalid email format' };
    }

    if (rules.url && !isValidUrl(value)) {
      return { isValid: false, error: rules.message || 'Invalid URL format' };
    }

    if (rules.pattern && !rules.pattern.test(value)) {
      return { isValid: false, error: rules.message || 'Invalid format' };
    }
  }

  if (rules.custom && !rules.custom(value)) {
    return { isValid: false, error: rules.message || 'Validation failed' };
  }

  return { isValid: true };
};
