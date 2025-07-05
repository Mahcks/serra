import { useState, useMemo, useEffect, useCallback } from "react";
import { Progress } from "@/components/ui/progress";
import {
  useWebSocketContext,
  useWebSocketMessage,
  useWebSocketEvent,
  WebSocketEvent,
  WebSocketState,
} from "@/lib/WebSocketContext";
import { type DownloadProgressBatchPayload, type Download } from "@/types";
import {
  DownloadCloud,
  Film,
  Tv,
  Search,
  InfinityIcon,
  DownloadIcon,
  Pause,
  AlertTriangle,
  CheckCircle,
  Clock,
  ArrowUp,
  Filter,
  TrendingDown,
  Activity,
  Play,
  X,
  RotateCcw,
  HardDrive,
  Zap,
  User,
  Settings,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useQuery } from "@tanstack/react-query";
import { backendApi } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useSettings } from "@/lib/settings";
import { Switch } from "@/components/ui/switch";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

// Label component
const Label = ({
  htmlFor,
  className = "",
  children,
  ...props
}: {
  htmlFor?: string;
  className?: string;
  children: React.ReactNode;
  [key: string]: any;
}) => (
  <label
    htmlFor={htmlFor}
    className={`text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 ${className}`}
    {...props}
  >
    {children}
  </label>
);

// Helper function to format file size
const formatFileSize = (bytes?: number): string => {
  if (!bytes) return "Unknown";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

// Helper function to format download speed
const formatSpeed = (bytesPerSecond?: number): string => {
  if (!bytesPerSecond) return "";
  return `${formatFileSize(bytesPerSecond)}/s`;
};

// Helper function to calculate ETA
const calculateETA = (
  progress: number,
  speed?: number,
  totalSize?: number
): string => {
  if (!speed || !totalSize || progress >= 100) return "";

  const remainingBytes = totalSize * (1 - progress / 100);
  const secondsRemaining = remainingBytes / speed;

  if (secondsRemaining < 60) return `${Math.round(secondsRemaining)}s`;
  if (secondsRemaining < 3600) return `${Math.round(secondsRemaining / 60)}m`;
  if (secondsRemaining < 86400)
    return `${Math.round(secondsRemaining / 3600)}h`;
  return `${Math.round(secondsRemaining / 86400)}d`;
};

// Extended download type to match the UI expectations
interface ExtendedDownload extends Download {
  update_at: string;
  download_speed?: number; // bytes per second
  upload_speed?: number; // bytes per second
  total_size?: number; // total bytes
  downloaded_size?: number; // downloaded bytes
  eta?: string; // estimated time remaining
  user_id?: string; // user who initiated the download
}

export default function DownloadWidget() {
  const { connectionInfo } = useWebSocketContext();
  const { isAuthenticated, user } = useAuth();
  const { settings } = useSettings();
  const [downloads, setDownloads] = useState<ExtendedDownload[]>([]);
  const [sortKey, setSortKey] = useState<
    "progress" | "title" | "updated_at" | "speed"
  >("progress");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [showOwnOnly, setShowOwnOnly] = useState(false);
  const [showActions, setShowActions] = useState(false);

  // Fetch the currently stored downloads from backend for initial load
  const {
    data: apiDownloads,
    isLoading,
    error,
    refetch,
  } = useQuery<ExtendedDownload[]>({
    queryKey: ["downloads"],
    queryFn: async () => {
      console.log(
        "📥 Making /downloads API call - isAuthenticated:",
        isAuthenticated
      );
      try {
        const response = await backendApi.getDownloads();
        // Ensure we always return an array, even if response is null/undefined
        const downloads = Array.isArray(response) ? response : [];
        console.log(
          "✅ /downloads API call successful:",
          downloads.length,
          "downloads"
        );
        return downloads.map((d: Download) => ({
          ...d,
          update_at: d.update_at || new Date().toISOString(),
        }));
      } catch (err) {
        console.error("❌ /downloads API call failed:", err);
        throw err;
      }
    },
    refetchInterval: false, // Don't auto-refetch, let WebSocket handle updates
    staleTime: Infinity, // Keep data fresh indefinitely
    retry: 3,
    retryDelay: 1000,
    enabled: isAuthenticated, // Only run when authenticated
  });

  // Initialize with API data when it loads
  useEffect(() => {
    if (apiDownloads) {
      console.log("Initializing downloads from API:", apiDownloads.length);
      setDownloads(apiDownloads);
      setLastUpdate(new Date());
    } else {
      // Ensure downloads is always an array, even when no data
      setDownloads([]);
    }
  }, [apiDownloads]);

  // Handle download actions
  const handleDownloadAction = useCallback(
    async (
      downloadId: string,
      action: "pause" | "resume" | "cancel" | "retry"
    ) => {
      try {
        // TODO: Implement API calls for download actions
        console.log(`${action} download:`, downloadId);
        // await backendApi.downloadAction(downloadId, action);
      } catch (error) {
        console.error(`Failed to ${action} download:`, error);
      }
    },
    []
  );

  // Handle WebSocket download progress updates
  const handleDownloadProgressBatch = useCallback(
    (payload: DownloadProgressBatchPayload) => {
      console.log(
        "🚀 Received download progress batch:",
        payload.downloads.length,
        "downloads",
        payload
      );

      // Transform WebSocket data to match our format
      const transformedDownloads: ExtendedDownload[] = payload.downloads.map(
        (d) => ({
          id: d.id,
          title: d.title || "Unknown Title",
          torrent_title: d.torrent_title,
          source: d.source || "unknown",
          progress: d.progress || 0,
          time_left: d.time_left || "",
          status: d.status || "queued",
          update_at: new Date().toISOString(),
          // Optional fields that might not be in WebSocket data
          tmdb_id: d.tmdb_id,
          tvdb_id: d.tvdb_id,
          hash: d.hash,
          // Enhanced fields (would come from backend)
          download_speed: d.download_speed, // Mock data for demo
          upload_speed: d.upload_speed,
          downloaded_size: d.download_size,
          user_id: user?.id, // Mock - would come from backend
        })
      );

      // Merge with existing downloads instead of replacing
      setDownloads((prevDownloads) => {
        // Ensure prevDownloads is always an array
        const currentDownloads = Array.isArray(prevDownloads)
          ? prevDownloads
          : [];
        const existingMap = new Map(currentDownloads.map((d) => [d.id, d]));

        // Update or add new downloads
        transformedDownloads.forEach((newDownload) => {
          existingMap.set(newDownload.id, newDownload);
        });

        const mergedDownloads = Array.from(existingMap.values());
        console.log("✅ Merged downloads:", mergedDownloads.length);
        return mergedDownloads;
      });

      setLastUpdate(new Date());
    },
    [user?.id]
  );

  // Subscribe to download progress batch messages
  useWebSocketMessage("downloadProgressBatch", handleDownloadProgressBatch);

  // Debug: Log subscription status
  useEffect(() => {
    console.log(
      "📡 DownloadWidget: Subscribing to downloadProgressBatch messages"
    );
    return () => {
      console.log(
        "📡 DownloadWidget: Unsubscribing from downloadProgressBatch messages"
      );
    };
  }, []);

  // Handle connection events
  useWebSocketEvent(
    WebSocketEvent.CONNECTED,
    useCallback(() => {
      console.log(
        "WebSocket connected, server info:",
        connectionInfo.serverInfo
      );
      // Optionally refetch data when reconnected
      if (downloads.length === 0) {
        refetch();
      }
    }, [connectionInfo.serverInfo, downloads.length, refetch])
  );

  useWebSocketEvent(
    WebSocketEvent.DISCONNECTED,
    useCallback((reason) => {
      console.log("WebSocket disconnected:", reason);
    }, [])
  );

  useWebSocketEvent(
    WebSocketEvent.ERROR,
    useCallback((error) => {
      console.error("WebSocket error:", error);
    }, [])
  );

  // Refetch data if WebSocket is disconnected for too long
  useEffect(() => {
    console.log(
      "DownloadWidget: Connection state changed to:",
      connectionInfo.state
    );
    if (
      connectionInfo.state === WebSocketState.DISCONNECTED ||
      connectionInfo.state === WebSocketState.ERROR
    ) {
      console.log(
        "DownloadWidget: Detected disconnection, will refetch in 5 seconds",
        {
          state: connectionInfo.state,
          connectedAt: connectionInfo.connectedAt,
          disconnectedAt: connectionInfo.disconnectedAt,
          reconnectAttempts: connectionInfo.reconnectAttempts,
        }
      );
      const timeout = setTimeout(() => {
        console.log("WebSocket disconnected, refetching data");
        refetch();
      }, 5000); // Wait 5 seconds before refetching

      return () => clearTimeout(timeout);
    }
  }, [connectionInfo.state, refetch]);

  // Memoized sorted downloads with stable dependencies
  const sortedDownloads = useMemo(() => {
    const safeDate = (d: string) =>
      d && !isNaN(new Date(d).getTime()) ? new Date(d).getTime() : 0;

    return [...downloads].sort((a, b) => {
      const aWarning = (a.status ?? "").toLowerCase() === "warning";
      const bWarning = (b.status ?? "").toLowerCase() === "warning";

      if (aWarning && !bWarning) return 1;
      if (!aWarning && bWarning) return -1;

      switch (sortKey) {
        case "progress":
          return (b.progress ?? 0) - (a.progress ?? 0);
        case "title":
          return (a.title ?? "").localeCompare(b.title ?? "");
        case "updated_at":
          return safeDate(b.update_at || "") - safeDate(a.update_at || "");
        case "speed":
          return (b.download_speed ?? 0) - (a.download_speed ?? 0);
        default:
          return 0;
      }
    });
  }, [downloads, sortKey]);

  // Check if user can see all downloads or only their own
  const canSeeAllDownloads = useMemo(() => {
    const downloadVisibility = settings?.download_visibility || "all";
    return downloadVisibility === "all" || user?.is_admin;
  }, [settings?.download_visibility, user?.is_admin]);

  // Memoized filtered downloads
  const filteredDownloads = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    let filtered = sortedDownloads.filter((d) =>
      `${d.title ?? ""} ${d.torrent_title ?? ""}`.toLowerCase().includes(query)
    );

    // Filter by status
    if (selectedStatus !== "all") {
      filtered = filtered.filter(
        (d) => (d.status ?? "").toLowerCase() === selectedStatus.toLowerCase()
      );
    }

    // Filter by ownership if needed
    if (!canSeeAllDownloads || showOwnOnly) {
      filtered = filtered.filter((d) => d.user_id === user?.id);
    }

    return filtered;
  }, [
    sortedDownloads,
    searchQuery,
    selectedStatus,
    canSeeAllDownloads,
    showOwnOnly,
    user?.id,
  ]);

  const isTimeInfinite = (timeLeft?: string) => {
    if (!timeLeft) return false;
    const match = timeLeft.match(/(\d+)h/);
    const hours = match ? parseInt(match[1], 10) : 0;
    return hours >= 2400;
  };

  const renderTime = (timeLeft?: string) =>
    isTimeInfinite(timeLeft) ? (
      <InfinityIcon className="w-3 h-3 sm:w-4 sm:h-4 text-muted-foreground" />
    ) : (
      <span className="text-muted-foreground">({timeLeft})</span>
    );

  const getStatusConfig = (status: string) => {
    const statusLower = status.toLowerCase();
    switch (statusLower) {
      case "queued":
        return {
          icon: Clock,
          color: "text-muted-foreground",
          bg: "bg-muted",
          border: "border-border",
          label: "Queued",
        };
      case "downloading":
        return {
          icon: DownloadIcon,
          color: "text-blue-500",
          bg: "bg-blue-50 dark:bg-blue-950",
          border: "border-blue-200 dark:border-blue-800",
          label: "Downloading",
        };
      case "completed":
        return {
          icon: CheckCircle,
          color: "text-green-500",
          bg: "bg-green-50 dark:bg-green-950",
          border: "border-green-200 dark:border-green-800",
          label: "Completed",
        };
      case "warning":
        return {
          icon: AlertTriangle,
          color: "text-yellow-500",
          bg: "bg-yellow-50 dark:bg-yellow-950",
          border: "border-yellow-200 dark:border-yellow-800",
          label: "Warning",
        };
      case "paused":
        return {
          icon: Pause,
          color: "text-orange-500",
          bg: "bg-orange-50 dark:bg-orange-950",
          border: "border-orange-200 dark:border-orange-800",
          label: "Paused",
        };
      case "error":
        return {
          icon: X,
          color: "text-red-500",
          bg: "bg-red-50 dark:bg-red-950",
          border: "border-red-200 dark:border-red-800",
          label: "Error",
        };
      default:
        return {
          icon: Activity,
          color: "text-muted-foreground",
          bg: "bg-muted",
          border: "border-border",
          label: status,
        };
    }
  };

  const getUniqueStatuses = () => {
    const statuses = new Set(downloads?.map((d) => d.status ?? "queued") || []);
    return Array.from(statuses);
  };

  const getDownloadStats = () => {
    const total = downloads?.length || 0;
    const downloading =
      downloads?.filter((d) => (d.status ?? "").toLowerCase() === "downloading")
        .length || 0;
    const completed =
      downloads?.filter((d) => (d.status ?? "").toLowerCase() === "completed")
        .length || 0;
    const warnings =
      downloads?.filter((d) => (d.status ?? "").toLowerCase() === "warning")
        .length || 0;
    const paused =
      downloads?.filter((d) => (d.status ?? "").toLowerCase() === "paused")
        .length || 0;

    return { total, downloading, completed, warnings, paused };
  };

  // Show error state
  if (error) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 bg-destructive rounded-xl">
            <AlertTriangle className="w-4 h-4 sm:w-5 sm:h-5 text-destructive-foreground" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Error Loading Downloads
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Failed to load download data
            </p>
          </div>
        </div>
        <Button onClick={() => refetch()} variant="outline">
          Retry
        </Button>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4 sm:mb-6">
          <div className="p-2 bg-primary rounded-xl">
            <DownloadIcon className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Active Downloads
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Loading download queue...
            </p>
          </div>
        </div>

        <div className="space-y-3 sm:space-y-4">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="bg-muted/50 rounded-xl p-3 sm:p-4 animate-pulse"
            >
              <div className="flex justify-between items-center mb-2 sm:mb-3">
                <div className="flex items-center gap-2 sm:gap-3">
                  <Skeleton className="w-4 h-4 sm:w-5 sm:h-5 rounded" />
                  <Skeleton className="h-3 sm:h-4 w-32 sm:w-48" />
                </div>
                <Skeleton className="h-3 sm:h-4 w-12 sm:w-16" />
              </div>
              <Skeleton className="h-2 w-full mb-2" />
              <Skeleton className="h-3 w-16 sm:w-20" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  const stats = getDownloadStats();

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden">
      {/* Header */}
      <div className="p-4 sm:p-6 border-b border-border">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-3 sm:gap-0">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary rounded-xl">
              <DownloadIcon className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
            </div>
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-foreground">
                Active Downloads
              </h2>
              <p className="text-muted-foreground text-xs sm:text-sm">
                Monitor the download queue
                {connectionInfo.state === WebSocketState.CONNECTING && (
                  <span className="text-blue-500 ml-2">(Connecting...)</span>
                )}
                {connectionInfo.state === WebSocketState.RECONNECTING && (
                  <span className="text-yellow-500 ml-2">
                    (Reconnecting... attempt {connectionInfo.reconnectAttempts})
                  </span>
                )}
                {connectionInfo.state === WebSocketState.DISCONNECTED && (
                  <span className="text-red-500 ml-2">(Offline)</span>
                )}
                {connectionInfo.state === WebSocketState.ERROR && (
                  <span className="text-red-500 ml-2">(Connection Error)</span>
                )}
                {connectionInfo.state === WebSocketState.CONNECTED &&
                  connectionInfo.latency && (
                    <span className="text-green-500 ml-2">
                      (Online - {connectionInfo.latency}ms)
                    </span>
                  )}
              </p>
            </div>
          </div>

          {/* Controls */}
          <div className="flex gap-2 w-full sm:w-auto">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="sm"
                  variant="outline"
                  className="flex-1 sm:flex-none text-xs sm:text-sm"
                >
                  <Filter className="w-3 h-3 sm:w-4 sm:h-4 mr-1 sm:mr-2" />
                  Filter
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuItem
                  onClick={() => setSelectedStatus("all")}
                  className="text-sm"
                >
                  All Status
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                {getUniqueStatuses().map((status) => (
                  <DropdownMenuItem
                    key={status}
                    onClick={() => setSelectedStatus(status)}
                    className="text-sm"
                  >
                    {getStatusConfig(status).label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="sm"
                  variant="outline"
                  className="flex-1 sm:flex-none text-xs sm:text-sm"
                >
                  {sortKey === "progress" && (
                    <TrendingDown className="w-3 h-3 sm:w-4 sm:h-4 mr-1 sm:mr-2" />
                  )}
                  {sortKey === "title" && (
                    <ArrowUp className="w-3 h-3 sm:w-4 sm:h-4 mr-1 sm:mr-2" />
                  )}
                  {sortKey === "updated_at" && (
                    <Clock className="w-3 h-3 sm:w-4 sm:h-4 mr-1 sm:mr-2" />
                  )}
                  {sortKey === "speed" && (
                    <Zap className="w-3 h-3 sm:w-4 sm:h-4 mr-1 sm:mr-2" />
                  )}
                  Sort
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuItem
                  onClick={() => setSortKey("progress")}
                  className="text-sm"
                >
                  <TrendingDown className="w-4 h-4 mr-2" />
                  Progress
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => setSortKey("title")}
                  className="text-sm"
                >
                  <ArrowUp className="w-4 h-4 mr-2" />
                  Title
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => setSortKey("updated_at")}
                  className="text-sm"
                >
                  <Clock className="w-4 h-4 mr-2" />
                  Last Updated
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => setSortKey("speed")}
                  className="text-sm"
                >
                  <Zap className="w-4 h-4 mr-2" />
                  Download Speed
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Privacy Controls */}
        {canSeeAllDownloads && (
          <div className="flex items-center gap-4 mb-4 p-3 bg-muted/30 rounded-lg border">
            <div className="flex items-center space-x-2">
              <Switch
                id="own-downloads"
                checked={showOwnOnly}
                onCheckedChange={setShowOwnOnly}
              />
              <Label htmlFor="own-downloads" className="text-sm">
                <User className="w-4 h-4 inline mr-1" />
                Show only my downloads
              </Label>
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="show-actions"
                checked={showActions}
                onCheckedChange={setShowActions}
              />
              <Label htmlFor="show-actions" className="text-sm">
                <Settings className="w-4 h-4 inline mr-1" />
                Show actions
              </Label>
            </div>
          </div>
        )}

        {/* Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-5 gap-2 sm:gap-3 mb-4">
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <Activity className="w-3 h-3 sm:w-4 sm:h-4 text-primary" />
              <span className="text-muted-foreground text-xs sm:text-sm">
                Total
              </span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.total}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <DownloadIcon className="w-3 h-3 sm:w-4 sm:h-4 text-blue-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">
                Active
              </span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.downloading}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <CheckCircle className="w-3 h-3 sm:w-4 sm:h-4 text-green-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">
                Done
              </span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.completed}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <AlertTriangle className="w-3 h-3 sm:w-4 sm:h-4 text-yellow-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">
                Issues
              </span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.warnings}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <Pause className="w-3 h-3 sm:w-4 sm:h-4 text-orange-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">
                Paused
              </span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.paused}
            </div>
          </div>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-3 w-3 sm:h-4 sm:w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search downloads..."
            className="pl-8 sm:pl-10 text-sm sm:text-base h-9 sm:h-10"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Downloads List */}
      <div className="p-4 sm:p-6">
        {filteredDownloads.length === 0 ? (
          <div className="text-center py-8 sm:py-12">
            <div className="p-3 sm:p-4 bg-muted rounded-full w-fit mx-auto mb-3 sm:mb-4">
              <DownloadCloud className="w-6 h-6 sm:w-8 sm:h-8 text-muted-foreground" />
            </div>
            <h3 className="text-base sm:text-lg font-semibold text-foreground mb-1 sm:mb-2">
              No downloads found
            </h3>
            <p className="text-sm sm:text-base text-muted-foreground">
              {searchQuery
                ? "Try adjusting your search terms"
                : "Your download queue is empty"}
            </p>
          </div>
        ) : (
          <TooltipProvider>
            <div
              className="space-y-2 sm:space-y-3"
              role="list"
              aria-label="Download queue"
            >
              {filteredDownloads.map((d) => {
                const progress = Number.isFinite(d.progress) ? d.progress : 0;
                const status = d.status ?? "Queued";
                const statusConfig = getStatusConfig(status);
                const updatedAt = d.update_at
                  ? formatDistanceToNow(new Date(d.update_at), {
                      addSuffix: true,
                    })
                  : "Recently";

                const isActive = ["downloading", "queued"].includes(
                  status.toLowerCase()
                );
                const isCompleted = status.toLowerCase() === "completed";
                const hasWarning = status.toLowerCase() === "warning";

                return (
                  <div
                    key={d.id}
                    role="listitem"
                    aria-label={`Download: ${d.title || "Unknown Title"}`}
                    tabIndex={0}
                    className={`relative bg-card rounded-xl p-3 sm:p-4 border transition-all duration-300 hover:bg-muted/50 group focus:ring-2 focus:ring-primary focus:outline-none ${
                      hasWarning
                        ? "border-yellow-200 dark:border-yellow-800 bg-yellow-50/50 dark:bg-yellow-950/20"
                        : isCompleted
                        ? "border-green-200 dark:border-green-800 bg-green-50/50 dark:bg-green-950/20"
                        : isActive
                        ? "border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-950/20"
                        : "border-border"
                    }`}
                  >
                    {/* Header */}
                    <div className="flex items-start sm:items-center justify-between mb-2 sm:mb-3 gap-2">
                      <div className="flex items-start sm:items-center gap-2 sm:gap-3 flex-1 min-w-0">
                        {/* Media Type Icon */}
                        <div
                          className={`p-1.5 sm:p-2 rounded-lg flex-shrink-0 ${
                            d.source?.includes("movie") ||
                            d.source?.includes("radarr")
                              ? "bg-orange-500/20 text-orange-400"
                              : "bg-blue-500/20 text-blue-400"
                          }`}
                        >
                          {d.source?.includes("movie") ||
                          d.source?.includes("radarr") ? (
                            <Film className="w-3 h-3 sm:w-4 sm:h-4" />
                          ) : (
                            <Tv className="w-3 h-3 sm:w-4 sm:h-4" />
                          )}
                        </div>

                        {/* Title */}
                        <div className="flex-1 min-w-0">
                          <h4 className="font-semibold text-foreground text-xs sm:text-sm leading-tight mb-0.5 sm:mb-1 line-clamp-2 sm:truncate group-hover:text-primary transition-colors">
                            {d.title || "Unknown Title"}
                          </h4>
                          <p
                            className="text-xs text-muted-foreground truncate"
                            title={d.torrent_title || "No torrent info"}
                          >
                            {d.torrent_title || "No torrent info"}
                          </p>
                        </div>
                      </div>

                      {/* Progress & Speed */}
                      <div className="flex flex-col items-end gap-1 flex-shrink-0">
                        <div className="text-xs sm:text-sm font-bold text-foreground">
                          {Math.round(progress)}%
                        </div>
                        {d.download_speed && d.download_speed > 0 && (
                          <div className="text-xs text-green-600 font-medium">
                            <Zap className="w-3 h-3 inline mr-1" />
                            {formatSpeed(d.download_speed)}
                          </div>
                        )}
                        {d.time_left && (
                          <div className="text-xs text-muted-foreground">
                            {renderTime(d.time_left)}
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Progress Bar */}
                    <div className="mb-2 sm:mb-3">
                      <Progress
                        value={progress}
                        className={`h-1.5 sm:h-2 overflow-hidden ${
                          hasWarning
                            ? "[&>div]:bg-yellow-500"
                            : isCompleted
                            ? "[&>div]:bg-green-500"
                            : "[&>div]:bg-blue-500"
                        }`}
                      />
                    </div>

                    {/* File Size & Downloaded */}
                    {(d.total_size || d.downloaded_size) && (
                      <div className="flex items-center gap-4 mb-2 text-xs text-muted-foreground">
                        {d.total_size && (
                          <div className="flex items-center gap-1">
                            <HardDrive className="w-3 h-3" />
                            <span>Size: {formatFileSize(d.total_size)}</span>
                          </div>
                        )}
                        {d.downloaded_size && (
                          <div className="flex items-center gap-1">
                            <DownloadIcon className="w-3 h-3" />
                            <span>
                              Downloaded: {formatFileSize(d.downloaded_size)}
                            </span>
                          </div>
                        )}
                      </div>
                    )}

                    {/* Status & Actions */}
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                      <div className="flex items-center gap-2">
                        <div
                          className={`flex items-center gap-1 sm:gap-1.5 px-2 py-1 rounded-md text-xs font-medium ${statusConfig.bg} ${statusConfig.border} border`}
                        >
                          <statusConfig.icon
                            className={`w-2.5 h-2.5 sm:w-3 sm:h-3 ${statusConfig.color}`}
                          />
                          <span className={statusConfig.color}>
                            {statusConfig.label}
                          </span>
                        </div>

                        {/* Download Actions */}
                        {showActions &&
                          (canSeeAllDownloads || d.user_id === user?.id) && (
                            <div className="flex items-center gap-1">
                              {status.toLowerCase() === "downloading" && (
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      size="sm"
                                      variant="ghost"
                                      className="h-6 w-6 p-0"
                                      onClick={() =>
                                        handleDownloadAction(d.id, "pause")
                                      }
                                    >
                                      <Pause className="w-3 h-3" />
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Pause download</p>
                                  </TooltipContent>
                                </Tooltip>
                              )}
                              {status.toLowerCase() === "paused" && (
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      size="sm"
                                      variant="ghost"
                                      className="h-6 w-6 p-0"
                                      onClick={() =>
                                        handleDownloadAction(d.id, "resume")
                                      }
                                    >
                                      <Play className="w-3 h-3" />
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Resume download</p>
                                  </TooltipContent>
                                </Tooltip>
                              )}
                              {["error", "warning"].includes(
                                status.toLowerCase()
                              ) && (
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      size="sm"
                                      variant="ghost"
                                      className="h-6 w-6 p-0"
                                      onClick={() =>
                                        handleDownloadAction(d.id, "retry")
                                      }
                                    >
                                      <RotateCcw className="w-3 h-3" />
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Retry download</p>
                                  </TooltipContent>
                                </Tooltip>
                              )}
                              {!isCompleted && (
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      size="sm"
                                      variant="ghost"
                                      className="h-6 w-6 p-0 text-red-500 hover:text-red-600"
                                      onClick={() =>
                                        handleDownloadAction(d.id, "cancel")
                                      }
                                    >
                                      <X className="w-3 h-3" />
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Cancel download</p>
                                  </TooltipContent>
                                </Tooltip>
                              )}
                            </div>
                          )}
                      </div>

                      <div className="flex flex-col items-end gap-1">
                        <div className="text-xs text-muted-foreground">
                          <span className="sm:hidden">Updated </span>
                          {updatedAt}
                        </div>
                        {!canSeeAllDownloads && d.user_id && (
                          <div className="text-xs text-muted-foreground flex items-center gap-1">
                            <User className="w-3 h-3" />
                            <span>Your download</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </TooltipProvider>
        )}
      </div>
    </div>
  );
}
