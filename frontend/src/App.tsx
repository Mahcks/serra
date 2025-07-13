import { useEffect } from "react";
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
import { SetupStepper } from "@/components/SetupStepper";
import { LoginForm } from "@/components/LoginForm";
import { Dashboard } from "@/pages/DashboardPage";
import { RequestPage } from "@/pages/RequestPage";
import { MediaDetailsPage } from "@/pages/MediaDetailsPage";
import { SearchPage } from "@/pages/SearchPage";
import { backendApi } from "@/lib/api";
import { AuthProvider, useAuth } from "@/lib/auth";
import { SettingsProvider } from "@/lib/settings";
import { ThemeProvider } from "@/lib/theme";
import { AppSidebar } from "@/components/AppSidebar";
import { WebSocketProvider } from "@/lib/WebSocketContext";
import { WebSocketStatus } from "@/components/WebSocketStatus";
import { FloatingSearchBar } from "@/components/FloatingSearchBar";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import UsersPage from "@/pages/users/UsersPage";
import UserSettingsPage from "@/pages/users/UserSettingsPage";
import RequestsPage from "@/pages/admin/RequestsPage";
import { useTokenRefresh } from "@/hooks/useTokenRefresh";
import { Toaster } from "@/components/ui/sonner";

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
  
  // Enable automatic token refresh for authenticated users
  useTokenRefresh(); // Check every 15 minutes (default)

  const isFullWidthPage = location.pathname.startsWith("/request");

  return (
    <SidebarProvider>
      <AppSidebar onLogout={logout} />
      <SidebarInset className="flex flex-col">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <div className="flex-1">
            <h1 className="text-lg font-semibold">
              {location.pathname === "/dashboard" || location.pathname === "/" 
                ? "Dashboard" 
                : location.pathname === "/requests"
                ? "Requests"
                : "Serra"}
            </h1>
          </div>
          <WebSocketStatus showDetails />
        </header>
        
        {/* Floating Search Bar */}
        <FloatingSearchBar />
        
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
      if (!setupStatus.setup_complete && location.pathname !== "/setup") {
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
          <Route path="requests/:tmdb_id/details" element={<MediaDetailsPage />} />
          <Route path="search" element={<SearchPage />} />
          <Route path="admin/users" element={<UsersPage />} />
          <Route path="admin/users/:userId/settings" element={<UserSettingsPage />} />
          <Route path="admin/requests" element={<RequestsPage />} />
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
