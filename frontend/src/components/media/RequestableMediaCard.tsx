import { memo, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { MediaCard } from "@/components/ui/media-card";
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
  
  // Use the actual status from the enhanced API response
  const isInLibrary = 'in_library' in item ? item.in_library : false;
  const isRequested = 'requested' in item ? item.requested : false;
  
  // Extract the TMDBMediaItem from TMDBFullMediaItem if needed
  const mediaItem = 'TMDBMediaItem' in item ? item.TMDBMediaItem : item;

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