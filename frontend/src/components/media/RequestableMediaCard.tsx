import { memo, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { MediaCard } from "@/components/ui/media-card";
import { discoverApi } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { type TMDBMediaItem, type TMDBFullMediaItem, type EmbyMediaItem } from "@/types";

interface RequestableMediaCardProps {
  item: TMDBMediaItem | TMDBFullMediaItem;
  onRequest?: (item: TMDBMediaItem) => void;
  size?: "sm" | "md" | "lg";
  isRequestLoading?: boolean;
}

export const RequestableMediaCard = memo(function RequestableMediaCard({
  item,
  onRequest,
  size = "md",
  isRequestLoading = false,
}: RequestableMediaCardProps) {
  const navigate = useNavigate();
  const { user } = useAuth();
  
  // Extract the TMDBMediaItem from TMDBFullMediaItem if needed
  const mediaItem = 'TMDBMediaItem' in item ? item.TMDBMediaItem : item;
  
  // Check if we have enriched status data
  const hasEnrichedData = 'in_library' in item || 'requested' in item;
  
  // Use enriched status if available, otherwise query individual status
  const enrichedIsInLibrary = 'in_library' in item ? item.in_library : false;
  const enrichedIsRequested = 'requested' in item ? item.requested : false;
  
  // Get individual media status if enriched data is not available
  const mediaType = mediaItem.media_type || (mediaItem.first_air_date ? 'tv' : 'movie');
  const { data: mediaStatus } = useQuery({
    queryKey: ["mediaStatus", mediaItem.id, mediaType],
    queryFn: () => discoverApi.getMediaStatus(mediaItem.id, mediaType),
    enabled: !!user && !hasEnrichedData, // Only query if user is logged in and no enriched data
    staleTime: 30000, // Cache for 30 seconds
  });
  
  // Determine final status (prefer enriched data, fallback to individual query)
  const isInLibrary = hasEnrichedData ? enrichedIsInLibrary : (mediaStatus?.in_library || false);
  const isRequested = hasEnrichedData ? enrichedIsRequested : (mediaStatus?.requested || false);

  const embyItem = useMemo(
    (): EmbyMediaItem & {
      vote_average?: number;
      release_date?: string;
      first_air_date?: string;
      media_type?: string;
      overview?: string;
      character?: string;
      job?: string;
      department?: string;
    } => ({
      id: mediaItem.id.toString(),
      name: mediaItem.title || mediaItem.name || "Unknown Title",
      type:
        mediaItem.media_type === "tv" || mediaItem.first_air_date ? "Series" : "Movie",
      poster: mediaItem.poster_path
        ? `https://image.tmdb.org/t/p/w500${mediaItem.poster_path}`
        : "",
      vote_average: mediaItem.vote_average,
      release_date: mediaItem.release_date,
      first_air_date: mediaItem.first_air_date,
      media_type: mediaItem.media_type,
      overview: mediaItem.overview,
      character: (mediaItem as any).character,
      job: (mediaItem as any).job,
      department: (mediaItem as any).department,
    }),
    [mediaItem]
  );

  const handleCardClick = useCallback(() => {
    const mediaType = mediaItem.media_type || (mediaItem.first_air_date ? 'tv' : 'movie');
    navigate(`/requests/${mediaType}/${mediaItem.id}/details`);
  }, [navigate, mediaItem.id, mediaItem.media_type, mediaItem.first_air_date]);

  return (
    <div className="group relative">
      <MediaCard
        item={embyItem}
        size={size}
        onClick={handleCardClick}
        onRequest={() => {
          console.log("🎯 RequestableMediaCard onRequest called");
          onRequest?.(mediaItem);
        }}
        className="h-full transition-all duration-200 group-hover:scale-105 group-hover:shadow-xl"
        status={{
          isInLibrary,
          isRequested,
        }}
      />
    </div>
  );
});