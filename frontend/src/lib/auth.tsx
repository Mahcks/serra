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
}

const AuthContext = createContext<AuthContextType | null>(null);

const PUBLIC_ROUTES = ['/login', '/setup'];

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
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

  // Only run /me query if setup is complete
  const { data: userData, isLoading, error } = useQuery<User, Error>({
    queryKey: ["me"],
    queryFn: async () => {
      try {
        const data = await backendApi.getCurrentUser();
        if (!data || !data.id || !data.username || typeof data.is_admin !== 'boolean') {
          throw new Error("Invalid user data received");
        }
        return {
          id: data.id,
          username: data.username,
          is_admin: data.is_admin
        };
      } catch (err) {
        console.error("Failed to fetch user data:", err);
        throw err;
      }
    },
    retry: 1,
    enabled: !!setupStatus?.setup_complete,
    refetchInterval: 5 * 60 * 1000,
    refetchOnWindowFocus: !isAuthenticated,
    staleTime: 0,
    gcTime: 0
  });

  // Handle auth state updates
  useEffect(() => {
    const updateAuthState = () => {
      const isAuthed = !!userData;
      const currentPath = location.pathname;
      const isPublicRoute = PUBLIC_ROUTES.includes(currentPath);

      // Wait for the initial load to complete
      if (!isLoading) {
        initialLoadDone.current = true;

        if (error || !isAuthed) {
          // Only redirect to login if we're not on a public route
          if (!isPublicRoute) {
            console.log("No auth, redirecting to login");
            setIsAuthenticated(false);
            setUser(null);
            navigate("/login", { replace: true });
          }
        } else {
          setIsAuthenticated(true);
          setUser(userData);
          
          // If we're authenticated and on a public route, redirect to dashboard
          if (isPublicRoute) {
            navigate("/dashboard", { replace: true });
          }
        }
      }
    };

    updateAuthState();
  }, [userData, isLoading, error, navigate, location]);

  const login = async (username: string, password: string) => {
    try {
      await backendApi.login(username, password);
      // Invalidate user data query to trigger a refetch
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    } catch (error) {
      console.error("Login failed:", error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await backendApi.logout();
      // Clear all queries from the cache
      queryClient.clear();
      setIsAuthenticated(false);
      setUser(null);
      navigate("/login", { replace: true });
    } catch (error) {
      console.error("Logout error:", error);
      // Force logout even if API call fails
      queryClient.clear();
      setIsAuthenticated(false);
      setUser(null);
      navigate("/login", { replace: true });
    }
  };

  // Show loading state during initial load or setup check
  if ((isLoading && !initialLoadDone.current) || setupLoading) {
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
        isLoading,
        user,
        login,
        logout,
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
