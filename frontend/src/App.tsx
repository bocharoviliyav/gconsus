import React from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "./contexts/AuthContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import { Layout } from "./components/layout/Layout";
import { ProtectedRoute } from "./components/common/ProtectedRoute";
import { Loading } from "./components/common/Loading";
import { ROUTES, ROLES } from "./config";
import "./config/i18n";

// Lazy load pages
const Dashboard = React.lazy(() => import("./pages/Dashboard/Dashboard"));
const Teams = React.lazy(() => import("./pages/Teams/Teams"));
const TeamAnalytics = React.lazy(
  () => import("./pages/TeamAnalytics/TeamAnalytics"),
);
const UserProfile = React.lazy(() => import("./pages/UserProfile/UserProfile"));
const RepositoriesAnalytics = React.lazy(
  () => import("./pages/RepositoriesAnalytics/RepositoriesAnalytics"),
);
const Analytics = React.lazy(() => import("./pages/Analytics/Analytics"));
const Settings = React.lazy(() => import("./pages/Settings/Settings"));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthProvider>
          <BrowserRouter>
            <React.Suspense fallback={<Loading fullScreen />}>
              <Routes>
                {/* Redirect root to dashboard */}
                <Route
                  path="/"
                  element={<Navigate to={ROUTES.dashboard} replace />}
                />

                {/* Protected routes with layout */}
                <Route element={<Layout />}>
                  <Route
                    path={ROUTES.dashboard}
                    element={
                      <ProtectedRoute>
                        <Dashboard />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path={ROUTES.teams}
                    element={
                      <ProtectedRoute>
                        <Teams />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path="/teams/:teamId"
                    element={
                      <ProtectedRoute>
                        <TeamAnalytics />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path="/users/:userId"
                    element={
                      <ProtectedRoute>
                        <UserProfile />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path={ROUTES.repositories}
                    element={
                      <ProtectedRoute>
                        <RepositoriesAnalytics />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path={ROUTES.analytics}
                    element={
                      <ProtectedRoute>
                        <Analytics />
                      </ProtectedRoute>
                    }
                  />
                  <Route
                    path={ROUTES.settings}
                    element={
                      <ProtectedRoute roles={[ROLES.admin]}>
                        <Settings />
                      </ProtectedRoute>
                    }
                  />
                </Route>

                {/* Catch all - redirect to dashboard */}
                <Route
                  path="*"
                  element={<Navigate to={ROUTES.dashboard} replace />}
                />
              </Routes>
            </React.Suspense>
          </BrowserRouter>
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
