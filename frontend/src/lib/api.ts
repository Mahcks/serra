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
    await backendApi.refreshToken();
    // Only log successful refreshes in development
    if (process.env.NODE_ENV === 'development') {
      console.log("✅ Token refresh check completed");
    }
  } catch (error: any) {
    // Only log actual errors (not "token still valid" responses)
    if (error.response?.status !== 200) {
      console.log("⚠️ Token refresh failed:", error.response?.data?.message || error.message);
    }
  }
};

// Track last refresh time to avoid excessive refresh calls
let lastRefreshTime = 0;
const REFRESH_COOLDOWN = 30 * 1000; // 30 seconds

// Add request interceptor with smarter token refresh logic
api.interceptors.request.use(
  async (config) => {
    // Skip token check for auth endpoints to prevent infinite loops
    if (config.url?.includes('/auth/') || config.url?.includes('/setup')) {
      return config;
    }

    // Only proactively refresh for critical endpoints and respect cooldown
    const now = Date.now();
    const shouldRefresh = (
      config.url?.includes('/me') || // Always refresh before checking auth status
      config.url?.includes('/permissions') // Critical auth-dependent endpoints
    ) && (now - lastRefreshTime > REFRESH_COOLDOWN);

    if (shouldRefresh) {
      try {
        await proactiveTokenRefresh();
        lastRefreshTime = now;
      } catch (error) {
        // Continue with request even if refresh fails - the response interceptor will handle 401s
      }
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
    sonarr?: any[],
    downloadClients?: any[]
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
    if (downloadClients) {
      payload.downloadClients = downloadClients;
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

export const discoverApi = {
  // Get trending media (movies and TV shows)
  getTrending: async (page: number = 1) => {
    const response = await api.get(`/discover/trending?page=${page}`);
    return response.data;
  },

  // Search for movies
  searchMovies: async (query: string, page: number = 1) => {
    const response = await api.get(`/discover/search/movie?query=${encodeURIComponent(query)}&page=${page}`);
    return response.data;
  },

  // Search for TV shows
  searchTV: async (query: string, page: number = 1) => {
    const response = await api.get(`/discover/search/tv?query=${encodeURIComponent(query)}&page=${page}`);
    return response.data;
  },

  // Search for companies/studios
  searchCompanies: async (query: string, page: number = 1) => {
    const response = await api.get(`/discover/search/company?query=${encodeURIComponent(query)}&page=${page}`);
    return response.data;
  },

  // Combined search for both movies and TV shows
  searchAll: async (query: string, page: number = 1) => {
    const [moviesResponse, tvResponse] = await Promise.all([
      api.get(`/discover/search/movie?query=${encodeURIComponent(query)}&page=${page}`),
      api.get(`/discover/search/tv?query=${encodeURIComponent(query)}&page=${page}`)
    ]);
    
    return {
      movies: moviesResponse.data,
      tv: tvResponse.data,
    };
  },

  // Get popular movies (basic - no sorting)
  getPopularMovies: async (page: number = 1) => {
    const response = await api.get(`/discover/movie/popular?page=${page}`);
    return response.data;
  },

  // Get popular TV shows (basic - no sorting)
  getPopularTV: async (page: number = 1) => {
    const response = await api.get(`/discover/tv/popular?page=${page}`);
    return response.data;
  },

  // Get movies with sorting and filtering using discover endpoint
  getMoviesWithSort: async (page: number = 1, sortBy: string = 'popularity.desc') => {
    const params = new URLSearchParams({ 
      page: page.toString(),
      sort_by: sortBy
    });
    const response = await api.get(`/discover/movie?${params.toString()}`);
    return response.data;
  },

  // Get TV shows with sorting - for now use popular endpoint until backend adds discover/tv
  getTVWithSort: async (page: number = 1, sortBy: string = 'popularity.desc') => {
    // Since there's no discover/tv endpoint yet, fall back to popular for now
    // TODO: Backend needs to add discover/tv endpoint with sorting support
    const response = await api.get(`/discover/tv/popular?page=${page}`);
    return response.data;
  },

  // Get upcoming movies
  getUpcomingMovies: async (page: number = 1) => {
    const response = await api.get(`/discover/movie/upcoming?page=${page}`);
    return response.data;
  },

  // Get upcoming TV shows (airing today/this week)
  getUpcomingTV: async (page: number = 1) => {
    const response = await api.get(`/discover/tv/upcoming?page=${page}`);
    return response.data;
  },

  // Get media details (movie or TV show)
  getMediaDetails: async (id: string, type: 'movie' | 'tv') => {
    const response = await api.get(`/discover/media/details/${id}?type=${type}`);
    return response.data;
  },

  // Get movie recommendations
  getMovieRecommendations: async (movieId: string, page: number = 1) => {
    const response = await api.get(`/discover/movie/${movieId}/recommendations?page=${page}`);
    return response.data;
  },

  // Get movie watch providers
  getMovieWatchProviders: async (movieId: string) => {
    const response = await api.get(`/discover/movie/${movieId}/watch-providers`);
    return response.data;
  },

  // Get watch providers for movies or TV shows
  getWatchProviders: async (type: 'movie' | 'tv' = 'movie') => {
    const response = await api.get(`/discover/watch/providers?type=${type}`);
    return response.data;
  },

  // Get watch provider regions
  getWatchProviderRegions: async () => {
    const response = await api.get('/discover/watch/regions');
    return response.data;
  },

  // Discover movies with comprehensive filters matching TMDB API
  discoverMovies: async (params: {
    // Basic parameters
    page?: number;
    language?: string;
    sort_by?: string;
    include_adult?: boolean;
    include_video?: boolean;
    region?: string;
    
    // Date filters
    primary_release_year?: number;
    primary_release_date_gte?: string;
    primary_release_date_lte?: string;
    release_date_gte?: string;
    release_date_lte?: string;
    year?: number;
    
    // Rating and popularity filters
    vote_average_gte?: number;
    vote_average_lte?: number;
    vote_count_gte?: number;
    vote_count_lte?: number;
    
    // Content filters
    with_genres?: string;
    without_genres?: string;
    with_companies?: string;
    without_companies?: string;
    with_keywords?: string;
    without_keywords?: string;
    with_cast?: string;
    with_crew?: string;
    with_people?: string;
    with_original_language?: string;
    with_origin_country?: string;
    
    // Runtime filters
    with_runtime_gte?: number;
    with_runtime_lte?: number;
    
    // Certification filters
    certification?: string;
    certification_gte?: string;
    certification_lte?: string;
    certification_country?: string;
    
    // Watch provider filters
    with_watch_providers?: string;
    without_watch_providers?: string;
    with_watch_monetization_types?: string;
    watch_region?: string;
    
    // Release type filter
    with_release_type?: string;
  } = {}) => {
    const searchParams = new URLSearchParams();
    Object.keys(params).forEach(key => {
      const value = params[key as keyof typeof params];
      if (value !== undefined && value !== null && value !== '') {
        searchParams.append(key, value.toString());
      }
    });
    const response = await api.get(`/discover/movie?${searchParams.toString()}`);
    return response.data;
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
