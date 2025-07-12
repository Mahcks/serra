import { Film, Tv, Play, Plus, CheckCircle, Clock } from "lucide-react";
import { type EmbyMediaItem } from "@/types";

interface MediaCardProps {
  item: EmbyMediaItem & {
    // Extended properties for TMDB support
    vote_average?: number;
    release_date?: string;
    first_air_date?: string;
    media_type?: string;
    overview?: string;
  };
  className?: string;
  size?: "sm" | "md" | "lg";
  onClick?: (item: EmbyMediaItem) => void;
  onRequest?: (item: EmbyMediaItem) => void;
  status?: {
    isInLibrary?: boolean;
    isRequested?: boolean;
  };
}

export function MediaCard({
  item,
  className = "",
  size = "md",
  onClick,
  onRequest,
  status,
}: MediaCardProps) {
  const isMovie =
    item.type?.toLowerCase().includes("movie") || item.media_type === "movie";
  const isSeries =
    item.type?.toLowerCase().includes("series") ||
    item.type?.toLowerCase().includes("show") ||
    item.media_type === "tv";

  // Get year from release date
  const releaseDate = item.release_date || item.first_air_date;
  const year = releaseDate ? new Date(releaseDate).getFullYear() : null;

  const sizeClasses = {
    sm: "w-full min-w-32",
    md: "w-full min-w-40", 
    lg: "w-full min-w-48",
  };

  const iconSizes = {
    sm: "w-6 h-6 sm:w-8 sm:h-8",
    md: "w-8 h-8 sm:w-12 sm:h-12",
    lg: "w-10 h-10 sm:w-16 sm:h-16",
  };

  return (
    <div
      className={`group cursor-pointer transition-all duration-300 ${sizeClasses[size]} ${className}`}
      onClick={() => onClick?.(item)}
    >
      {/* Poster */}
      <div className="relative aspect-[2/3] bg-muted rounded-lg overflow-hidden mb-2 border border-border transition-all duration-300 group-hover:shadow-[0_0_30px_rgba(var(--primary),0.6)] group-hover:border-primary group-hover:ring-2 group-hover:ring-primary/40">
        {item.poster ? (
          <img
            src={item.poster}
            alt={item.name}
            className="w-full h-full object-cover transition-all duration-300"
            onError={(e) => {
              e.currentTarget.style.display = "none";
              e.currentTarget.nextElementSibling?.classList.remove("hidden");
            }}
          />
        ) : null}

        {/* Fallback when no poster or poster fails to load */}
        <div
          className={`absolute inset-0 flex items-center justify-center bg-gradient-to-br from-muted to-muted/80 ${
            item.poster ? "hidden" : ""
          }`}
        >
          {isMovie ? (
            <Film className={`${iconSizes[size]} text-muted-foreground`} />
          ) : isSeries ? (
            <Tv className={`${iconSizes[size]} text-muted-foreground`} />
          ) : (
            <Play className={`${iconSizes[size]} text-muted-foreground`} />
          )}
        </div>

        {/* Media Type Badge - Top Left */}
        <div className="absolute top-2 left-2 z-10">
          <div
            className={`px-2 py-1 rounded-md backdrop-blur-md text-white font-medium text-xs ${
              isMovie
                ? "bg-orange-500/90"
                : isSeries
                ? "bg-blue-500/90"
                : "bg-purple-500/90"
            }`}
          >
            <div className="flex items-center gap-1">
              {isMovie ? (
                <>
                  <span className="hidden sm:inline">MOVIE</span>
                </>
              ) : isSeries ? (
                <>
                  <span className="hidden sm:inline">SERIES</span>
                </>
              ) : (
                <>
                  <span className="hidden sm:inline">Media</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Status Badges - Top Right */}
        {status && (status.isInLibrary || status.isRequested) && (
          <div className="absolute top-2 right-2 z-10">
            {status.isInLibrary && (
              <div className="px-2 py-1 rounded-md backdrop-blur-md bg-green-500/90 text-white font-medium text-xs flex items-center gap-1 mb-1">
                <CheckCircle className="w-3 h-3" />
              </div>
            )}
            {status.isRequested && !status.isInLibrary && (
              <div className="px-2 py-1 rounded-md backdrop-blur-md bg-yellow-500/90 text-white font-medium text-xs flex items-center gap-1">
                <Clock className="w-3 h-3" />
              </div>
            )}
          </div>
        )}

        {/* Hover Overlay with Content */}
        <div className="absolute inset-0 bg-gradient-to-t from-black via-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-3">
          <div className="text-white space-y-1">
            {/* Year */}
            {year && <p className="text-gray-300 text-xs">{year}</p>}

            {/* Title */}
            <h3
              className="font-medium text-sm leading-tight line-clamp-2"
              title={item.name}
            >
              {item.name}
            </h3>

            {/* Description */}
            {item.overview && (
              <p
                className="text-gray-200 text-xs leading-snug line-clamp-4"
                title={item.overview}
              >
                {item.overview}
              </p>
            )}

            {/* Request Button */}
            {onRequest && !status?.isInLibrary && !status?.isRequested && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onRequest(item);
                }}
                className="w-full mt-2 bg-primary hover:bg-primary/90 text-primary-foreground py-1.5 px-3 rounded-md transition-colors duration-200 flex items-center justify-center gap-1.5 text-xs font-medium"
              >
                <Plus className="w-3 h-3" />
                Request
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
