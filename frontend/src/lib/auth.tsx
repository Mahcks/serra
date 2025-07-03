import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { backendApi } from "./api";

interface User {
  id: string;
  username: string;
  is_admin: boolean;
}

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: User | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

const PUBLIC_ROUTES = ['/login', '/setup'];

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [initialCheckDone, setInitialCheckDone] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();

  // Check setup status first
  const { data: setupStatus, isLoading: setupLoading } = useQuery({
    queryKey: ["setupStatus"],
    queryFn: backendApi.getSetupStatus,
    retry: false,
  });

  // Check authentication status by calling /me endpoint
  const { data: userData, isLoading: userLoading, error: userError } = useQuery<User, Error>({
    queryKey: ["me"],
    queryFn: async () => {
      console.log("🔐 Making /me API call to check authentication status");
      const data = await backendApi.getCurrentUser();
      console.log("✅ /me API call successful:", data.username);
      return data;
    },
    retry: false,
    enabled: setupStatus?.setup_complete && !initialCheckDone, // Only run once when setup is complete
    refetchInterval: false,
    refetchOnWindowFocus: false,
    staleTime: Infinity,
  });

  // Handle authentication state
  useEffect(() => {
    if (setupLoading) return;

    // If setup is not complete, clear auth state
    if (!setupStatus?.setup_complete) {
      setIsAuthenticated(false);
      setUser(null);
      setInitialCheckDone(true);
      return;
    }

    // If we're still loading user data, wait
    if (userLoading) return;

    // Mark initial check as done
    setInitialCheckDone(true);

    if (userData && !userError) {
      // User is authenticated
      console.log("✅ User authenticated:", userData.username);
      setIsAuthenticated(true);
      setUser(userData);

      // If on login page and authenticated, redirect to dashboard
      if (location.pathname === '/login') {
        navigate('/dashboard', { replace: true });
      }
    } else {
      // User is not authenticated
      console.log("❌ User not authenticated");
      setIsAuthenticated(false);
      setUser(null);

      // If on protected route and not authenticated, redirect to login
      const isPublicRoute = PUBLIC_ROUTES.includes(location.pathname);
      if (!isPublicRoute) {
        navigate('/login', { replace: true });
      }
    }
  }, [setupStatus, setupLoading, userData, userLoading, userError, location.pathname, navigate, initialCheckDone]);

  const refreshToken = async () => {
    try {
      await backendApi.refreshToken();
      // Invalidate and refetch user data
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      const newUserData = await queryClient.fetchQuery({
        queryKey: ["me"],
        queryFn: backendApi.getCurrentUser,
      });
      if (newUserData) {
        setIsAuthenticated(true);
        setUser(newUserData);
      }
    } catch (error) {
      console.error("Token refresh failed:", error);
      await logout();
      throw error;
    }
  };

  const login = async (username: string, password: string) => {
    try {
      console.log("🔐 Attempting login for user:", username);
      await backendApi.login(username, password);
      
      console.log("🔑 Login successful, fetching user data");
      // Fetch user data immediately after login
      const newUserData = await queryClient.fetchQuery({
        queryKey: ["me"],
        queryFn: backendApi.getCurrentUser,
      });
      
      if (newUserData) {
        setIsAuthenticated(true);
        setUser(newUserData);
        setInitialCheckDone(true);
        console.log("✅ Login complete, user authenticated:", newUserData.username);
      }
    } catch (error) {
      console.error("❌ Login failed:", error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      console.log("🚪 Attempting logout");
      await backendApi.logout();
    } catch (error) {
      console.error("❌ Logout error:", error);
    } finally {
      // Always clear state regardless of API call success
      console.log("🔑 Clearing auth state");
      queryClient.clear();
      setIsAuthenticated(false);
      setUser(null);
      setInitialCheckDone(true);
      navigate("/login", { replace: true });
    }
  };

  // Show loading state during initial checks
  if (setupLoading || (!initialCheckDone && userLoading)) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-muted border-t-primary mx-auto"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading: setupLoading || userLoading,
        user,
        login,
        logout,
        refreshToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
