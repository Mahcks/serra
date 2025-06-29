import {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
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

// Helper function to check if we have a valid token
const hasValidToken = (): boolean => {
  // Since the cookie is HTTPOnly, we can't check it from JavaScript
  // We'll rely on the /me API call to determine authentication status
  console.log("🔍 HTTPOnly cookie - cannot check from JavaScript");
  return false;
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [shouldFetchUser, setShouldFetchUser] = useState(false);
  const [initialLoadComplete, setInitialLoadComplete] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const initialLoadDone = useRef(false);

  // Check setup status first
  const { data: setupStatus, isLoading: setupLoading } = useQuery({
    queryKey: ["setupStatus"],
    queryFn: backendApi.getSetupStatus,
    retry: false,
  });

  // Check for existing token on mount and enable user fetching if found
  useEffect(() => {
    if (setupStatus?.setup_complete) {
      console.log("🔍 Setup complete, enabling user fetch to check authentication");
      setShouldFetchUser(true);
      
      // Add a small delay to prevent flashing
      const timer = setTimeout(() => {
        setInitialLoadComplete(true);
      }, 100);
      
      return () => clearTimeout(timer);
    }
  }, [setupStatus?.setup_complete]);

  // Only run /me query if setup is complete and we have a token
  const { data: userData, isLoading, error } = useQuery<User, Error>({
    queryKey: ["me"],
    queryFn: async () => {
      console.log("🔐 Making /me API call to check authentication status");
      try {
        const data = await backendApi.getCurrentUser();
        if (!data || !data.id || !data.username || typeof data.is_admin !== 'boolean') {
          throw new Error("Invalid user data received");
        }
        console.log("✅ /me API call successful:", data.username);
        return {
          id: data.id,
          username: data.username,
          is_admin: data.is_admin
        };
      } catch (err) {
        console.log("❌ /me API call failed - user not authenticated:", err);
        throw err;
      }
    },
    retry: false, // Don't retry - if it fails, user is not authenticated
    enabled: shouldFetchUser, // Only enable when we explicitly want to fetch user data
    refetchInterval: false,
    refetchOnWindowFocus: false,
    staleTime: 0,
    gcTime: 0,
    // Don't trigger global error handling for this query
    meta: {
      skipGlobalErrorHandler: true,
    },
  });

  // Handle auth state updates
  useEffect(() => {
    const updateAuthState = async () => {
      // Only consider authenticated if we have user data, no errors, and not loading
      const isAuthed = !!userData && !isLoading && !error;
      const currentPath = location.pathname;
      const isPublicRoute = PUBLIC_ROUTES.includes(currentPath);

      console.log("🔄 Auth state update:", {
        hasUserData: !!userData,
        isLoading,
        hasError: !!error,
        isAuthed,
        shouldFetchUser,
        currentPath
      });

      // Wait for the initial load to complete
      if (!isLoading) {
        initialLoadDone.current = true;

        if (error || !isAuthed) {
          // Don't try to refresh token here - let the API interceptor handle it
          // Don't redirect here either - let the interceptor handle redirects
          console.log("❌ Auth state: not authenticated, but letting interceptor handle redirect");
          setIsAuthenticated(false);
          setUser(null);
          setInitialLoadComplete(true);
        } else {
          console.log("✅ Auth state: authenticated with user data");
          setIsAuthenticated(true);
          setUser(userData as User);
          setInitialLoadComplete(true);
          
          // If we're authenticated and on a public route, redirect to dashboard
          if (isPublicRoute) {
            navigate("/dashboard", { replace: true });
          }
        }
      }
    };

    updateAuthState();
  }, [userData, isLoading, error, navigate, location, shouldFetchUser]);

  const refreshToken = async () => {
    try {
      await backendApi.refreshToken();
      // Invalidate user data query to trigger a refetch
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    } catch (error) {
      console.error("Token refresh failed:", error);
      // If refresh fails, logout the user
      await logout();
      throw error;
    }
  };

  const login = async (username: string, password: string) => {
    try {
      console.log("🔐 Attempting login for user:", username);
      await backendApi.login(username, password);
      
      console.log("🔑 Login successful, enabling user fetch");
      // Enable user data fetching after successful login
      setShouldFetchUser(true);
    } catch (error) {
      console.error("❌ Login failed:", error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      console.log("🚪 Attempting logout");
      await backendApi.logout();
      
      console.log("🔑 Logout successful, disabling user fetch");
      
      // Clear all queries from the cache
      queryClient.clear();
      setIsAuthenticated(false);
      setUser(null);
      setShouldFetchUser(false);
      navigate("/login", { replace: true });
    } catch (error) {
      console.error("❌ Logout error:", error);
      // Force logout even if API call fails
      queryClient.clear();
      setIsAuthenticated(false);
      setUser(null);
      setShouldFetchUser(false);
      navigate("/login", { replace: true });
    }
  };

  // Show loading state during initial load or setup check
  if (setupLoading || (shouldFetchUser && isLoading && !initialLoadComplete)) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading: isLoading || setupLoading,
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
