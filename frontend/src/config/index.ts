export const config = {
  api: {
    baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
    timeout: 30000,
  },
  keycloak: {
    url: import.meta.env.VITE_KEYCLOAK_URL || "http://localhost:8090",
    realm: import.meta.env.VITE_KEYCLOAK_REALM || "gconsus",
    clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || "gconsus-frontend",
  },
  app: {
    name: import.meta.env.VITE_APP_NAME || "GConsus",
    version: import.meta.env.VITE_APP_VERSION || "1.0.0",
    defaultLocale: import.meta.env.VITE_DEFAULT_LOCALE || "en",
  },
} as const;

export const ROUTES = {
  home: "/",
  dashboard: "/dashboard",
  teams: "/teams",
  teamDetail: (id: string) => `/teams/${id}`,
  analytics: "/analytics",
  userProfile: (id: string) => `/users/${id}`,
  repositories: "/repositories",
  settings: "/settings",
  login: "/login",
} as const;

export const ROLES = {
  admin: "admin",
  manager: "manager",
  user: "user",
} as const;

export const STORAGE_KEYS = {
  theme: "git-info-theme",
  locale: "git-info-locale",
  sidebarCollapsed: "git-info-sidebar-collapsed",
} as const;
