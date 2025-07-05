import { createContext, useContext } from "react";
import { useQuery } from "@tanstack/react-query";
import { backendApi } from "./api";
import { useAuth } from "./auth";
import type { RequestSystem } from "@/types";

type Settings = {
  request_system: RequestSystem;
  request_system_url?: string;
  download_visibility: "own" | "all";
};

const SettingsContext = createContext<{
  settings: Settings | null;
  isLoading: boolean;
  error: any;
} | null>(null);

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  console.log("⚙️ SettingsProvider render - isAuthenticated:", isAuthenticated);

  const { data, isLoading, error } = useQuery({
    queryKey: ["settings"],
    queryFn: async () => {
      console.log(
        "⚙️ Making /settings API call - isAuthenticated:",
        isAuthenticated
      );
      return await backendApi.getSettings();
    },
    staleTime: 5 * 60 * 1000,
    enabled: isAuthenticated,
  });

  return (
    <SettingsContext.Provider value={{ settings: data, isLoading, error }}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const ctx = useContext(SettingsContext);
  if (!ctx)
    throw new Error("useSettings must be used within a SettingsProvider");
  return ctx;
}
