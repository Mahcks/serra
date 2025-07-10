import { useState } from "react";
import { Button } from "@/components/ui/button";
import { mediaServerApi, backendApi, radarrApi, sonarrApi, api } from "@/lib/api";
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
  is_4k?: boolean;
  testStatus?: "idle" | "testing" | "success" | "error";
  testError?: string;
  profiles?: RadarrQualityProfile[];
  profilesLoading?: boolean;
  folders?: any[];
  foldersLoading?: boolean;
}

interface DownloadClient {
  id: string;
  type: "qbittorrent" | "sabnzbd";
  name: string;
  host: string;
  port: number;
  username?: string;
  password?: string;
  api_key?: string; // For SABnzbd
  use_ssl?: boolean;
  testStatus?: "idle" | "testing" | "success" | "error";
  testError?: string;
}

interface ServerConfig {
  type: Provider;
  url: string;
  apiKey: string;
  requestSystem: RequestSystem;
  requestSystemUrl: string;
  radarr: ArrInstance[];
  sonarr: ArrInstance[];
  downloadClients: DownloadClient[];
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
  const needsConnection = instance.testStatus !== "success";
  
  return (
    <div className="border rounded-lg p-4 mb-4 bg-white shadow-sm">
      {/* Connection Notice */}
      {needsConnection && (
        <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <p className="text-sm text-blue-800">
            <span className="font-medium">💡 Getting Started:</span> Fill in the connection details below and click "Test Connection" to load quality profiles and root folders automatically.
          </p>
        </div>
      )}
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Basic Connection Fields */}
        <div>
          <label className="block text-sm font-medium mb-1">
            Instance Name
            <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={instance.name || ""}
            onChange={(e) => onChange({ ...instance, name: e.target.value })}
            placeholder={`My ${type === "radarr" ? "Movie" : "TV"} Server`}
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Base URL
            <span className="text-red-500">*</span>
          </label>
          <input
            type="url"
            className="w-full p-2 rounded-md border"
            value={instance.base_url || ""}
            onChange={(e) =>
              onChange({ ...instance, base_url: e.target.value })
            }
            placeholder={`http://localhost:${type === "radarr" ? "7878" : "8989"}`}
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            API Key
            <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={instance.api_key || ""}
            onChange={(e) => onChange({ ...instance, api_key: e.target.value })}
            placeholder="Found in Settings → General → Security"
          />
        </div>
        <div className="flex items-center">
          <input
            type="checkbox"
            id={`4k-${instance.id}`}
            className="mr-2"
            checked={instance.is_4k || false}
            onChange={(e) =>
              onChange({ ...instance, is_4k: e.target.checked })
            }
          />
          <label htmlFor={`4k-${instance.id}`} className="text-sm font-medium">
            4K Instance
          </label>
          <span className="ml-2 text-xs text-muted-foreground">
            (for high-resolution content)
          </span>
        </div>
        
        {/* Advanced Configuration Fields */}
        <div>
          <label className="block text-sm font-medium mb-1 flex items-center gap-2">
            Quality Profile
            <span className="text-red-500">*</span>
            {needsConnection && (
              <span className="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded">
                Test connection first
              </span>
            )}
          </label>
          <select
            className={`w-full p-2 rounded-md border ${needsConnection ? 'bg-gray-50' : ''}`}
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
              {instance.profilesLoading 
                ? "Loading profiles..." 
                : needsConnection
                  ? "Connect to load profiles"
                  : "Select a quality profile"
              }
            </option>
            {Array.isArray(instance.profiles) &&
              instance.profiles.map((profile) => (
                <option key={profile.id} value={String(profile.id)}>
                  {profile.name}
                </option>
              ))}
          </select>
          {instance.profilesLoading && (
            <p className="text-xs text-blue-600 mt-1">Loading profiles...</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1 flex items-center gap-2">
            Root Folder Path
            <span className="text-red-500">*</span>
            {needsConnection && (
              <span className="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded">
                Test connection first
              </span>
            )}
          </label>
          <select
            className={`w-full p-2 rounded-md border ${needsConnection ? 'bg-gray-50' : ''}`}
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
              {instance.foldersLoading 
                ? "Loading folders..." 
                : needsConnection
                  ? "Connect to load folders"
                  : "Select a root folder"
              }
            </option>
            {instance.folders?.map((folder: any) => (
              <option key={folder.id || folder.path} value={folder.path}>
                {folder.path}
              </option>
            ))}
          </select>
          {instance.foldersLoading && (
            <p className="text-xs text-blue-600 mt-1">Loading folders...</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Minimum Availability
            {type === "radarr" && <span className="text-red-500">*</span>}
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
          <p className="text-xs text-muted-foreground mt-1">
            When {type === "radarr" ? "movies" : "episodes"} become available for download
          </p>
        </div>
      </div>
      
      {/* Action Buttons */}
      <div className="flex justify-between mt-6 pt-4 border-t items-center">
        <Button variant="destructive" size="sm" onClick={onRemove}>
          Remove Instance
        </Button>
        
        {/* Test Connection Button for Radarr and Sonarr */}
        {(type === "radarr" || type === "sonarr") && (
          <div className="flex flex-col items-end">
            <Button
              type="button"
              variant={instance.testStatus === "success" ? "default" : "outline"}
              size="sm"
              disabled={
                instance.testStatus === "testing" || 
                instance.profilesLoading ||
                !instance.base_url ||
                !instance.api_key
              }
              className={instance.testStatus === "success" ? "bg-green-600 hover:bg-green-700" : ""}
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
                ? "Testing Connection..."
                : instance.testStatus === "success"
                  ? "✓ Test Again"
                  : "Test Connection"}
            </Button>
            {(!instance.base_url || !instance.api_key) && (
              <p className="text-xs text-muted-foreground mt-1">
                Fill in URL and API key first
              </p>
            )}
          </div>
        )}
      </div>
      
      {/* Show test result */}
      {instance.testStatus === "error" && (
        <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-800 text-sm">
            <span className="font-medium">Connection Failed:</span> {instance.testError}
          </p>
        </div>
      )}
      {instance.testStatus === "success" && (
        <div className="mt-3 p-3 bg-green-50 border border-green-200 rounded-md">
          <p className="text-green-800 text-sm">
            <span className="font-medium">✓ Connection Successful!</span> Quality profiles and root folders have been loaded.
          </p>
        </div>
      )}
    </div>
  );
}

// Download Client Form Component
function DownloadClientForm({
  client,
  onChange,
  onRemove,
}: {
  client: DownloadClient;
  onChange: (updated: DownloadClient) => void;
  onRemove: () => void;
}) {
  const isQBittorrent = client.type === "qbittorrent";
  const isSABnzbd = client.type === "sabnzbd";
  const needsCredentials = isQBittorrent || isSABnzbd;
  const canTest = client.host && client.port && (
    !needsCredentials || 
    (isQBittorrent && client.username && client.password) ||
    (isSABnzbd && client.api_key)
  );

  return (
    <div className="border rounded-lg p-4 mb-4 bg-white shadow-sm">
      {/* Client Type Notice */}
      <div className="mb-4 p-3 bg-gray-50 border border-gray-200 rounded-md">
        <div className="flex items-center gap-2">
          <span className="font-medium">
            {isQBittorrent ? "🌊 qBittorrent Client" : "📰 SABnzbd Client"}
          </span>
          <span className="text-xs bg-gray-200 px-2 py-1 rounded-full">
            {client.type}
          </span>
        </div>
        <p className="text-sm text-muted-foreground mt-1">
          {isQBittorrent 
            ? "For torrent downloads. Requires username and password authentication."
            : "For usenet downloads. Requires API key authentication."
          }
        </p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Basic Configuration */}
        <div>
          <label className="block text-sm font-medium mb-1">
            Client Name
            <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={client.name || ""}
            onChange={(e) => onChange({ ...client, name: e.target.value })}
            placeholder={`My ${isQBittorrent ? "Torrent" : "Usenet"} Client`}
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Client Type
            <span className="text-red-500">*</span>
          </label>
          <select
            className="w-full p-2 rounded-md border"
            value={client.type}
            onChange={(e) => {
              const newType = e.target.value as "qbittorrent" | "sabnzbd";
              onChange({ 
                ...client, 
                type: newType,
                port: newType === "qbittorrent" ? 8080 : 8080, // Default ports
                username: "",
                password: "",
                api_key: ""
              });
            }}
          >
            <option value="qbittorrent">qBittorrent (Torrents)</option>
            <option value="sabnzbd">SABnzbd (Usenet)</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Host
            <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            className="w-full p-2 rounded-md border"
            value={client.host || ""}
            onChange={(e) => onChange({ ...client, host: e.target.value })}
            placeholder="localhost"
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">
            Port
            <span className="text-red-500">*</span>
          </label>
          <input
            type="number"
            className="w-full p-2 rounded-md border"
            value={client.port || ""}
            onChange={(e) => onChange({ ...client, port: parseInt(e.target.value) || 0 })}
            placeholder={isQBittorrent ? "8080 (default)" : "8080 (default)"}
          />
        </div>
        
        {/* Authentication Fields */}
        {isQBittorrent && (
          <>
            <div>
              <label className="block text-sm font-medium mb-1">
                Username
                <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                className="w-full p-2 rounded-md border"
                value={client.username || ""}
                onChange={(e) => onChange({ ...client, username: e.target.value })}
                placeholder="qBittorrent username"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                Password
                <span className="text-red-500">*</span>
              </label>
              <input
                type="password"
                className="w-full p-2 rounded-md border"
                value={client.password || ""}
                onChange={(e) => onChange({ ...client, password: e.target.value })}
                placeholder="qBittorrent password"
              />
            </div>
          </>
        )}
        
        {isSABnzbd && (
          <div>
            <label className="block text-sm font-medium mb-1">
              API Key
              <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              className="w-full p-2 rounded-md border"
              value={client.api_key || ""}
              onChange={(e) => onChange({ ...client, api_key: e.target.value })}
              placeholder="Found in SABnzbd Config → General"
            />
          </div>
        )}
        
        <div className="flex items-center">
          <input
            type="checkbox"
            id={`ssl-${client.id}`}
            className="mr-2"
            checked={client.use_ssl || false}
            onChange={(e) => onChange({ ...client, use_ssl: e.target.checked })}
          />
          <label htmlFor={`ssl-${client.id}`} className="text-sm font-medium">
            Use SSL/HTTPS
          </label>
          <span className="ml-2 text-xs text-muted-foreground">
            (if using secure connection)
          </span>
        </div>
      </div>
      
      {/* Action Buttons */}
      <div className="flex justify-between mt-6 pt-4 border-t items-center">
        <Button variant="destructive" size="sm" onClick={onRemove}>
          Remove Client
        </Button>
        
        <div className="flex flex-col items-end">
          <Button
            type="button"
            variant={client.testStatus === "success" ? "default" : "outline"}
            size="sm"
            disabled={client.testStatus === "testing" || !canTest}
            className={client.testStatus === "success" ? "bg-green-600 hover:bg-green-700" : ""}
            onClick={async () => {
              onChange({ ...client, testStatus: "testing", testError: "" });
              try {
                const response = await api.post("/downloadclient/test", {
                  type: client.type,
                  host: client.host,
                  port: client.port,
                  username: client.username || "",
                  password: client.password || "",
                  api_key: client.api_key || "",
                  use_ssl: client.use_ssl || false,
                });
                onChange({ ...client, testStatus: "success", testError: "" });
              } catch (err: any) {
                const errorMessage = err.response?.data?.detail || err.message || "Failed to connect";
                onChange({
                  ...client,
                  testStatus: "error",
                  testError: errorMessage,
                });
              }
            }}
          >
            {client.testStatus === "testing" 
              ? "Testing Connection..." 
              : client.testStatus === "success"
                ? "✓ Test Again"
                : "Test Connection"}
          </Button>
          {!canTest && (
            <p className="text-xs text-muted-foreground mt-1">
              {isQBittorrent 
                ? "Fill in host, port, username, and password" 
                : "Fill in host, port, and API key"}
            </p>
          )}
        </div>
      </div>
      
      {/* Status Messages */}
      {client.testStatus === "error" && (
        <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-800 text-sm">
            <span className="font-medium">Connection Failed:</span> {client.testError}
          </p>
        </div>
      )}
      {client.testStatus === "success" && (
        <div className="mt-3 p-3 bg-green-50 border border-green-200 rounded-md">
          <p className="text-green-800 text-sm">
            <span className="font-medium">✓ Connection Successful!</span> Download client is ready to use.
          </p>
        </div>
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
    downloadClients: [],
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
        serverConfig.sonarr,
        serverConfig.downloadClients
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
    
    // Validation checks
    const hasAnyService = serverConfig.radarr.length > 0 || serverConfig.sonarr.length > 0;
    
    if (!hasAnyService) {
      setConnectionError("Please add at least one Radarr or Sonarr instance to continue with setup.");
      return;
    }

    // Check if any required fields are missing
    const allServices = [...serverConfig.radarr, ...serverConfig.sonarr];
    const incompleteServices = allServices.filter(service => 
      !service.name || 
      !service.base_url || 
      !service.api_key || 
      !service.quality_profile || 
      !service.root_folder_path
    );

    if (incompleteServices.length > 0) {
      setConnectionError(`Please complete configuration for all services. ${incompleteServices.length} service(s) are missing required fields.`);
      return;
    }

    // Check if any services haven't been tested
    const untestedServices = allServices.filter(service => service.testStatus !== "success");
    if (untestedServices.length > 0) {
      setConnectionError(`Please test connections for all services. ${untestedServices.length} service(s) haven't been tested successfully.`);
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

  // Download Client handlers
  const handleDownloadClientChange = (id: string, updated: DownloadClient) => {
    setServerConfig((prev) => ({
      ...prev,
      downloadClients: prev.downloadClients.map((client) =>
        client.id === id ? updated : client
      ),
    }));
  };

  const handleDownloadClientRemove = (id: string) => {
    setServerConfig((prev) => ({
      ...prev,
      downloadClients: prev.downloadClients.filter((client) => client.id !== id),
    }));
  };

  const handleDownloadClientAdd = () => {
    setServerConfig((prev) => ({
      ...prev,
      downloadClients: [
        ...prev.downloadClients,
        {
          id: uuidv4(),
          type: "qbittorrent",
          name: "",
          host: "localhost",
          port: 8080,
          username: "",
          password: "",
          api_key: "",
          use_ssl: false,
        },
      ],
    }));
  };

  return (
    <div className="max-w-4xl mx-auto p-4 md:p-6">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {[
            { num: 1, label: "Media Server", desc: "Choose your platform" },
            { num: 2, label: "Server Config", desc: "Connection details" },
            { num: 3, label: "Request System", desc: "How to handle requests" },
            { num: 4, label: "Automation", desc: "Optional services" }
          ].map((step, index) => {
            const isCompleted = currentStep > step.num;
            const isCurrent = currentStep === step.num;
            const isAccessible = currentStep >= step.num;
            
            // Determine step completion status
            let stepStatus = "pending";
            if (isCompleted) stepStatus = "completed";
            else if (isCurrent) stepStatus = "current";
            
            // Check actual completion for better status
            if (step.num === 1 && serverConfig.type) stepStatus = "completed";
            if (step.num === 2 && connectionStatus === "success") stepStatus = "completed";
            if (step.num === 3 && serverConfig.requestSystem) stepStatus = "completed";
            
            return (
              <div
                key={step.num}
                className={`flex items-center ${step.num !== 4 ? "flex-1" : ""}`}
              >
                <div className="flex flex-col items-center">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center border-2 transition-all duration-200 ${
                      stepStatus === "completed"
                        ? "bg-green-500 border-green-500 text-white"
                        : stepStatus === "current"
                        ? "bg-blue-500 border-blue-500 text-white"
                        : isAccessible
                        ? "bg-white border-blue-300 text-blue-500 hover:border-blue-500"
                        : "bg-gray-100 border-gray-300 text-gray-400"
                    } ${isAccessible ? "cursor-pointer" : "cursor-not-allowed"}`}
                    onClick={() => isAccessible && setCurrentStep(step.num)}
                  >
                    {stepStatus === "completed" ? (
                      <span className="text-sm font-bold">✓</span>
                    ) : (
                      <span className="text-sm font-semibold">{step.num}</span>
                    )}
                  </div>
                  <div className="mt-2 text-center">
                    <div className={`text-xs md:text-sm font-medium ${
                      isCurrent ? "text-blue-600" : isCompleted ? "text-green-600" : "text-gray-500"
                    }`}>
                      {step.label}
                    </div>
                    <div className="text-xs text-gray-400 mt-1 hidden md:block">
                      {step.desc}
                    </div>
                  </div>
                </div>
                {step.num !== 4 && (
                  <div
                    className={`flex-1 h-0.5 mx-4 transition-colors duration-200 ${
                      isCompleted ? "bg-green-400" : "bg-gray-200"
                    }`}
                  />
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="mt-8">
        {currentStep === 1 && (
          <div className="space-y-6">
            <div className="text-center space-y-2">
              <h2 className="text-2xl font-bold">Choose Your Media Server</h2>
              <p className="text-muted-foreground">
                Select the media server platform you want to connect to Serra
              </p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div
                className={`border-2 rounded-lg p-6 cursor-pointer transition-all hover:border-blue-300 hover:bg-blue-50 ${
                  serverConfig.type === "emby" ? "border-blue-500 bg-blue-50" : "border-gray-200"
                }`}
                onClick={() => handleServerTypeSelect("emby")}
              >
                <div className="text-center space-y-3">
                  <div className="text-3xl">🎬</div>
                  <div>
                    <div className="text-xl font-semibold">Emby</div>
                    <div className="text-sm text-muted-foreground mt-1">
                      Premium media server with advanced features
                    </div>
                  </div>
                  <div className="text-xs text-gray-500 space-y-1">
                    <div>• Rich metadata and artwork</div>
                    <div>• Live TV and DVR support</div>
                    <div>• Mobile sync capabilities</div>
                  </div>
                </div>
              </div>
              
              <div
                className={`border-2 rounded-lg p-6 cursor-pointer transition-all hover:border-blue-300 hover:bg-blue-50 ${
                  serverConfig.type === "jellyfin" ? "border-blue-500 bg-blue-50" : "border-gray-200"
                }`}
                onClick={() => handleServerTypeSelect("jellyfin")}
              >
                <div className="text-center space-y-3">
                  <div className="text-3xl">🆓</div>
                  <div>
                    <div className="text-xl font-semibold">Jellyfin</div>
                    <div className="text-sm text-muted-foreground mt-1">
                      Free and open-source media server
                    </div>
                  </div>
                  <div className="text-xs text-gray-500 space-y-1">
                    <div>• Completely free to use</div>
                    <div>• Open source and community-driven</div>
                    <div>• No premium features or limits</div>
                  </div>
                </div>
              </div>
            </div>
            
            {serverConfig.type && (
              <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
                <p className="text-sm text-blue-800">
                  <span className="font-medium">✓ {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} Selected</span>
                  <br />
                  Next, we'll configure the connection to your {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} server.
                </p>
              </div>
            )}
          </div>
        )}

        {currentStep === 2 && (
          <form onSubmit={handleServerConfigSubmit} className="space-y-6">
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">
                Configure {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} Server
              </h2>
              <p className="text-muted-foreground">
                Connect Serra to your {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} server to enable media management
              </p>
            </div>

            {/* Configuration Guide */}
            <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
              <h3 className="font-medium text-blue-900 mb-2">
                🔧 {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} Setup Guide
              </h3>
              <div className="text-sm text-blue-800 space-y-1">
                <div>1. Make sure your {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} server is running</div>
                <div>2. Find your API key in Settings → Advanced → API Keys</div>
                <div>3. Ensure Serra can reach your server URL</div>
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Server URL
                  <span className="text-red-500">*</span>
                </label>
                <input
                  type="url"
                  required
                  className={`w-full p-3 rounded-md border transition-colors ${
                    connectionStatus === "error" 
                      ? "border-red-300 focus:border-red-500" 
                      : connectionStatus === "success"
                      ? "border-green-300 focus:border-green-500"
                      : "border-gray-300 focus:border-blue-500"
                  }`}
                  placeholder={`http://localhost:8096`}
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
                <p className="text-xs text-gray-500 mt-1">
                  Default port: 8096 for {serverConfig.type === "emby" ? "Emby" : "Jellyfin"}
                </p>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">
                  API Key
                  <span className="text-red-500">*</span>
                </label>
                <input
                  type="password"
                  className="w-full p-3 rounded-md border focus:border-blue-500 transition-colors"
                  placeholder="Your API key from server settings..."
                  value={serverConfig.apiKey}
                  onChange={(e) => {
                    setServerConfig((prev) => ({
                      ...prev,
                      apiKey: e.target.value,
                    }));
                  }}
                />
                <p className="text-xs text-gray-500 mt-1">
                  {serverConfig.type === "emby" 
                    ? "Found in Emby Settings → Advanced → API Keys"
                    : "Found in Jellyfin Settings → Advanced → API Keys"
                  }
                </p>
              </div>

              {/* Connection Status */}
              {connectionStatus === "success" && (
                <div className="p-3 bg-green-50 border border-green-200 rounded-md">
                  <p className="text-sm text-green-800">
                    <span className="font-medium">✓ Connection Successful!</span>
                    <br />
                    Successfully connected to your {serverConfig.type === "emby" ? "Emby" : "Jellyfin"} server.
                  </p>
                </div>
              )}
              
              {connectionStatus === "error" && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                  <p className="text-sm text-red-800">
                    <span className="font-medium">Connection Failed:</span>
                    <br />
                    {connectionError}
                  </p>
                  <div className="mt-2 text-xs text-red-700">
                    <div className="font-medium">Common issues:</div>
                    <div>• Check if the server URL is correct and accessible</div>
                    <div>• Verify the API key is valid</div>
                    <div>• Ensure the server is running</div>
                  </div>
                </div>
              )}
            </div>
            
            <div className="flex justify-between pt-4 border-t">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(1)}
              >
                Back
              </Button>
              <Button 
                type="submit" 
                disabled={connectionStatus === "testing" || !serverConfig.url.trim()}
                className={connectionStatus === "success" ? "bg-green-600 hover:bg-green-700" : ""}
              >
                {connectionStatus === "testing" 
                  ? "Testing Connection..." 
                  : connectionStatus === "success"
                  ? "✓ Continue"
                  : "Test & Continue"}
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
          <form onSubmit={handleAutomationSubmit} className="space-y-6">
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">Automation Services</h2>
              <p className="text-muted-foreground">
                Configure your media management and download automation services
              </p>
            </div>

            {/* Radarr Section */}
            <div className="space-y-4">
              <div className="border-b pb-2">
                <h3 className="text-xl font-semibold flex items-center gap-2">
                  <span className="text-blue-600">🎬</span>
                  Radarr (Movies)
                </h3>
                <p className="text-sm text-muted-foreground mt-1">
                  Manage and automatically download movies
                  {serverConfig.radarr.length > 0 && (
                    <span className="ml-2 text-blue-600 font-medium">
                      • {serverConfig.radarr.length} instance{serverConfig.radarr.length !== 1 ? 's' : ''} configured
                    </span>
                  )}
                </p>
              </div>
              
              {serverConfig.radarr.length === 0 && (
                <div className="border-2 border-dashed border-gray-200 rounded-lg p-6 text-center">
                  <p className="text-muted-foreground mb-3">No Radarr instances configured</p>
                  <Button
                    variant="outline"
                    type="button"
                    onClick={() => handleArrInstanceAdd("radarr")}
                  >
                    Add Your First Radarr Instance
                  </Button>
                </div>
              )}
              
              {serverConfig.radarr.map((instance, index) => (
                <div key={instance.id} className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
                    <span>Radarr Instance #{index + 1}</span>
                    {instance.is_4k && (
                      <span className="bg-purple-100 text-purple-800 px-2 py-1 rounded-full text-xs font-semibold">
                        4K
                      </span>
                    )}
                    {instance.testStatus === "success" && (
                      <span className="bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-semibold">
                        ✓ Connected
                      </span>
                    )}
                  </div>
                  <ArrInstanceForm
                    instance={instance}
                    onChange={(updated) =>
                      handleArrInstanceChange("radarr", instance.id, updated)
                    }
                    onRemove={() => handleArrInstanceRemove("radarr", instance.id)}
                    type="radarr"
                  />
                </div>
              ))}
              
              {serverConfig.radarr.length > 0 && (
                <Button
                  variant="outline"
                  type="button"
                  onClick={() => handleArrInstanceAdd("radarr")}
                  className="w-full"
                >
                  + Add Another Radarr Instance
                </Button>
              )}
            </div>

            {/* Sonarr Section */}
            <div className="space-y-4">
              <div className="border-b pb-2">
                <h3 className="text-xl font-semibold flex items-center gap-2">
                  <span className="text-green-600">📺</span>
                  Sonarr (TV Shows)
                </h3>
                <p className="text-sm text-muted-foreground mt-1">
                  Manage and automatically download TV shows and series
                  {serverConfig.sonarr.length > 0 && (
                    <span className="ml-2 text-green-600 font-medium">
                      • {serverConfig.sonarr.length} instance{serverConfig.sonarr.length !== 1 ? 's' : ''} configured
                    </span>
                  )}
                </p>
              </div>
              
              {serverConfig.sonarr.length === 0 && (
                <div className="border-2 border-dashed border-gray-200 rounded-lg p-6 text-center">
                  <p className="text-muted-foreground mb-3">No Sonarr instances configured</p>
                  <Button
                    variant="outline"
                    type="button"
                    onClick={() => handleArrInstanceAdd("sonarr")}
                  >
                    Add Your First Sonarr Instance
                  </Button>
                </div>
              )}
              
              {serverConfig.sonarr.map((instance, index) => (
                <div key={instance.id} className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
                    <span>Sonarr Instance #{index + 1}</span>
                    {instance.is_4k && (
                      <span className="bg-purple-100 text-purple-800 px-2 py-1 rounded-full text-xs font-semibold">
                        4K
                      </span>
                    )}
                    {instance.testStatus === "success" && (
                      <span className="bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-semibold">
                        ✓ Connected
                      </span>
                    )}
                  </div>
                  <ArrInstanceForm
                    instance={instance}
                    onChange={(updated) =>
                      handleArrInstanceChange("sonarr", instance.id, updated)
                    }
                    onRemove={() => handleArrInstanceRemove("sonarr", instance.id)}
                    type="sonarr"
                  />
                </div>
              ))}
              
              {serverConfig.sonarr.length > 0 && (
                <Button
                  variant="outline"
                  type="button"
                  onClick={() => handleArrInstanceAdd("sonarr")}
                  className="w-full"
                >
                  + Add Another Sonarr Instance
                </Button>
              )}
            </div>

            {/* Download Clients Section */}
            <div className="space-y-4">
              <div className="border-b pb-2">
                <h3 className="text-xl font-semibold flex items-center gap-2">
                  <span className="text-orange-600">⬇️</span>
                  Download Clients
                  <span className="text-sm font-normal text-muted-foreground">(Optional)</span>
                </h3>
                <p className="text-sm text-muted-foreground mt-1">
                  Configure torrent and usenet download clients
                  {serverConfig.downloadClients.length > 0 && (
                    <span className="ml-2 text-orange-600 font-medium">
                      • {serverConfig.downloadClients.length} client{serverConfig.downloadClients.length !== 1 ? 's' : ''} configured
                    </span>
                  )}
                </p>
              </div>
              
              {serverConfig.downloadClients.length === 0 && (
                <div className="border-2 border-dashed border-gray-200 rounded-lg p-6 text-center">
                  <p className="text-muted-foreground mb-3">No download clients configured</p>
                  <p className="text-xs text-muted-foreground mb-3">
                    You can add these later in settings if needed
                  </p>
                  <Button
                    variant="outline"
                    type="button"
                    onClick={handleDownloadClientAdd}
                  >
                    Add Download Client
                  </Button>
                </div>
              )}
              
              {serverConfig.downloadClients.map((client, index) => (
                <div key={client.id} className="space-y-2">
                  <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
                    <span>Download Client #{index + 1}</span>
                    <span className="bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs font-semibold">
                      {client.type}
                    </span>
                    {client.testStatus === "success" && (
                      <span className="bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-semibold">
                        ✓ Connected
                      </span>
                    )}
                  </div>
                  <DownloadClientForm
                    client={client}
                    onChange={(updated) => handleDownloadClientChange(client.id, updated)}
                    onRemove={() => handleDownloadClientRemove(client.id)}
                  />
                </div>
              ))}
              
              {serverConfig.downloadClients.length > 0 && (
                <Button
                  variant="outline"
                  type="button"
                  onClick={handleDownloadClientAdd}
                  className="w-full"
                >
                  + Add Another Download Client
                </Button>
              )}
            </div>

            {/* Setup Summary */}
            {(serverConfig.radarr.length > 0 || serverConfig.sonarr.length > 0 || serverConfig.downloadClients.length > 0) && (
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                <h3 className="font-medium text-gray-900 mb-3">📋 Setup Summary</h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span>Media Server:</span>
                    <span className="font-medium">{serverConfig.type === "emby" ? "Emby" : "Jellyfin"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Request System:</span>
                    <span className="font-medium capitalize">{serverConfig.requestSystem.replace('_', ' ')}</span>
                  </div>
                  {serverConfig.radarr.length > 0 && (
                    <div className="flex justify-between">
                      <span>Radarr Instances:</span>
                      <span className="font-medium">{serverConfig.radarr.length}</span>
                    </div>
                  )}
                  {serverConfig.sonarr.length > 0 && (
                    <div className="flex justify-between">
                      <span>Sonarr Instances:</span>
                      <span className="font-medium">{serverConfig.sonarr.length}</span>
                    </div>
                  )}
                  {serverConfig.downloadClients.length > 0 && (
                    <div className="flex justify-between">
                      <span>Download Clients:</span>
                      <span className="font-medium">{serverConfig.downloadClients.length}</span>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 sm:justify-between pt-6 border-t">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(3)}
                className="order-2 sm:order-1"
              >
                Back
              </Button>
              <Button 
                type="submit" 
                disabled={
                  setupMutation.isPending || 
                  (serverConfig.radarr.length === 0 && serverConfig.sonarr.length === 0)
                }
                className="min-w-[140px] order-1 sm:order-2"
              >
                {setupMutation.isPending ? (
                  <span className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    Setting up...
                  </span>
                ) : (
                  "Complete Setup"
                )}
              </Button>
            </div>
            
            {connectionError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-800">{connectionError}</p>
              </div>
            )}
          </form>
        )}
      </div>
    </div>
  );
}
