import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { discoverApi, requestsApi } from "@/lib/api";
import { toast } from "sonner";
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
import { type TMDBMediaItem, type TMDBFullMediaItem, type CreateRequestRequest } from "@/types";
import { useAuth } from "@/lib/auth";
import { useState, useCallback } from "react";

interface DiscoverySectionProps {
  title: string;
  icon: React.ReactNode;
  data: TMDBFullMediaItem[] | undefined;
  isLoading: boolean;
  error: unknown;
  onViewAll?: () => void;
  onRequest?: (item: TMDBMediaItem) => void;
}

function DiscoverySection({ title, icon, data, isLoading, error, onViewAll, onRequest }: DiscoverySectionProps) {
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
            // Extract the TMDBMediaItem from TMDBFullMediaItem
            const mediaItem = 'TMDBMediaItem' in item ? item.TMDBMediaItem : item;
            
            const embyItem = {
              id: mediaItem.id.toString(),
              name: mediaItem.title || mediaItem.name || "Unknown Title",
              type: mediaItem.media_type === "tv" || mediaItem.first_air_date ? "Series" : "Movie",
              poster: mediaItem.poster_path
                ? `https://image.tmdb.org/t/p/w500${mediaItem.poster_path}`
                : "",
              vote_average: mediaItem.vote_average,
              release_date: mediaItem.release_date,
              first_air_date: mediaItem.first_air_date,
              media_type: mediaItem.media_type,
              overview: mediaItem.overview,
            };

            // Use the enriched status if available
            const status = 'in_library' in item && 'requested' in item ? {
              isInLibrary: item.in_library,
              isRequested: item.requested,
            } : undefined;

            return (
              <div key={mediaItem.id} className="flex-none w-44">
                <MediaCard 
                  item={embyItem} 
                  size="md"
                  onClick={() => handleItemClick(mediaItem)}
                  onRequest={onRequest ? () => onRequest(mediaItem) : undefined}
                  status={status}
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
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [currentRequestItem, setCurrentRequestItem] = useState<TMDBMediaItem | null>(null);

  // Create request mutation
  const createRequestMutation = useMutation({
    mutationFn: (data: CreateRequestRequest) => {
      return requestsApi.createRequest(data);
    },
    onSuccess: (newRequest) => {
      const displayTitle = newRequest.title || 
                          currentRequestItem?.title || 
                          currentRequestItem?.name || 
                          "the requested content";
      
      if (newRequest.status === 'approved') {
        toast.success(`🎉 Request Approved!`, {
          description: `"${displayTitle}" was automatically approved and will be downloaded soon.`,
          duration: 5000,
        });
      } else if (newRequest.status === 'fulfilled') {
        toast.success(`✅ Request Fulfilled!`, {
          description: `"${displayTitle}" is already available in your library.`,
          duration: 4000,
        });
      } else {
        toast.success(`📝 Request Submitted!`, {
          description: `Your request for "${displayTitle}" has been submitted for review.`,
          duration: 4000,
        });
      }
      
      // Invalidate all discovery queries to update request status in the UI
      queryClient.invalidateQueries({ queryKey: ["trending"] });
      queryClient.invalidateQueries({ queryKey: ["popularMovies"] });
      queryClient.invalidateQueries({ queryKey: ["popularTV"] });
      queryClient.invalidateQueries({ queryKey: ["upcomingMovies"] });
      queryClient.invalidateQueries({ queryKey: ["upcomingTV"] });
      
      setCurrentRequestItem(null);
    },
    onError: (error: any) => {
      const statusCode = error.response?.status;
      const errorData = error.response?.data;
      const errorMessage = errorData?.error?.message || errorData?.message || error.message;
      
      if (statusCode === 400 && errorMessage?.toLowerCase().includes('already requested')) {
        toast.error(`🔄 Already Requested`, {
          description: `You've already requested this content.`,
          duration: 4000,
        });
      } else {
        toast.error(`❌ Request Failed`, {
          description: errorMessage || "Failed to create request. Please try again.",
          duration: 4000,
        });
      }
      
      setCurrentRequestItem(null);
    },
  });

  const handleRequest = useCallback((item: TMDBMediaItem) => {
    if (!user) return;
    
    const mediaType = item.media_type || (item.first_air_date ? 'tv' : 'movie');
    const title = item.title || item.name || 'Unknown Title';
    const posterUrl = item.poster_path ? `https://image.tmdb.org/t/p/w500${item.poster_path}` : undefined;

    const requestData = {
      media_type: mediaType,
      tmdb_id: item.id,
      title: title,
      poster_url: posterUrl,
    };
    
    setCurrentRequestItem(item);
    createRequestMutation.mutate(requestData);
  }, [createRequestMutation, user]);

  // Get trending content (enriched with library/request status when user is logged in)
  const { data: trendingResponse, isLoading: trendingLoading, error: trendingError } = useQuery({
    queryKey: ["trending"],
    queryFn: () => discoverApi.getTrending(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Get popular movies (enriched with library/request status when user is logged in)
  const { data: popularMoviesResponse, isLoading: popularMoviesLoading, error: popularMoviesError } = useQuery({
    queryKey: ["popularMovies"],
    queryFn: () => discoverApi.getPopularMovies(),
    staleTime: 15 * 60 * 1000, // 15 minutes
  });

  // Get popular TV shows (enriched with library/request status when user is logged in)
  const { data: popularTVResponse, isLoading: popularTVLoading, error: popularTVError } = useQuery({
    queryKey: ["popularTV"],
    queryFn: () => discoverApi.getPopularTV(),
    staleTime: 15 * 60 * 1000, // 15 minutes
  });

  // Get upcoming movies (enriched with library/request status when user is logged in)
  const { data: upcomingMoviesResponse, isLoading: upcomingMoviesLoading, error: upcomingMoviesError } = useQuery({
    queryKey: ["upcomingMovies"],
    queryFn: () => discoverApi.getUpcomingMovies(),
    staleTime: 30 * 60 * 1000, // 30 minutes
  });

  // Extract results from response and handle both enriched and basic formats
  const trending = trendingResponse?.results;
  const popularMovies = popularMoviesResponse?.results;
  const popularTV = popularTVResponse?.results;
  const upcomingMovies = upcomingMoviesResponse?.results;

  // Get upcoming TV shows (enriched with library/request status when user is logged in)
  const { data: upcomingTVResponse, isLoading: upcomingTVLoading, error: upcomingTVError } = useQuery({
    queryKey: ["upcomingTV"],
    queryFn: () => discoverApi.getUpcomingTV(),
    staleTime: 30 * 60 * 1000, // 30 minutes
  });

  // Extract results from response
  const upcomingTV = upcomingTVResponse?.results;

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
        data={trending}
        isLoading={trendingLoading}
        error={trendingError}
        onRequest={handleRequest}
      />
      
      <DiscoverySection
        title="Popular Movies"
        icon={<Film className="w-5 h-5 text-blue-500" />}
        data={popularMovies}
        isLoading={popularMoviesLoading}
        error={popularMoviesError}
        onViewAll={handleViewAllMovies}
        onRequest={handleRequest}
      />
      
      <DiscoverySection
        title="Popular TV Shows"
        icon={<Tv className="w-5 h-5 text-purple-500" />}
        data={popularTV}
        isLoading={popularTVLoading}
        error={popularTVError}
        onViewAll={handleViewAllSeries}
        onRequest={handleRequest}
      />
      
      <DiscoverySection
        title="Upcoming Movies"
        icon={<Calendar className="w-5 h-5 text-green-500" />}
        data={upcomingMovies}
        isLoading={upcomingMoviesLoading}
        error={upcomingMoviesError}
        onViewAll={handleViewAllMovies}
        onRequest={handleRequest}
      />
      
      <DiscoverySection
        title="Upcoming TV Shows"
        icon={<Clock className="w-5 h-5 text-indigo-500" />}
        data={upcomingTV}
        isLoading={upcomingTVLoading}
        error={upcomingTVError}
        onViewAll={handleViewAllSeries}
        onRequest={handleRequest}
      />
    </div>
  );
}