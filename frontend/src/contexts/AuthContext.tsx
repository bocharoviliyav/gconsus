import React, {
  createContext,
  useState,
  useEffect,
  type ReactNode,
} from "react";
import { initKeycloak, login, logout, getUserInfo } from "../services/keycloak";
import type { UserClaims, AuthState } from "../types";

export interface AuthContextType extends AuthState {
  login: () => Promise<void>;
  logout: () => void;
  hasRole: (role: string) => boolean;
  hasAnyRole: (roles: string[]) => boolean;
  isLoading: boolean;
}

export const AuthContext = createContext<AuthContextType | undefined>(
  undefined,
);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    user: null,
    token: null,
    roles: [],
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    initializeAuth();
  }, []);

  const initializeAuth = async () => {
    try {
      initKeycloak();
      const authenticated = await login();

      if (authenticated) {
        const userInfo = getUserInfo();
        if (userInfo) {
          setAuthState({
            isAuthenticated: true,
            user: userInfo as UserClaims,
            token: null, // Token managed by Keycloak
            roles: userInfo.roles,
          });
        }
      }
    } catch (error) {
      console.error("Failed to initialize authentication", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogin = async () => {
    setIsLoading(true);
    try {
      const authenticated = await login();
      if (authenticated) {
        const userInfo = getUserInfo();
        if (userInfo) {
          setAuthState({
            isAuthenticated: true,
            user: userInfo as UserClaims,
            token: null,
            roles: userInfo.roles,
          });
        }
      }
    } catch (error) {
      console.error("Login failed", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    setAuthState({
      isAuthenticated: false,
      user: null,
      token: null,
      roles: [],
    });
  };

  const hasRole = (role: string): boolean => {
    return authState.roles.includes(role);
  };

  const hasAnyRole = (roles: string[]): boolean => {
    return roles.some((role) => authState.roles.includes(role));
  };

  const value: AuthContextType = {
    ...authState,
    login: handleLogin,
    logout: handleLogout,
    hasRole,
    hasAnyRole,
    isLoading,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
