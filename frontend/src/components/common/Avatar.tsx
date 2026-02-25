import React, { useState } from 'react';

interface AvatarProps {
  src?: string | null;
  name: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizeClasses: Record<string, { container: string; text: string }> = {
  sm: { container: 'w-7 h-7', text: 'text-xs' },
  md: { container: 'w-10 h-10', text: 'text-sm' },
  lg: { container: 'w-24 h-24', text: 'text-4xl' },
};

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) {
    return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
  }
  return (parts[0]?.charAt(0) || '?').toUpperCase();
}

export const Avatar: React.FC<AvatarProps> = ({ src, name, size = 'md', className = '' }) => {
  const [failed, setFailed] = useState(false);
  const s = sizeClasses[size];

  if (src && !failed) {
    return (
      <img
        src={src}
        alt={name}
        className={`${s.container} rounded-full object-cover flex-shrink-0 ${className}`}
        onError={() => setFailed(true)}
      />
    );
  }

  return (
    <div
      className={`${s.container} rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center flex-shrink-0 ${className}`}
    >
      <span className={`${s.text} text-primary-700 dark:text-primary-300 font-medium`}>
        {getInitials(name)}
      </span>
    </div>
  );
};
