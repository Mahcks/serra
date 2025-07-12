import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { discoverApi } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { MediaCard } from "@/components/ui/media-card";
import { 
  TrendingUp, 
  Calendar, 
  Film, 
  Tv, 
  Clock,
  ChevronRight,
  ChevronLeft,
  AlertTriangle
} from "lucide-react";
import { type TMDBMediaItem } from "@/types";

interface DiscoverySectionProps {
  title: string;
  icon: React.ReactNode;
  data: TMDBMediaItem[] | undefined;
  isLoading: boolean;
  error: unknown;
  onViewAll?: () => void;
}

function DiscoverySection({ title, icon, data, isLoading, error, onViewAll }: DiscoverySectionProps) {
  const navigate = useNavigate();
  const sectionId = title.replace(/\s+/g, '-').toLowerCase();

  const handleItemClick = (item: TMDBMediaItem) => {
    const mediaType = item.media_type || (item.title ? 'movie' : 'tv');
    navigate(`/requests/${item.id}/details?type=${mediaType}`);
  };

  const scrollLeft = () => {
    const container = document.getElementById(`scroll-${sectionId}`);
    if (container) {
      container.scrollBy({ left: -300, behavior: 'smooth' });
    }
  };

  const scrollRight = () => {
    const container = document.getElementById(`scroll-${sectionId}`);
    if (container) {
      container.scrollBy({ left: 300, behavior: 'smooth' });
    }
  };

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-xl font-semibold">
            {icon}
            {title}
          </h2>
        </div>
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <AlertTriangle className="w-8 h-8 text-destructive mx-auto mb-2" />
            <p className="text-sm text-muted-foreground">Failed to load content</p>
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-xl font-semibold">
            {icon}
            {title}
          </h2>
        </div>
        <div className="flex gap-4 overflow-hidden">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="flex-shrink-0 w-32 space-y-2">
              <Skeleton className="aspect-[2/3] w-full rounded-lg" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-3 w-2/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return null;
  }

  return (
    <div className="space-y-4 w-full max-w-[calc(100vw-4rem)] sm:max-w-[calc(100vw-8rem)] md:max-w-[calc(100vw-12rem)] lg:max-w-[calc(100vw-16rem)] xl:max-w-[calc(100vw-20rem)]">
      {/* Section Header - stays within container */}
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xl font-semibold">
          {icon}
          {title}
        </h2>
        <div className="flex items-center gap-2 flex-shrink-0">
          <Button
            variant="outline"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={scrollLeft}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={scrollRight}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          {onViewAll && (
            <Button variant="ghost" size="sm" onClick={onViewAll}>
              View All
              <ChevronRight className="w-4 h-4 ml-1" />
            </Button>
          )}
        </div>
      </div>

      {/* Media Carousel - simple grid approach */}
      <div className="relative overflow-hidden">
        <div 
          id={`scroll-${sectionId}`}
          className="flex gap-2 overflow-x-auto pb-2 scrollbar-hidden"
          style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
        >
          {data.slice(0, 20).map((item) => {
            const embyItem = {
              id: item.id.toString(),
              name: item.title || item.name || "Unknown Title",
              type: item.media_type === "tv" || item.first_air_date ? "Series" : "Movie",
              poster: item.poster_path
                ? `https://image.tmdb.org/t/p/w500${item.poster_path}`
                : "",
              vote_average: item.vote_average,
              release_date: item.release_date,
              first_air_date: item.first_air_date,
              media_type: item.media_type,
              overview: item.overview,
            };

            return (
              <div key={item.id} className="flex-none w-44">
                <MediaCard 
                  item={embyItem} 
                  size="md"
                  onClick={() => handleItemClick(item)}
                  className="w-full"
                />
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export function DiscoverySections() {
  const navigate = useNavigate();

  // Get trending content
  const { data: trending, isLoading: trendingLoading, error: trendingError } = useQuery({
    queryKey: ["trending"],
    queryFn: () => discoverApi.getTrending(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Get popular movies
  const { data: popularMovies, isLoading: popularMoviesLoading, error: popularMoviesError } = useQuery({
    queryKey: ["popularMovies"],
    queryFn: () => discoverApi.getPopularMovies(),
    staleTime: 15 * 60 * 1000, // 15 minutes
  });

  // Get popular TV shows
  const { data: popularTV, isLoading: popularTVLoading, error: popularTVError } = useQuery({
    queryKey: ["popularTV"],
    queryFn: () => discoverApi.getPopularTV(),
    staleTime: 15 * 60 * 1000, // 15 minutes
  });

  // Get upcoming movies
  const { data: upcomingMovies, isLoading: upcomingMoviesLoading, error: upcomingMoviesError } = useQuery({
    queryKey: ["upcomingMovies"],
    queryFn: () => discoverApi.getUpcomingMovies(),
    staleTime: 30 * 60 * 1000, // 30 minutes
  });

  // Get upcoming TV shows
  const { data: upcomingTV, isLoading: upcomingTVLoading, error: upcomingTVError } = useQuery({
    queryKey: ["upcomingTV"],
    queryFn: () => discoverApi.getUpcomingTV(),
    staleTime: 30 * 60 * 1000, // 30 minutes
  });

  const handleViewAllMovies = () => {
    navigate('/requests?tab=movies');
  };

  const handleViewAllSeries = () => {
    navigate('/requests?tab=series');
  };

  return (
    <div className="space-y-12">
      <DiscoverySection
        title="Trending Now"
        icon={<TrendingUp className="w-5 h-5 text-orange-500" />}
        data={trending?.results}
        isLoading={trendingLoading}
        error={trendingError}
      />
      
      <DiscoverySection
        title="Popular Movies"
        icon={<Film className="w-5 h-5 text-blue-500" />}
        data={popularMovies?.results}
        isLoading={popularMoviesLoading}
        error={popularMoviesError}
        onViewAll={handleViewAllMovies}
      />
      
      <DiscoverySection
        title="Popular TV Shows"
        icon={<Tv className="w-5 h-5 text-purple-500" />}
        data={popularTV?.results}
        isLoading={popularTVLoading}
        error={popularTVError}
        onViewAll={handleViewAllSeries}
      />
      
      <DiscoverySection
        title="Upcoming Movies"
        icon={<Calendar className="w-5 h-5 text-green-500" />}
        data={upcomingMovies?.results}
        isLoading={upcomingMoviesLoading}
        error={upcomingMoviesError}
        onViewAll={handleViewAllMovies}
      />
      
      <DiscoverySection
        title="Upcoming TV Shows"
        icon={<Clock className="w-5 h-5 text-indigo-500" />}
        data={upcomingTV?.results}
        isLoading={upcomingTVLoading}
        error={upcomingTVError}
        onViewAll={handleViewAllSeries}
      />
    </div>
  );
}