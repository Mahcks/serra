import { useState } from "react";
import { Calendar } from "@/components/ui/calendar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { format, isToday, isTomorrow, isThisWeek, addDays } from "date-fns";
import {
  Film,
  Tv,
  Sparkles,
  CalendarDays,
  Clock,
  TrendingUp,
  Star,
  Calendar as CalendarIcon,
  ChevronRight,
  Zap,
} from "lucide-react";
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
  const [selectedDate, setSelectedDate] = useState<Date | undefined>(
    new Date()
  );
  const [viewMode, setViewMode] = useState<"calendar" | "upcoming">("upcoming");
  const { isAuthenticated } = useAuth();

  const {
    data: calendarData,
    isLoading,
    error,
  } = useQuery<CalendarItem[]>({
    queryKey: ["calendar"],
    queryFn: async () => {
      console.log("📅 Making /calendar/upcoming API call - isAuthenticated:", isAuthenticated);
      const response = await api.get("/calendar/upcoming");
      console.log("✅ /calendar/upcoming API call successful:", response.data.length, "items");
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Handle loading state
  if (isLoading) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4 sm:mb-6">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Upcoming Releases
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">Loading your schedule...</p>
          </div>
        </div>

        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="bg-muted/50 rounded-xl p-3 sm:p-4 animate-pulse"
            >
              <div className="flex items-center gap-3">
                <div className="w-6 h-6 bg-muted rounded-full" />
                <div className="flex-1">
                  <div className="h-4 bg-muted rounded w-3/4 mb-2" />
                  <div className="h-3 bg-muted rounded w-1/2" />
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
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4 sm:mb-6">
          <div className="p-2 bg-destructive rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-destructive-foreground" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Upcoming Releases
            </h2>
            <p className="text-destructive text-xs sm:text-sm">Failed to load calendar data</p>
          </div>
        </div>
      </div>
    );
  }

  // Handle empty data state
  if (!calendarData || calendarData.length === 0) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4 sm:mb-6">
          <div className="p-2 bg-primary rounded-xl">
            <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Upcoming Releases
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">No upcoming releases found</p>
          </div>
        </div>
      </div>
    );
  }

  const eventsByDate = calendarData.reduce((acc, event) => {
    const dateStr = format(new Date(event.releaseDate), "yyyy-MM-dd");
    if (!acc[dateStr]) acc[dateStr] = [];
    acc[dateStr].push({
      id: Math.random(), // Generate a unique ID since CalendarItem doesn't have one
      title: event.title,
      source: event.source,
      releaseDate: event.releaseDate,
    });
    return acc;
  }, {} as Record<string, UpcomingItem[]>);

  const eventsForSelected = selectedDate
    ? eventsByDate[format(selectedDate, "yyyy-MM-dd")] || []
    : [];

  const datesWithSources = Object.keys(eventsByDate).reduce((acc, date) => {
    const sources = new Set(eventsByDate[date].map((e) => e.source));
    acc[date] = sources;
    return acc;
  }, {} as Record<string, Set<string>>);

  // Get upcoming events for the next 7 days
  const upcomingEvents = calendarData
    .filter((event) => {
      const eventDate = new Date(event.releaseDate);
      const now = new Date();
      const weekFromNow = addDays(now, 7);
      return eventDate >= now && eventDate <= weekFromNow;
    })
    .sort(
      (a, b) =>
        new Date(a.releaseDate).getTime() - new Date(b.releaseDate).getTime()
    )
    .slice(0, 5);

  const getDateLabel = (date: Date) => {
    if (isToday(date)) return "Today";
    if (isTomorrow(date)) return "Tomorrow";
    if (isThisWeek(date)) return format(date, "EEEE");
    return format(date, "MMM d");
  };

  const totalMovies = calendarData.filter((e) => e.source === "radarr").length;
  const totalShows = calendarData.filter((e) => e.source === "sonarr").length;

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden">
      {/* Header */}
      <div className="p-4 sm:p-6 border-b border-border">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary rounded-xl">
              <CalendarDays className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
            </div>
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-foreground">
                Upcoming Releases
              </h2>
              <p className="text-muted-foreground text-xs sm:text-sm">
                Your digital media schedule
              </p>
            </div>
          </div>

          {/* View Toggle */}
          <div className="flex bg-muted rounded-lg p-1 w-full sm:w-auto">
            <button
              onClick={() => setViewMode("upcoming")}
              className={`flex-1 sm:flex-initial px-3 py-2 sm:py-2 rounded-md text-sm font-medium transition-all min-h-[44px] sm:min-h-0 ${
                viewMode === "upcoming"
                  ? "bg-primary text-primary-foreground shadow-lg"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <Clock className="w-4 h-4 inline mr-1" />
              <span className="sm:inline">Upcoming</span>
            </button>
            <button
              onClick={() => setViewMode("calendar")}
              className={`flex-1 sm:flex-initial px-3 py-2 sm:py-2 rounded-md text-sm font-medium transition-all min-h-[44px] sm:min-h-0 ${
                viewMode === "calendar"
                  ? "bg-primary text-primary-foreground shadow-lg"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <CalendarIcon className="w-4 h-4 inline mr-1" />
              <span className="sm:inline">Calendar</span>
            </button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-2 sm:gap-3">
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <Film className="w-3 h-3 sm:w-4 sm:h-4 text-orange-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">Movies</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {totalMovies}
            </div>
          </div>
          <div className="bg-muted/50 rounded-lg p-2 sm:p-3 border border-border">
            <div className="flex items-center gap-1 sm:gap-2 mb-1">
              <Tv className="w-3 h-3 sm:w-4 sm:h-4 text-blue-500" />
              <span className="text-muted-foreground text-xs sm:text-sm">TV Shows</span>
            </div>
            <div className="text-base sm:text-lg font-bold text-foreground">
              {totalShows}
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4 sm:p-6">
        {viewMode === "calendar" ? (
          <div className="space-y-4 sm:space-y-6">
            <div className="flex justify-center overflow-x-auto">
              <div className="min-w-fit">
                <Calendar
                  mode="single"
                  selected={selectedDate}
                  onSelect={setSelectedDate}
                  className="border border-border rounded-xl bg-card mx-auto"
                  modifiers={{
                    radarrOnly: (date) => {
                      const day = format(date, "yyyy-MM-dd");
                      const sources = datesWithSources[day];
                      return sources?.has("radarr") && !sources?.has("sonarr");
                    },
                    sonarrOnly: (date) => {
                      const day = format(date, "yyyy-MM-dd");
                      const sources = datesWithSources[day];
                      return sources?.has("sonarr") && !sources?.has("radarr");
                    },
                    bothSources: (date) => {
                      const day = format(date, "yyyy-MM-dd");
                      const sources = datesWithSources[day];
                      return sources?.has("radarr") && sources?.has("sonarr");
                    },
                  }}
                  modifiersClassNames={{
                    radarrOnly:
                      "bg-orange-500 text-white rounded-full hover:bg-orange-600",
                    sonarrOnly:
                      "bg-blue-500 text-white rounded-full hover:bg-blue-600",
                    bothSources:
                      "bg-gradient-to-r from-orange-500 to-blue-500 text-white rounded-full",
                  }}
                />
              </div>
            </div>

            {/* Selected date events */}
            {selectedDate && (
              <div>
                <h3 className="text-lg font-semibold text-foreground mb-3 flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-2">
                  <div className="flex items-center gap-2">
                    <Star className="w-5 h-5 text-yellow-500" />
                    {getDateLabel(selectedDate)}
                  </div>
                  <span className="text-muted-foreground text-sm font-normal ml-7 sm:ml-0">
                    ({format(selectedDate, "MMM d, yyyy")})
                  </span>
                </h3>

                {eventsForSelected.length === 0 ? (
                  <div className="text-center py-6 sm:py-8">
                    <Sparkles className="w-10 h-10 sm:w-12 sm:h-12 mx-auto text-muted-foreground mb-3" />
                    <p className="text-muted-foreground text-sm sm:text-base">
                      No releases scheduled for this date
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {eventsForSelected.map((event) => (
                      <TooltipProvider key={`${event.source}-${event.id}`}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className="bg-card rounded-xl p-3 sm:p-4 border border-border hover:bg-muted/50 transition-all cursor-pointer group min-h-[60px] sm:min-h-0">
                              <div className="flex items-center gap-3">
                                <div
                                  className={`p-2 rounded-lg ${
                                    event.source === "radarr"
                                      ? "bg-orange-500/20 text-orange-500"
                                      : "bg-blue-500/20 text-blue-500"
                                  }`}
                                >
                                  {event.source === "radarr" ? (
                                    <Film className="w-4 h-4" />
                                  ) : (
                                    <Tv className="w-4 h-4" />
                                  )}
                                </div>
                                <div className="flex-1 min-w-0">
                                  <h4 className="font-semibold text-foreground group-hover:text-primary transition-colors text-sm sm:text-base truncate">
                                    {event.title}
                                  </h4>
                                  <p className="text-muted-foreground text-xs sm:text-sm">
                                    {event.source === "radarr"
                                      ? "Movie"
                                      : "TV Show"}{" "}
                                    ·{" "}
                                    {format(
                                      new Date(event.releaseDate),
                                      "h:mm a"
                                    )}
                                  </p>
                                </div>
                                <ChevronRight className="w-4 h-4 text-muted-foreground group-hover:text-foreground transition-colors flex-shrink-0" />
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent
                            side="left"
                            className="bg-popover border-border"
                          >
                            <p className="font-medium">{event.title}</p>
                            <p className="text-muted-foreground text-sm">
                              {event.source === "radarr" ? "Movie" : "TV Show"}{" "}
                              releasing{" "}
                              {format(
                                new Date(event.releaseDate),
                                "PPP 'at' h:mm a"
                              )}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {/* Quick stats for upcoming */}
            <div className="bg-primary/10 rounded-xl p-3 sm:p-4 border border-primary/20">
              <div className="flex items-center gap-2 mb-2">
                <TrendingUp className="w-4 h-4 sm:w-5 sm:h-5 text-primary" />
                <span className="text-primary font-semibold text-sm sm:text-base">
                  Next 7 Days
                </span>
              </div>
              <p className="text-foreground text-xs sm:text-sm">
                {upcomingEvents.length} releases coming up this week
              </p>
            </div>

            {upcomingEvents.length === 0 ? (
              <div className="text-center py-8 sm:py-12">
                <div className="p-3 sm:p-4 bg-muted rounded-full w-fit mx-auto mb-4">
                  <Sparkles className="w-6 h-6 sm:w-8 sm:h-8 text-muted-foreground" />
                </div>
                <h3 className="text-base sm:text-lg font-semibold text-foreground mb-2">
                  All caught up!
                </h3>
                <p className="text-muted-foreground text-sm sm:text-base">
                  No releases scheduled for the next week
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {upcomingEvents.map((event, index) => {
                  const eventDate = new Date(event.releaseDate);
                  const isNext = index === 0;

                  return (
                    <TooltipProvider key={`${event.source}-${event.title}`}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div
                            className={`bg-card rounded-xl p-3 sm:p-4 border transition-all cursor-pointer group relative overflow-hidden min-h-[80px] sm:min-h-0 ${
                              isNext
                                ? "border-primary bg-primary/5"
                                : "border-border hover:bg-muted/50"
                            }`}
                          >
                            {isNext && (
                              <div className="absolute top-2 right-2 sm:top-3 sm:right-3">
                                <div className="flex items-center gap-1 bg-primary text-primary-foreground px-2 py-1 rounded-full text-xs font-medium">
                                  <Zap className="w-3 h-3" />
                                  Next
                                </div>
                              </div>
                            )}

                            <div className="flex items-start gap-3 pr-16 sm:pr-0">
                              <div
                                className={`p-2 sm:p-3 rounded-xl ${
                                  event.source === "radarr"
                                    ? "bg-orange-500/20 text-orange-500"
                                    : "bg-blue-500/20 text-blue-500"
                                }`}
                              >
                                {event.source === "radarr" ? (
                                  <Film className="w-4 h-4 sm:w-5 sm:h-5" />
                                ) : (
                                  <Tv className="w-4 h-4 sm:w-5 sm:h-5" />
                                )}
                              </div>

                              <div className="flex-1 min-w-0">
                                <h4 className="font-semibold text-foreground group-hover:text-primary transition-colors mb-1 text-sm sm:text-base pr-2">
                                  {event.title}
                                </h4>
                                <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3 text-xs sm:text-sm text-muted-foreground">
                                  <span className="flex items-center gap-1">
                                    <CalendarIcon className="w-3 h-3" />
                                    {getDateLabel(eventDate)}
                                  </span>
                                  <span className="flex items-center gap-1">
                                    <Clock className="w-3 h-3" />
                                    {format(eventDate, "h:mm a")}
                                  </span>
                                  <span
                                    className={`px-2 py-0.5 rounded-full text-xs font-medium w-fit ${
                                      event.source === "radarr"
                                        ? "bg-orange-500/20 text-orange-400"
                                        : "bg-blue-500/20 text-blue-400"
                                    }`}
                                  >
                                    {event.source === "radarr" ? "Movie" : "TV"}
                                  </span>
                                </div>
                              </div>
                            </div>
                          </div>
                        </TooltipTrigger>
                      </Tooltip>
                    </TooltipProvider>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
