import { Film, Tv, Play } from "lucide-react";
import { type EmbyMediaItem } from "@/types";

interface MediaCardProps {
  item: EmbyMediaItem;
  className?: string;
  size?: "sm" | "md" | "lg";
  onClick?: (item: EmbyMediaItem) => void;
}

export function MediaCard({ item, className = "", size = "md", onClick }: MediaCardProps) {
  const isMovie = item.type?.toLowerCase().includes('movie');
  const isSeries = item.type?.toLowerCase().includes('series') || item.type?.toLowerCase().includes('show');
  
  const sizeClasses = {
    sm: "w-24 sm:w-32",
    md: "w-32 sm:w-40", 
    lg: "w-40 sm:w-48"
  };

  const iconSizes = {
    sm: "w-6 h-6 sm:w-8 sm:h-8",
    md: "w-8 h-8 sm:w-12 sm:h-12",
    lg: "w-10 h-10 sm:w-16 sm:h-16"
  };

  const textSizes = {
    sm: "text-xs",
    md: "text-sm",
    lg: "text-base"
  };

  return (
    <div
      className={`group cursor-pointer transition-all duration-200 hover:scale-105 ${sizeClasses[size]} ${className}`}
      onClick={() => onClick?.(item)}
    >
      {/* Poster */}
      <div className="relative aspect-[2/3] bg-muted rounded-lg overflow-hidden mb-2 border border-border">
        {item.poster ? (
          <img
            src={item.poster}
            alt={item.name}
            className="w-full h-full object-cover transition-transform duration-200 group-hover:scale-110"
            onError={(e) => {
              e.currentTarget.style.display = 'none';
              e.currentTarget.nextElementSibling?.classList.remove('hidden');
            }}
          />
        ) : null}
        
        {/* Fallback when no poster or poster fails to load */}
        <div className={`absolute inset-0 flex items-center justify-center bg-gradient-to-br from-muted to-muted/80 ${item.poster ? 'hidden' : ''}`}>
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
          <div className={`px-2 py-1 rounded-md backdrop-blur-md text-white font-medium text-xs ${
            isMovie 
              ? "bg-orange-500/90" 
              : isSeries 
              ? "bg-blue-500/90"
              : "bg-purple-500/90"
          }`}>
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

        {/* Hover Overlay */}
        <div className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
      </div>

      {/* Media Info */}
      <div className="space-y-1">
        <h3 
          className={`${textSizes[size]} font-medium text-foreground line-clamp-2 group-hover:text-primary transition-colors duration-200`}
          title={item.name}
        >
          {item.name}
        </h3>
      </div>
    </div>
  );
}