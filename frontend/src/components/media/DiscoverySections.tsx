import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { discoverApi, requestsApi } from "@/lib/api";
import { toast } from "sonner";
import { MediaCard } from "@/components/ui/media-card";
import MediaCarousel from "@/components/ui/media-carousel";
import { 
  TrendingUp, 
  Calendar, 
  Film, 
  Tv, 
  Clock,
} from "lucide-react";
import { type TMDBMediaItem, type TMDBFullMediaItem, type CreateRequestRequest } from "@/types";
import { useAuth } from "@/lib/auth";
import { useState, useCallback } from "react";
import { handleApiError, ERROR_CODES, getErrorCode } from "@/utils/errorHandling";

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

  const handleItemClick = (item: TMDBFullMediaItem) => {
    // Extract the TMDBMediaItem from TMDBFullMediaItem
    const mediaItem = 'TMDBMediaItem' in item ? item.TMDBMediaItem : item;
    const mediaType = mediaItem.media_type || (mediaItem.title ? 'movie' : 'tv');
    navigate(`/requests/${mediaType}/${mediaItem.id}/details`);
  };

  const renderMediaItem = (item: TMDBFullMediaItem) => {
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
      <MediaCard 
        item={embyItem} 
        size="md"
        onRequest={onRequest ? () => onRequest(mediaItem) : undefined}
        status={status}
        className="w-full"
      />
    );
  };

  return (
    <MediaCarousel
      title={title}
      icon={icon}
      data={data}
      isLoading={isLoading}
      error={error instanceof Error ? error : error ? new Error('Unknown error') : null}
      onViewAll={onViewAll}
      onItemClick={handleItemClick}
      renderItem={renderMediaItem}
      keyExtractor={(item) => {
        const mediaItem = 'TMDBMediaItem' in item ? item.TMDBMediaItem : item;
        return mediaItem.id;
      }}
      itemWidth="w-44"
      scrollAmount={300}
      maxItems={20}
    />
  );
}

interface DiscoverySectionsProps {
  onRequest?: (item: TMDBMediaItem) => void;
}

export function DiscoverySections({ onRequest: externalOnRequest = undefined }: DiscoverySectionsProps = {}) {
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
      console.error("Request creation failed:", error);
      
      const { message } = handleApiError(error);
      const errorCode = getErrorCode(error);
      
      if (errorCode === ERROR_CODES.DUPLICATE_REQUEST) {
        toast.error(`🔄 Already Requested`, {
          description: `You've already requested this content.`,
          duration: 4000,
        });
      } else {
        const title = errorCode ? `❌ Request Failed (${errorCode})` : `❌ Request Failed`;
        toast.error(title, {
          description: message,
          duration: 4000,
        });
      }
      
      setCurrentRequestItem(null);
    },
  });

  const handleRequest = useCallback((item: TMDBMediaItem) => {
    if (!user) return;
    
    // Use external request handler if provided (from RequestPage)
    if (externalOnRequest) {
      externalOnRequest(item);
      return;
    }
    
    // Otherwise use internal request handling (for home page)
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
  }, [createRequestMutation, user, externalOnRequest]);

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

  const handleViewAllUpcomingMovies = () => {
    const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
    const params = new URLSearchParams({
      tab: 'movies',
      'release_date.gte': today,
      sort_by: 'popularity.desc'
    });
    navigate(`/requests?${params.toString()}`);
  };

  const handleViewAllUpcomingTV = () => {
    const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
    const params = new URLSearchParams({
      tab: 'series',
      'first_air_date.gte': today,
      sort_by: 'popularity.desc'
    });
    navigate(`/requests?${params.toString()}`);
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
        onViewAll={handleViewAllUpcomingMovies}
        onRequest={handleRequest}
      />
      
      <DiscoverySection
        title="Upcoming TV Shows"
        icon={<Clock className="w-5 h-5 text-indigo-500" />}
        data={upcomingTV}
        isLoading={upcomingTVLoading}
        error={upcomingTVError}
        onViewAll={handleViewAllUpcomingTV}
        onRequest={handleRequest}
      />
    </div>
  );
}