import { useEffect, useState } from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useNavigate,
  useLocation,
  Outlet,
} from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { SetupStepper } from "@/components/auth/SetupStepper";
import { LoginForm } from "@/components/auth/LoginForm";
import { Dashboard } from "@/pages/DashboardPage";
import { RequestPage } from "@/pages/user/RequestPage";
import { SearchPage } from "@/pages/search/SearchPage";
import { backendApi } from "@/lib/api";
import { AuthProvider, useAuth } from "@/lib/auth";
import { SettingsProvider } from "@/lib/settings";
import { ThemeProvider } from "@/lib/theme";
import { AppSidebar } from "@/components/layout/AppSidebar";
import { WebSocketProvider } from "@/lib/WebSocketContext";
import { WebSocketStatus } from "@/components/layout/WebSocketStatus";
import { HeaderSearchBar } from "@/components/layout/HeaderSearchBar";
import { CommandPalette } from "@/components/search/CommandPalette";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import UserSettingsPage from "@/pages/UserSettingsPage";
import AdminUserSettingsPage from "@/pages/admin/AdminUserSettingsPage";
import RequestsPage from "@/pages/admin/RequestsPage";
import { useTokenRefresh } from "@/hooks/useTokenRefresh";
import { Toaster } from "@/components/ui/sonner";
import { useGlobalSearch } from "@/hooks/useGlobalSearch";
import MediaDetailsPage from "@/pages/media/MediaDetailsPage";
import ReleaseDatesPage from "@/pages/media/ReleaseDatesPage";
import CollectionPage from "@/pages/media/CollectionPage";
import PersonPage from "@/pages/media/PersonPage";
import Settings from "@/pages/admin/Settings";
import AnalyticsPage from "@/pages/admin/AnalyticsPage";
import InviteAcceptPage from "@/pages/InviteAcceptPage";
import UsersAndInvitationsPage from "./pages/admin/UsersAndInvitationsPage";

// Protected Route wrapper component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-muted border-t-primary mx-auto"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Don't render children if user is not authenticated
  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}

// Dashboard layout with sidebar and main content area
function DashboardLayout() {
  const { logout } = useAuth();
  const location = useLocation();
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  // Enable automatic token refresh for authenticated users
  useTokenRefresh(); // Check every 15 minutes (default)

  // Enable global search functionality
  const { searchBuffer, isSearching } = useGlobalSearch({
    onCommandPaletteOpen: () => setCommandPaletteOpen(true),
    enableInstantSearch: true,
    excludePaths: ["/login", "/setup"], // Removed /search since we go to /requests now
  });

  const isFullWidthPage =
    location.pathname.startsWith("/request") ||
    location.pathname.startsWith("/collection") ||
    location.pathname.startsWith("/person");

  // Determine if we should show search in header (not on admin pages)
  const showHeaderSearch =
    !location.pathname.startsWith("/admin") &&
    !location.pathname.startsWith("/setup") &&
    !location.pathname.startsWith("/login");

  return (
    <>
      <SidebarProvider>
        <AppSidebar onLogout={logout} />
        <SidebarInset className="flex flex-col">
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <div className="flex-1 flex items-center gap-4">
              <h1 className="text-lg font-semibold">
                {location.pathname === "/dashboard" || location.pathname === "/"
                  ? "Dashboard"
                  : location.pathname === "/requests"
                  ? "Requests"
                  : location.pathname === "/settings"
                  ? "Settings"
                  : location.pathname.startsWith("/admin/users")
                  ? "Users"
                  : location.pathname.startsWith("/admin/requests")
                  ? "Admin Requests"
                  : location.pathname.startsWith("/admin/analytics")
                  ? "Analytics"
                  : location.pathname.startsWith("/admin/invitations")
                  ? "Invitations"
                  : location.pathname.startsWith("/collection")
                  ? "Collection"
                  : location.pathname.startsWith("/person")
                  ? "Person"
                  : "Serra"}
              </h1>

              {/* Show search buffer indicator */}
              {isSearching && (
                <div className="text-sm text-muted-foreground bg-muted px-2 py-1 rounded">
                  Searching: "{searchBuffer}"
                </div>
              )}
            </div>

            {/* Header Search Bar - only on non-admin pages */}
            {showHeaderSearch && (
              <div className="flex items-center gap-3">
                <HeaderSearchBar
                  onCommandPaletteOpen={() => setCommandPaletteOpen(true)}
                  className="w-64 hidden md:block"
                  placeholder="Search or press Ctrl+K"
                />
              </div>
            )}

            <WebSocketStatus showDetails />
          </header>

          {isFullWidthPage ? (
            <main className="flex-1 overflow-auto">
              <Outlet />
            </main>
          ) : (
            <main className="flex-1 overflow-auto p-4 sm:p-6 lg:p-8">
              <Outlet />
            </main>
          )}
        </SidebarInset>
      </SidebarProvider>

      {/* Command Palette */}
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
      />
    </>
  );
}

// SetupGuard: Redirects to /setup if setup is not complete
function SetupGuard({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { data: setupStatus, isLoading: setupLoading } = useQuery({
    queryKey: ["setupStatus"],
    queryFn: backendApi.getSetupStatus,
    retry: false,
  });

  useEffect(() => {
    if (!setupLoading && setupStatus) {
      // Allow invitation acceptance routes even if setup isn't complete
      const isInviteRoute = location.pathname.startsWith("/invite/accept/");
      if (!setupStatus.setup_complete && location.pathname !== "/setup" && !isInviteRoute) {
        navigate("/setup", { replace: true });
      }
    }
  }, [setupStatus, setupLoading, location, navigate]);

  if (setupLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-muted border-t-primary mx-auto"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}

function AppRoutes() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const handleSetupComplete = () => {
    navigate("/login", { replace: true });
  };
  return (
    <SetupGuard>
      <Routes>
        <Route path="/login" element={<LoginForm onLoginSuccess={login} />} />
        <Route
          path="/setup"
          element={<SetupStepper onSetupComplete={handleSetupComplete} />}
        />
        {/* Public invitation acceptance route */}
        <Route path="/invite/accept/:token" element={<InviteAcceptPage />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="requests" element={<RequestPage />} />
          <Route
            path="requests/:media_type/:tmdb_id/details"
            element={<MediaDetailsPage />}
          />
          <Route
            path="requests/:media_type/:tmdb_id/release-dates"
            element={<ReleaseDatesPage />}
          />
          <Route
            path="collection/:collection_id"
            element={<CollectionPage />}
          />
          <Route path="person/:person_id" element={<PersonPage />} />
          <Route path="search" element={<SearchPage />} />
          <Route path="settings" element={<UserSettingsPage />} />
          <Route path="admin/users" element={<UsersAndInvitationsPage />} />
          <Route
            path="admin/users/:userId/settings"
            element={<AdminUserSettingsPage />}
          />
          <Route path="admin/requests" element={<RequestsPage />} />
          <Route path="admin/settings" element={<Settings />} />
          <Route path="admin/settings/:tab" element={<Settings />} />
          <Route path="admin/settings/email" element={<Settings />} />
          <Route path="admin/analytics" element={<AnalyticsPage />} />
          <Route path="admin/analytics/storage" element={<AnalyticsPage />} />
          <Route path="admin/analytics/requests" element={<AnalyticsPage />} />
          <Route path="admin/analytics/watch" element={<AnalyticsPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </SetupGuard>
  );
}

export default function App() {
  return (
    <Router>
      <ThemeProvider defaultTheme="dark" storageKey="serra-ui-theme">
        <AuthProvider>
          <SettingsProvider>
            <WebSocketProvider
              autoReconnect={true}
              reconnectInterval={2000}
              maxReconnectAttempts={15}
              heartbeatInterval={30000}
              messageQueueSize={50}
            >
              <AppRoutes />
              <Toaster />
            </WebSocketProvider>
          </SettingsProvider>
        </AuthProvider>
      </ThemeProvider>
    </Router>
  );
}
