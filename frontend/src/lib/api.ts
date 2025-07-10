import axios from "axios";
import { QueryClient } from "@tanstack/react-query";
import type { Provider, UserWithPermissions } from "@/types";

// Create an axios instance with default config
export const api = axios.create({
  baseURL: "http://localhost:9090/v1",
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // This is important for CORS with credentials
});

export const publicApi = axios.create({
  withCredentials: false,
})

// Track if a refresh is in progress to prevent multiple concurrent attempts
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: any) => void;
  reject: (reason?: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
    } else {
      resolve(token);
    }
  });

  failedQueue = [];
};

// Proactive token refresh - call this periodically or before making important requests
export const proactiveTokenRefresh = async () => {
  if (isRefreshing) {
    return; // Already refreshing
  }

  try {
    console.log("🔄 Checking token status...");
    await backendApi.refreshToken();
    console.log("✅ Token refresh check completed");
  } catch (error: any) {
    // Only log error if it's not a "token still valid" response
    if (error.response?.status !== 200) {
      console.log("⚠️ Token refresh failed:", error.response?.data?.message || error.message);
    }
  }
};

// Add request interceptor to proactively check token before requests
api.interceptors.request.use(
  async (config) => {
    // Skip token check for auth endpoints to prevent infinite loops
    if (config.url?.includes('/auth/') || config.url?.includes('/setup')) {
      return config;
    }

    // Proactively check token status for important requests
    try {
      await proactiveTokenRefresh();
    } catch (error) {
      // Continue with request even if refresh fails - the response interceptor will handle 401s
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// Add response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // For /me endpoint 401s, just let them fail normally (no refresh, no redirect)
    if (error.response?.status === 401 && originalRequest.url?.includes('/me')) {
      console.log("🔐 /me endpoint returned 401 - user not authenticated");
      return Promise.reject(error);
    }

    // If it's a refresh token call that failed, just reject
    if (error.response?.status === 401 && originalRequest.url?.includes('/auth/refresh')) {
      console.log("🔐 Refresh token failed");
      return Promise.reject(error);
    }

    // If the error is 401 and we haven't already tried to refresh
    // AND it's not the refresh token endpoint itself (to prevent infinite loops)
    // AND it's not the /me endpoint (used for auth checking)
    if (error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/auth/refresh') &&
      !originalRequest.url?.includes('/me')) {
      if (isRefreshing) {
        // If a refresh is already in progress, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then(() => {
          return api(originalRequest);
        }).catch((err) => {
          return Promise.reject(err);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        console.log("🔄 Attempting to refresh token...");
        // Try to refresh the token
        const refreshResponse = await backendApi.refreshToken();
        console.log("✅ Token refreshed successfully");

        // Process any queued requests
        processQueue(null);

        // Retry the original request
        return api(originalRequest);
      } catch (refreshError: any) {
        // If refresh fails, process queue with error
        processQueue(refreshError);
        
        // Check if it's actually a successful "no refresh needed" response
        if (refreshError.response?.status === 200) {
          console.log("ℹ️ Token still valid, retrying original request");
          processQueue(null);
          return api(originalRequest);
        }
        
        console.error("❌ Token refresh failed:", refreshError.response?.data?.message || refreshError.message);
        
        // For authentication failures, you might want to redirect to login
        if (refreshError.response?.status === 401) {
          console.log("🔐 Authentication failed - user needs to login again");
          // You could dispatch a logout action here or redirect to login
          // window.location.href = '/login';
        }

        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// Create a query client
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 0,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000,
    },
  },
});

// Media server API functions
export const mediaServerApi = {
  testConnection: async (url: string) => {
    const normalizedUrl = url.replace(/\/+$/, "");
    const response = await publicApi.get(`${normalizedUrl}/System/Info/Public`);
    return response.data;
  },

  login: async (url: string, username: string, password: string) => {
    const normalizedUrl = url.replace(/\/+$/, "");
    const response = await api.post(
      `${normalizedUrl}/Users/AuthenticateByName`,
      {
        Username: username,
        Pw: password,
      }
    );
    return response.data;
  },
};

// Backend API functions
export const backendApi = {
  // Authenticate with the media server
  login: async (username: string, password: string) => {
    const response = await api.post("/auth/login", {
      username,
      password,
    });

    // Handle 204 No Content response
    if (response.status === 204) {
      return { success: true };
    }

    return response.data;
  },

  // Refresh the JWT token
  refreshToken: async () => {
    const response = await api.post("/auth/refresh");
    return response.data;
  },

  // Get current user info - used for auth status
  getCurrentUser: async () => {
    const response = await api.get("/me");
    return response.data;
  },

  // Logout from the media server
  logout: async () => {
    const response = await api.post("/auth/logout");
    return response.data;
  },

  // setup sends the setup data to the backend
  setup: async (
    type: Provider,
    url: string,
    apiKey: string,
    requestSystem: string,
    requestSystemUrl?: string,
    radarr?: any[],
    sonarr?: any[]
  ) => {
    const payload: any = {
      type,
      url,
      api_key: apiKey,
      request_system: requestSystem,
    };

    if (requestSystemUrl) {
      payload.request_system_url = requestSystemUrl;
    }
    if (radarr) {
      payload.radarr = radarr;
    }
    if (sonarr) {
      payload.sonarr = sonarr;
    }

    const response = await api.post("/setup", payload);
    return response.data;
  },

  // getSetupStatus checks if the setup process is complete
  getSetupStatus: async () => {
    const response = await api.get("/setup/status");
    return response.data;
  },

  getSettings: async () => {
    const response = await api.get("/settings");
    return response.data;
  },

  getDownloads: async () => {
    const response = await api.get("/downloads");
    return response.data;
  },

  getUsers: async () => {
    const response = await api.get("/users");
    return response.data;
  },

  getUser: async (userId: string): Promise<UserWithPermissions> => {
    const response = await api.get(`/users/${userId}`);
    return response.data;
  },

  getPermissions: async () => {
    const response = await api.get("/permissions");
    return response.data;
  },

  updateUserPermissions: async (userId: string, permissions: string[]) => {
    const response = await api.put(`/users/${userId}/permissions`, {
      permissions: permissions,
    });
    return response.data;
  },

  createLocalUser: async (userData: { username: string; email?: string; password: string }) => {
    const response = await api.post("/users/local", userData);
    return response.data;
  },

  changeUserPassword: async (userId: string, newPassword: string) => {
    const response = await api.put(`/users/${userId}/password`, {
      new_password: newPassword,
    });
    return response.data;
  },

  getLatestMedia: async () => {
    const response = await api.get("/media/latest");
    return response.data;
  },

  // Manual token refresh - can be called by components when needed
  checkTokenStatus: async () => {
    return proactiveTokenRefresh();
  }
};

export const radarrApi = {
  testConnection: async (base_url: string, api_key: string) => {
    const response = await api.post("/radarr/test", { base_url, api_key });
    return response.data;
  },
  fetchProfiles: async (base_url: string, api_key: string) => {
    const response = await api.post("/radarr/qualityprofiles", {
      base_url,
      api_key,
    });
    console.log(response.data);
    return response.data;
  },
  fetchFolders: async (base_url: string, api_key: string) => {
    const response = await api.post("/radarr/rootfolders", {
      base_url,
      api_key,
    });
    return response.data;
  },
};

export const sonarrApi = {
  testConnection: async (base_url: string, api_key: string) => {
    const response = await api.post("/sonarr/test", { base_url, api_key });
    return response.data;
  },
  fetchProfiles: async (base_url: string, api_key: string) => {
    const response = await api.post("/sonarr/qualityprofiles", { base_url, api_key });
    return response.data;
  },
  fetchFolders: async (base_url: string, api_key: string) => {
    const response = await api.post("/sonarr/rootfolders", { base_url, api_key });
    return response.data;
  },
};
