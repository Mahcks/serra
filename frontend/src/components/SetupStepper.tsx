import { useState } from "react";
import { Button } from "@/components/ui/button";
import { mediaServerApi, backendApi, radarrApi, sonarrApi } from "@/lib/api";
import { useMutation } from "@tanstack/react-query";
import {
  ProviderEmby,
  RequestSystemBuiltIn,
  type Provider,
  type RadarrQualityProfile,
  type RequestSystem,
} from "@/types";
import { v4 as uuidv4 } from "uuid";

interface ArrInstance {
  id: string;
  name: string;
  base_url: string;
  api_key: string;
  quality_profile: string;
  root_folder_path: string;
  minimum_availability: string;
  testStatus?: "idle" | "testing" | "success" | "error";
  testError?: string;
  profiles?: RadarrQualityProfile[];
  profilesLoading?: boolean;
  folders?: any[];
  foldersLoading?: boolean;
}

interface ServerConfig {
  type: Provider;
  url: string;
  apiKey: string;
  requestSystem: RequestSystem;
  requestSystemUrl: string;
  radarr: ArrInstance[];
  sonarr: ArrInstance[];
}

interface SetupStepperProps {
  onSetupComplete: () => void;
}

type ConnectionStatus = "idle" | "testing" | "success" | "error";

// Move ArrInstanceForm outside SetupStepper
function ArrInstanceForm({
  instance,
  onChange,
  onRemove,
  type,
}: {
  instance: ArrInstance;
  onChange: (updated: ArrInstance) => void;
  onRemove: () => void;
  type: "radarr" | "sonarr";
}) {
  // Always show test for radarr for now
  return (
    <div className="border rounded-lg p-4 mb-4 bg-white/80">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-1">Name</label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={instance.name || ""}
            onChange={(e) => onChange({ ...instance, name: e.target.value })}
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">Base URL</label>
          <input
            type="url"
            className="w-full p-2 rounded-md border"
            value={instance.base_url || ""}
            onChange={(e) =>
              onChange({ ...instance, base_url: e.target.value })
            }
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">API Key</label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={instance.api_key || ""}
            onChange={(e) => onChange({ ...instance, api_key: e.target.value })}
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Quality Profile
          </label>
          <select
            className="w-full p-2 rounded-md border"
            value={instance.quality_profile || ""}
            onChange={(e) =>
              onChange({ ...instance, quality_profile: e.target.value })
            }
            disabled={
              instance.profilesLoading ||
              !instance.profiles ||
              instance.profiles.length === 0
            }
          >
            <option value="">
              {instance.profilesLoading ? "Loading..." : "Select a profile"}
            </option>
            {Array.isArray(instance.profiles) &&
              instance.profiles.map((profile) => (
                <option key={profile.id} value={String(profile.id)}>
                  {profile.name}
                </option>
              ))}
          </select>
          {instance.profilesLoading && (
            <p className="text-xs text-gray-500">Loading profiles...</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Root Folder Path
          </label>
          <select
            className="w-full p-2 rounded-md border"
            value={instance.root_folder_path || ""}
            onChange={(e) =>
              onChange({ ...instance, root_folder_path: e.target.value })
            }
            disabled={
              instance.foldersLoading ||
              !instance.folders ||
              instance.folders.length === 0
            }
          >
            <option value="">
              {instance.foldersLoading ? "Loading..." : "Select a folder"}
            </option>
            {instance.folders?.map((folder: any) => (
              <option key={folder.id || folder.path} value={folder.path}>
                {folder.path}
              </option>
            ))}
          </select>
          {instance.foldersLoading && (
            <p className="text-xs text-gray-500">Loading folders...</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Minimum Availability
          </label>
          <select
            className="w-full p-2 rounded-md border"
            value={instance.minimum_availability || "released"}
            onChange={(e) =>
              onChange({ ...instance, minimum_availability: e.target.value })
            }
          >
            <option value="announced">Announced</option>
            <option value="released">Released</option>
            <option value="in_cinemas">In Cinemas</option>
          </select>
        </div>
      </div>
      <div className="flex justify-between mt-2 items-center">
        <Button variant="destructive" size="sm" onClick={onRemove}>
          Remove
        </Button>
        {/* Test Connection Button for Radarr and Sonarr */}
        {(type === "radarr" || type === "sonarr") && (
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={
              instance.testStatus === "testing" || instance.profilesLoading
            }
            onClick={async () => {
              onChange({ ...instance, testStatus: "testing", testError: "" });
              try {
                if (type === "radarr") {
                  await radarrApi.testConnection(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    profilesLoading: true,
                    testStatus: "success",
                    testError: "",
                  });
                  const profiles = await radarrApi.fetchProfiles(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    testStatus: "success",
                    testError: "",
                    profiles: profiles,
                    profilesLoading: false,
                    foldersLoading: true,
                  });
                  const folders = await radarrApi.fetchFolders(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    testStatus: "success",
                    testError: "",
                    profiles: profiles,
                    profilesLoading: false,
                    folders: folders,
                    foldersLoading: false,
                  });
                } else if (type === "sonarr") {
                  await sonarrApi.testConnection(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    profilesLoading: true,
                    testStatus: "success",
                    testError: "",
                  });
                  const profiles = await sonarrApi.fetchProfiles(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    testStatus: "success",
                    testError: "",
                    profiles: profiles,
                    profilesLoading: false,
                    foldersLoading: true,
                  });
                  const folders = await sonarrApi.fetchFolders(
                    instance.base_url,
                    instance.api_key
                  );
                  onChange({
                    ...instance,
                    testStatus: "success",
                    testError: "",
                    profiles: profiles,
                    profilesLoading: false,
                    folders: folders,
                    foldersLoading: false,
                  });
                }
              } catch (err) {
                onChange({
                  ...instance,
                  testStatus: "error",
                  testError:
                    err instanceof Error ? err.message : "Failed to connect",
                  profilesLoading: false,
                  foldersLoading: false,
                });
              }
            }}
          >
            {instance.testStatus === "testing" || instance.profilesLoading
              ? "Testing..."
              : "Test Connection"}
          </Button>
        )}
      </div>
      {/* Show test result */}
      {instance.testStatus === "error" && (
        <p className="text-red-600 text-sm mt-2">{instance.testError}</p>
      )}
      {instance.testStatus === "success" && (
        <p className="text-green-600 text-sm mt-2">Connection successful!</p>
      )}
    </div>
  );
}

export function SetupStepper({ onSetupComplete }: SetupStepperProps) {
  const [currentStep, setCurrentStep] = useState(1);
  const [serverConfig, setServerConfig] = useState<ServerConfig>({
    type: ProviderEmby,
    url: "",
    apiKey: "",
    requestSystem: RequestSystemBuiltIn,
    requestSystemUrl: "",
    radarr: [],
    sonarr: [],
  });
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>("idle");
  const [connectionError, setConnectionError] = useState<string>("");

  const testConnectionMutation = useMutation({
    mutationFn: (url: string) => mediaServerApi.testConnection(url),
    onSuccess: () => {
      setConnectionStatus("success");
      setConnectionError("");
    },
    onError: (error) => {
      setConnectionStatus("error");
      setConnectionError(
        error instanceof Error ? error.message : "Failed to connect to server"
      );
    },
  });

  const setupMutation = useMutation({
    mutationFn: () =>
      backendApi.setup(
        serverConfig.type,
        serverConfig.url,
        serverConfig.apiKey,
        serverConfig.requestSystem,
        serverConfig.requestSystem === "external"
          ? serverConfig.requestSystemUrl
          : undefined,
        serverConfig.radarr,
        serverConfig.sonarr
      ),
    onSuccess: () => {
      // Handle successful setup
      console.log("Setup completed successfully");
      onSetupComplete();
    },
    onError: (error) => {
      setConnectionError(
        error instanceof Error ? error.message : "Setup failed"
      );
    },
  });

  const handleServerTypeSelect = (type: Provider) => {
    setServerConfig((prev) => ({ ...prev, type }));
    setCurrentStep(2);
  };

  const testServerConnection = async () => {
    if (!serverConfig.url) {
      setConnectionError("Please enter a server URL");
      setConnectionStatus("error");
      return false;
    }

    setConnectionStatus("testing");
    setConnectionError("");

    try {
      await testConnectionMutation.mutateAsync(serverConfig.url);
      return true;
    } catch {
      return false;
    }
  };

  const handleServerConfigSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setConnectionError("");
    const success = await testServerConnection();
    if (success) {
      setCurrentStep(3);
    }
  };

  const handleRequestSystemSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setConnectionError("");
    if (
      serverConfig.requestSystem === "external" &&
      !serverConfig.requestSystemUrl
    ) {
      setConnectionError(
        "Please enter the URL for your external request system"
      );
      return;
    }
    setCurrentStep(4);
  };

  const handleAutomationSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setConnectionError("");
    if (serverConfig.radarr.length === 0 && serverConfig.sonarr.length === 0) {
      setConnectionError("Please add at least one Radarr or Sonarr instance.");
      return;
    }
    try {
      await setupMutation.mutateAsync();
    } catch (error) {
      // Error is handled by mutation onError
    }
  };

  // New Step: Automation Services (Radarr/Sonarr)
  const handleArrInstanceChange = (
    type: "radarr" | "sonarr",
    id: string,
    updated: ArrInstance
  ) => {
    setServerConfig((prev) => {
      const arr = prev[type].map((inst) => (inst.id === id ? updated : inst));
      return { ...prev, [type]: arr };
    });
  };

  const handleArrInstanceRemove = (type: "radarr" | "sonarr", id: string) => {
    setServerConfig((prev) => {
      const arr = prev[type].filter((inst) => inst.id !== id);
      return { ...prev, [type]: arr };
    });
  };

  const handleArrInstanceAdd = (type: "radarr" | "sonarr") => {
    setServerConfig((prev) => ({
      ...prev,
      [type]: [
        ...prev[type],
        {
          id: uuidv4(),
          name: "",
          base_url: "",
          api_key: "",
          quality_profile: "",
          root_folder_path: "",
          minimum_availability: "released",
        },
      ],
    }));
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {[1, 2, 3, 4].map((step) => (
            <div
              key={step}
              className={`flex items-center ${step !== 4 ? "flex-1" : ""}`}
            >
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center ${
                  currentStep >= step
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground"
                }`}
              >
                {step}
              </div>
              {step !== 4 && (
                <div
                  className={`flex-1 h-1 mx-2 ${
                    currentStep > step ? "bg-primary" : "bg-muted"
                  }`}
                />
              )}
            </div>
          ))}
        </div>
        <div className="flex justify-between mt-2 text-sm">
          <span>Choose Server</span>
          <span>Configure</span>
          <span>Request System</span>
          <span>Automation Services</span>
        </div>
      </div>

      <div className="mt-8">
        {currentStep === 1 && (
          <div className="space-y-4">
            <h2 className="text-2xl font-bold">Choose Your Media Server</h2>
            <div className="grid grid-cols-2 gap-4">
              <Button
                variant="outline"
                className="h-24"
                onClick={() => handleServerTypeSelect("emby")}
              >
                <div className="text-center">
                  <div className="text-xl font-semibold">Emby</div>
                  <div className="text-sm text-muted-foreground">
                    Choose Emby Server
                  </div>
                </div>
              </Button>
              <Button
                variant="outline"
                className="h-24"
                onClick={() => handleServerTypeSelect("jellyfin")}
              >
                <div className="text-center">
                  <div className="text-xl font-semibold">Jellyfin</div>
                  <div className="text-sm text-muted-foreground">
                    Choose Jellyfin Server
                  </div>
                </div>
              </Button>
            </div>
          </div>
        )}

        {currentStep === 2 && (
          <form onSubmit={handleServerConfigSubmit} className="space-y-4">
            <h2 className="text-2xl font-bold">
              Configure {serverConfig.type === "emby" ? "Emby" : "Jellyfin"}{" "}
              Server
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Server URL
                </label>
                <input
                  type="url"
                  required
                  className="w-full p-2 rounded-md border"
                  placeholder="http://your-server:8096"
                  value={serverConfig.url}
                  onChange={(e) => {
                    setServerConfig((prev) => ({
                      ...prev,
                      url: e.target.value,
                    }));
                    setConnectionStatus("idle");
                    setConnectionError("");
                  }}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  API Key
                </label>
                <input
                  type="password"
                  className="w-full p-2 rounded-md border"
                  placeholder="Enter your API key..."
                  value={serverConfig.apiKey}
                  onChange={(e) => {
                    setServerConfig((prev) => ({
                      ...prev,
                      apiKey: e.target.value,
                    }));
                  }}
                />
              </div>
              {connectionStatus === "success" && (
                <p className="mt-2 text-sm text-green-600">
                  ✓ Successfully connected to server
                </p>
              )}
              {connectionStatus === "error" && (
                <p className="mt-2 text-sm text-red-600">{connectionError}</p>
              )}
            </div>
            <div className="flex justify-between">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(1)}
              >
                Back
              </Button>
              <Button type="submit" disabled={connectionStatus === "testing"}>
                {connectionStatus === "testing" ? "Testing..." : "Next"}
              </Button>
            </div>
          </form>
        )}

        {currentStep === 3 && (
          <form onSubmit={handleRequestSystemSubmit} className="space-y-4">
            <h2 className="text-2xl font-bold">Request System Configuration</h2>
            <p className="text-muted-foreground">
              Choose how you want to handle media requests
            </p>

            <div className="space-y-4">
              <div className="space-y-2">
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="requestSystem"
                    value="built_in"
                    checked={serverConfig.requestSystem === "built_in"}
                    onChange={(e) => {
                      setServerConfig((prev) => ({
                        ...prev,
                        requestSystem: e.target.value as
                          | "built_in"
                          | "external",
                      }));
                      setConnectionError("");
                    }}
                  />
                  <span className="font-medium">Built-in Request System</span>
                </label>

                <p className="text-sm text-muted-foreground ml-6">
                  Use Serra's built-in request system for managing media
                  requests
                </p>
              </div>

              <div className="space-y-2">
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="requestSystem"
                    value="external"
                    checked={serverConfig.requestSystem === "external"}
                    onChange={(e) => {
                      setServerConfig((prev) => ({
                        ...prev,
                        requestSystem: e.target.value as
                          | "built_in"
                          | "external",
                      }));
                      setConnectionError("");
                    }}
                  />
                  <span className="font-medium">External Request System</span>
                </label>
                <p className="text-sm text-muted-foreground ml-6">
                  Use an external system like Jellyseerr in an iframe
                </p>
              </div>

              {serverConfig.requestSystem === "external" && (
                <div>
                  <label className="block text-sm font-medium mb-1">
                    External Request System URL
                  </label>
                  <input
                    type="url"
                    required
                    className="w-full p-2 rounded-md border"
                    placeholder="http://jellyseerr:5055"
                    value={serverConfig.requestSystemUrl}
                    onChange={(e) => {
                      setServerConfig((prev) => ({
                        ...prev,
                        requestSystemUrl: e.target.value,
                      }));
                      setConnectionError("");
                    }}
                  />
                  <p className="text-sm text-muted-foreground mt-1">
                    Enter the URL of your external request system (e.g.,
                    Jellyseerr)
                  </p>
                </div>
              )}

              {connectionError && (
                <p className="text-sm text-red-600">{connectionError}</p>
              )}
            </div>

            <div className="flex justify-between">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(2)}
              >
                Back
              </Button>
              <Button type="submit">Next</Button>
            </div>
          </form>
        )}

        {currentStep === 4 && (
          <form onSubmit={handleAutomationSubmit} className="space-y-4">
            <h2 className="text-2xl font-bold mb-4">Automation Services</h2>
            <h3 className="text-lg font-semibold mb-2">Radarr Instances</h3>
            {serverConfig.radarr.map((instance) => (
              <ArrInstanceForm
                key={instance.id}
                instance={instance}
                onChange={(updated) =>
                  handleArrInstanceChange("radarr", instance.id, updated)
                }
                onRemove={() => handleArrInstanceRemove("radarr", instance.id)}
                type="radarr"
              />
            ))}
            <Button
              variant="outline"
              className="mb-6"
              type="button"
              onClick={() => handleArrInstanceAdd("radarr")}
            >
              Add Radarr Instance
            </Button>
            <h3 className="text-lg font-semibold mt-6 mb-2">
              Sonarr Instances
            </h3>
            {serverConfig.sonarr.map((instance) => (
              <ArrInstanceForm
                key={instance.id}
                instance={instance}
                onChange={(updated) =>
                  handleArrInstanceChange("sonarr", instance.id, updated)
                }
                onRemove={() => handleArrInstanceRemove("sonarr", instance.id)}
                type="sonarr"
              />
            ))}
            <Button
              variant="outline"
              type="button"
              onClick={() => handleArrInstanceAdd("sonarr")}
            >
              Add Sonarr Instance
            </Button>
            <div className="flex justify-between mt-8">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(3)}
              >
                Back
              </Button>
              <Button type="submit" disabled={setupMutation.isPending}>
                {setupMutation.isPending ? "Setting up..." : "Complete Setup"}
              </Button>
            </div>
            {connectionError && (
              <p className="text-sm text-red-600 mt-2">{connectionError}</p>
            )}
          </form>
        )}
      </div>
    </div>
  );
}
