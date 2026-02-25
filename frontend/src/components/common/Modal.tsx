import React from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  size?: "sm" | "md" | "lg" | "xl";
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  children,
  size = "md",
}) => {
  if (!isOpen) return null;

  const sizeClasses = {
    sm: "max-w-md",
    md: "max-w-lg",
    lg: "max-w-2xl",
    xl: "max-w-4xl",
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div
        className="flex min-h-screen items-center justify-center p-4 text-center sm:p-0"
        onClick={handleBackdropClick}
      >
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />

        <div
          className={`relative w-full ${sizeClasses[size]} transform rounded-lg bg-white shadow-xl transition-all sm:my-8`}
        >
          <div className="bg-white dark:bg-gray-800 rounded-lg px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            {title && (
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg dark:text-white font-medium text-gray-900">
                  {title}
                </h3>
                <button
                  type="button"
                  className="rounded-md bg-white  dark:bg-gray-600  dark:text-white text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  onClick={onClose}
                >
                  <span className="sr-only">Close</span>
                  <XMarkIcon className="h-6 w-6" />
                </button>
              </div>
            )}
            <div>{children}</div>
          </div>
        </div>
      </div>
    </div>
  );
};
