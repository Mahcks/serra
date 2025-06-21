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
import { Dashboard } from "@/pages/Dashboard";
import { RequestPage } from "@/pages/RequestPage";
import { backendApi } from "@/lib/api";
import { AuthProvider, useAuth } from "@/lib/auth";
import { Header } from "@/components/Header";

// Protected Route wrapper component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // Check setup status
  const { data: setupStatus, isLoading: setupLoading } = useQuery({
    queryKey: ["setupStatus"],
    queryFn: backendApi.getSetupStatus,
    retry: false,
  });

  useEffect(() => {
    if (!setupLoading) {
      if (!setupStatus?.setup_complete) {
        navigate("/setup", { replace: true });
      } else if (!isAuthenticated) {
        navigate("/login", {
          replace: true,
          state: { from: location.pathname },
        });
      }
    }
  }, [setupStatus, setupLoading, isAuthenticated, navigate, location]);

  if (isLoading || setupLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  return isAuthenticated ? <>{children}</> : null;
}

// Dashboard layout with header and main content area
function DashboardLayout() {
  const { logout } = useAuth();
  const location = useLocation();

  const isFullWidthPage = location.pathname.startsWith("/request");

  return (
    <div className="min-h-screen bg-gray-50">
      <Header onLogout={logout} currentPath={location.pathname} />

      {isFullWidthPage ? (
        <main className="w-full h-full">
          <Outlet />
        </main>
      ) : (
        <main className="max-w-8xl mx-auto px-3 sm:px-6 py-4 sm:py-8">
          <Outlet />
        </main>
      )}
    </div>
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
    if (!setupLoading) {
      if (!setupStatus?.setup_complete && location.pathname !== "/setup") {
        navigate("/setup", { replace: true });
      }
    }
  }, [setupStatus, setupLoading, location, navigate]);

  if (setupLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p>Loading...</p>
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
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </SetupGuard>
  );
}

export default function App() {
  return (
    <Router>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </Router>
  );
}
