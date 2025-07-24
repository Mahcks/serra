import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { format, isToday, isTomorrow, isThisWeek, addDays } from "date-fns";
import { Film, Tv, CalendarDays, ChevronRight } from "lucide-react";
import type { CalendarItem } from "@/types";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";

export interface UpcomingItem {
  id: number;
  title: string;
  source: string; // "radarr" or "sonarr"
  releaseDate: string; // RFC3339
}

export default function CalendarWidget() {
  const [dayRange, setDayRange] = useState<7 | 30>(7);
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const {
    data: calendarData,
    isLoading,
    error,
  } = useQuery<CalendarItem[]>({
    queryKey: ["calendar"],
    queryFn: async () => {
      console.log(
        "📅 Making /calendar/upcoming API call - isAuthenticated:",
        isAuthenticated
      );
      const response = await api.get("/calendar/upcoming");
      console.log(
        "✅ /calendar/upcoming API call successful:",
        response.data.length,
        "items"
      );
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Handle loading state
  if (isLoading) {
    return (
      <div className="bg-card border border-border rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <h2 className="text-lg sm:text-xl font-bold text-foreground">
            Upcoming Digital Releases
          </h2>
        </div>
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-muted/50 rounded-md p-2 animate-pulse">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-muted rounded-full" />
                <div className="flex-1">
                  <div className="h-3 bg-muted rounded w-3/4 mb-1" />
                  <div className="h-2 bg-muted rounded w-1/2" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Handle error state
  if (error) {
    return (
      <div className="bg-card border border-border rounded-lg p-4">
        <div className="flex items-center gap-2 mb-2">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <h2 className="text-lg sm:text-xl font-bold text-foreground">
            Upcoming Digital Releases
          </h2>
        </div>
        <p className="text-destructive text-xs">Failed to load calendar data</p>
      </div>
    );
  }

  // Handle empty data state
  if (!calendarData || calendarData.length === 0) {
    return (
      <div className="bg-card border border-border rounded-lg p-4">
        <div className="flex items-center gap-2 mb-2">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <h2 className="text-lg sm:text-xl font-bold text-foreground">
            Upcoming Digital Releases
          </h2>
        </div>
        <p className="text-muted-foreground text-xs">
          No upcoming releases found
        </p>
      </div>
    );
  }

  // Get upcoming events for the selected day range
  const upcomingEvents = calendarData
    .filter((event) => {
      const eventDate = new Date(event.releaseDate);
      const now = new Date();
      const rangeEnd = addDays(now, dayRange);
      return eventDate >= now && eventDate <= rangeEnd;
    })
    .sort(
      (a, b) =>
        new Date(a.releaseDate).getTime() - new Date(b.releaseDate).getTime()
    )
    .slice(0, dayRange === 7 ? 8 : 12);

  const getDateLabel = (date: Date) => {
    if (isToday(date)) return "Today";
    if (isTomorrow(date)) return "Tomorrow";
    if (isThisWeek(date)) return format(date, "EEEE");
    return format(date, "MMM d");
  };

  const totalMovies = calendarData.filter((e) => e.source === "radarr").length;
  const totalShows = calendarData.filter((e) => e.source === "sonarr").length;

  // Handle navigation to media details page
  const handleItemClick = (event: CalendarItem) => {
    if (!event.tmdb_id) {
      console.warn("No TMDB ID available for navigation");
      return;
    }

    const mediaType = event.source === "radarr" ? "movie" : "tv";
    navigate(`/requests/${mediaType}/${event.tmdb_id}/details`);
  };

  return (
    <div className="bg-card border border-border rounded-lg p-4">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
          </div>
          <h2 className="text-lg sm:text-xl font-bold text-foreground">
            Upcoming Digital Releases
          </h2>
        </div>
        <div className="flex items-center gap-2">
          {/* Day Range Toggle */}
          <div className="flex bg-muted rounded-md p-0.5">
            <button
              onClick={() => setDayRange(7)}
              className={`px-2 py-1 text-xs font-medium rounded transition-all ${
                dayRange === 7
                  ? "bg-primary text-primary-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              7d
            </button>
            <button
              onClick={() => setDayRange(30)}
              className={`px-2 py-1 text-xs font-medium rounded transition-all ${
                dayRange === 30
                  ? "bg-primary text-primary-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              30d
            </button>
          </div>
          {/* Stats */}
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Film className="w-3 h-3 text-orange-500" />
            <span>{totalMovies}</span>
            <Tv className="w-3 h-3 text-blue-500 ml-1" />
            <span>{totalShows}</span>
          </div>
        </div>
      </div>

      {/* Upcoming Events */}
      {upcomingEvents.length === 0 ? (
        <div className="text-center py-6">
          <div className="w-8 h-8 bg-muted rounded-full mx-auto mb-2 flex items-center justify-center">
            <div className="p-2 bg-primary rounded-xl">
              <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
            </div>
          </div>
          <p className="text-sm text-muted-foreground">
            No upcoming releases in next {dayRange} days
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {upcomingEvents.map((event, index) => {
            const eventDate = new Date(event.releaseDate);
            const isToday =
              index === 0 &&
              new Date().toDateString() === eventDate.toDateString();

            return (
              <TooltipProvider key={`${event.source}-${event.title}`}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div
                      className={`flex items-center gap-2 p-2 rounded-md transition-all cursor-pointer group ${
                        isToday
                          ? "bg-primary/10 border border-primary/20"
                          : "hover:bg-muted/50"
                      }`}
                      onClick={() => handleItemClick(event)}
                    >
                      <div
                        className={`w-2 h-2 rounded-full flex-shrink-0 ${
                          event.source === "radarr"
                            ? "bg-orange-500"
                            : "bg-blue-500"
                        }`}
                      />
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-foreground truncate group-hover:text-primary transition-colors">
                          {event.title}
                        </p>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span>{getDateLabel(eventDate)}</span>
                          <span>•</span>
                          <span>{format(eventDate, "h:mm a")}</span>
                          <span
                            className={`px-1.5 py-0.5 rounded text-xs font-medium ${
                              event.source === "radarr"
                                ? "bg-orange-500/20 text-orange-600"
                                : "bg-blue-500/20 text-blue-600"
                            }`}
                          >
                            {event.source === "radarr" ? "Movie" : "TV"}
                          </span>
                        </div>
                      </div>
                      <ChevronRight className="w-3 h-3 text-muted-foreground group-hover:text-foreground transition-colors flex-shrink-0" />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent
                    side="left"
                    className="bg-popover border-border"
                  >
                    <p className="font-medium">{event.title}</p>
                    <p className="text-muted-foreground text-sm">
                      {event.source === "radarr" ? "Movie" : "TV Show"}{" "}
                      releasing {format(eventDate, "PPP 'at' h:mm a")}
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            );
          })}
        </div>
      )}
    </div>
  );
}
