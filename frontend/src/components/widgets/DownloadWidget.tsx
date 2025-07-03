import { useState, useMemo, useEffect, useCallback } from "react";
import { Progress } from "@/components/ui/progress";
import { useWebSocketContext } from "@/lib/WebSocketContext";
import {
  OpcodeDownloadProgressBatch,
  type Message,
  type DownloadProgressPayload,
  type Download,
} from "@/types";
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

// Extended download type to match the UI expectations
interface ExtendedDownload extends Download {
  update_at: string;
}

export default function DownloadWidget() {
  const { subscribe, isConnected } = useWebSocketContext();
  const { isAuthenticated } = useAuth();
  const [downloads, setDownloads] = useState<ExtendedDownload[]>([]);
  const [sortKey, setSortKey] = useState<"progress" | "title" | "updated_at">(
    "progress"
  );
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // Fetch the currently stored downloads from backend for initial load
  const {
    data: apiDownloads,
    isLoading,
    error,
    refetch,
  } = useQuery<ExtendedDownload[]>({
    queryKey: ["downloads"],
    queryFn: async () => {
      console.log("📥 Making /downloads API call - isAuthenticated:", isAuthenticated);
      try {
        const response = await backendApi.getDownloads();
        // Ensure we always return an array, even if response is null/undefined
        const downloads = Array.isArray(response) ? response : [];
        console.log("✅ /downloads API call successful:", downloads.length, "downloads");
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

  // Handle WebSocket updates with proper merging
  const handleWebSocketUpdate = useCallback((msg: Message) => {
    if (
      msg.op === OpcodeDownloadProgressBatch &&
      Array.isArray(msg.d.downloads)
    ) {
      console.log("Received WebSocket update:", msg.d.downloads.length, "downloads");
      
      // Transform WebSocket data to match our format
      const transformedDownloads: ExtendedDownload[] = msg.d.downloads.map(
        (d: DownloadProgressPayload) => ({
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
        })
      );

      // Merge with existing downloads instead of replacing
      setDownloads(prevDownloads => {
        // Ensure prevDownloads is always an array
        const currentDownloads = Array.isArray(prevDownloads) ? prevDownloads : [];
        const existingMap = new Map(currentDownloads.map(d => [d.id, d]));
        
        // Update or add new downloads
        transformedDownloads.forEach(newDownload => {
          existingMap.set(newDownload.id, newDownload);
        });
        
        const mergedDownloads = Array.from(existingMap.values());
        console.log("Merged downloads:", mergedDownloads.length);
        return mergedDownloads;
      });
      
      setLastUpdate(new Date());
    }
  }, []);

  // Subscribe to WebSocket updates
  useEffect(() => {
    if (!isConnected) {
      console.log("WebSocket not connected, skipping subscription");
      return;
    }

    console.log("Subscribing to WebSocket updates");
    const unsubscribe = subscribe(handleWebSocketUpdate);

    return () => {
      console.log("Unsubscribing from WebSocket updates");
      unsubscribe();
    };
  }, [subscribe, handleWebSocketUpdate, isConnected]);

  // Refetch data if WebSocket is disconnected for too long
  useEffect(() => {
    if (!isConnected) {
      const timeout = setTimeout(() => {
        console.log("WebSocket disconnected, refetching data");
        refetch();
      }, 5000); // Wait 5 seconds before refetching

      return () => clearTimeout(timeout);
    }
  }, [isConnected, refetch]);

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
        default:
          return 0;
      }
    });
  }, [downloads, sortKey]);

  // Memoized filtered downloads
  const filteredDownloads = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    let filtered = sortedDownloads.filter((d) =>
      `${d.title ?? ""} ${d.torrent_title ?? ""}`.toLowerCase().includes(query)
    );

    if (selectedStatus !== "all") {
      filtered = filtered.filter(
        (d) => (d.status ?? "").toLowerCase() === selectedStatus.toLowerCase()
      );
    }

    return filtered;
  }, [sortedDownloads, searchQuery, selectedStatus]);

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

    return { total, downloading, completed, warnings };
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
        <Button 
          onClick={() => refetch()} 
          variant="outline"
        >
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
                {!isConnected && (
                  <span className="text-yellow-500 ml-2">(Offline)</span>
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
              <DropdownMenuContent
                align="end"
                className="w-48"
              >
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
                  Sort
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="end"
                className="w-48"
              >
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
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 sm:gap-3 mb-4">
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <Activity className="w-3 h-3 sm:w-4 sm:h-4 text-primary" />
              <span className="text-muted-foreground text-xs sm:text-sm">Total</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.total}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <DownloadIcon className="w-3 h-3 sm:w-4 sm:h-4 text-blue-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">Active</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.downloading}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <CheckCircle className="w-3 h-3 sm:w-4 sm:h-4 text-green-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">Done</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.completed}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <AlertTriangle className="w-3 h-3 sm:w-4 sm:h-4 text-yellow-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">Issues</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {stats.warnings}
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
            <div className="space-y-2 sm:space-y-3">
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
                    className={`relative bg-card rounded-xl p-3 sm:p-4 border transition-all duration-300 hover:bg-muted/50 group ${
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
                          <p className="text-xs text-muted-foreground truncate">
                            {d.torrent_title || "No torrent info"}
                          </p>
                        </div>
                      </div>

                      {/* Progress & Time */}
                      <div className="flex flex-col items-end gap-1 flex-shrink-0">
                        <div className="text-xs sm:text-sm font-bold text-foreground">
                          {Math.round(progress)}%
                        </div>
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

                    {/* Status & Info */}
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
                      </div>

                      <div className="text-xs text-muted-foreground">
                        <span className="sm:hidden">Updated </span>
                        {updatedAt}
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
