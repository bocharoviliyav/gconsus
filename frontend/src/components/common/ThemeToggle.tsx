import React from "react";
import { useTheme } from "../../hooks/useTheme";
import {
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
} from "@heroicons/react/24/outline";

export const ThemeToggle: React.FC = () => {
  const { theme, setTheme } = useTheme();

  const themes: Array<{
    value: "light" | "dark" | "system";
    icon: React.ElementType;
    label: string;
  }> = [
    { value: "light", icon: SunIcon, label: "Light" },
    { value: "dark", icon: MoonIcon, label: "Dark" },
    { value: "system", icon: ComputerDesktopIcon, label: "System" },
  ];

  return (
    <div className="flex items-center gap-1 p-1 bg-gray-200 dark:bg-gray-700 rounded-lg">
      {themes.map(({ value, icon: Icon, label }) => (
        <button
          key={value}
          onClick={() => setTheme(value)}
          className={`
            p-2 rounded-md transition-colors
            ${
              theme === value
                ? "bg-white dark:bg-gray-600 text-primary-600 dark:text-primary-400 shadow-sm"
                : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
            }
          `}
          title={label}
          aria-label={`Switch to ${label.toLowerCase()} theme`}
        >
          <Icon className="w-5 h-5" />
        </button>
      ))}
    </div>
  );
};
