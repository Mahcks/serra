import { createContext, useContext } from "react";
import { useQuery } from "@tanstack/react-query";
import { backendApi } from "./api";
import type { RequestSystem } from "@/types";

type Settings = {
  request_system: RequestSystem;
  request_system_url?: string;
};

const SettingsContext = createContext<{
  settings: Settings | null;
  isLoading: boolean;
  error: any;
} | null>(null);

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const { data, isLoading, error } = useQuery({
    queryKey: ["settings"],
    queryFn: backendApi.getSettings,
    staleTime: 5 * 60 * 1000,
  });

  return (
    <SettingsContext.Provider value={{ settings: data, isLoading, error }}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const ctx = useContext(SettingsContext);
  if (!ctx) throw new Error("useSettings must be used within a SettingsProvider");
  return ctx;
}
