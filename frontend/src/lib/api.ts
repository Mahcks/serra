import axios from "axios";
import { QueryClient } from "@tanstack/react-query";
import type { Provider } from "@/types";

// Create an axios instance with default config
export const api = axios.create({
  baseURL: "http://localhost:9090/v1",
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // This is important for CORS with credentials
});

// Create a query client
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

// Media server API functions
export const mediaServerApi = {
  testConnection: async (url: string) => {
    const normalizedUrl = url.replace(/\/+$/, "");
    const response = await api.get(`${normalizedUrl}/System/Info/Public`);
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
