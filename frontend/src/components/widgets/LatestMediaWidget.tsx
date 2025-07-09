import { useQuery } from "@tanstack/react-query";
import { backendApi } from "@/lib/api";
import { type EmbyMediaItem } from "@/types";
import { Play, AlertTriangle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { MediaCard } from "@/components/ui/media-card";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";

export default function LatestMediaWidget() {
  const {
    data: latestMedia,
    isLoading,
    error,
    refetch,
  } = useQuery<EmbyMediaItem[]>({
    queryKey: ["latestMedia"],
    queryFn: backendApi.getLatestMedia,
    retry: 2,
    refetchInterval: 5 * 60 * 1000, // Refetch every 5 minutes
  });

  if (error) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 bg-destructive rounded-xl">
            <AlertTriangle className="w-4 h-4 sm:w-5 sm:h-5 text-destructive-foreground" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Error Loading Media
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Failed to load latest media items
            </p>
          </div>
        </div>
        <Button onClick={() => refetch()} variant="outline" size="sm">
          <Loader2 className="w-4 h-4 mr-2" />
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
            <Play className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground animate-pulse" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Latest Media
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Loading recent additions...
            </p>
          </div>
        </div>

        <div className="flex gap-4 overflow-hidden">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div key={i} className="flex-shrink-0 w-32 sm:w-40 space-y-2">
              <Skeleton className="aspect-[2/3] w-full rounded-lg" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-3 w-2/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!latestMedia || latestMedia.length === 0) {
    return (
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 bg-primary rounded-xl">
            <Play className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-foreground">
              Latest Media
            </h2>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Recently added to your library
            </p>
          </div>
        </div>
        
        <div className="text-center py-8 sm:py-12">
          <div className="p-3 sm:p-4 bg-muted rounded-full w-fit mx-auto mb-3 sm:mb-4">
            <Play className="w-6 h-6 sm:w-8 sm:h-8 text-muted-foreground" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-foreground mb-1 sm:mb-2">
            No media found
          </h3>
          <p className="text-sm sm:text-base text-muted-foreground">
            No recent additions to your media library
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
      {/* Header */}
      <div className="flex items-center gap-3 mb-4 sm:mb-6">
        <div className="p-2 bg-primary rounded-xl">
          <Play className="w-4 h-4 sm:w-5 sm:h-5 text-primary-foreground" />
        </div>
        <div>
          <h2 className="text-lg sm:text-xl font-bold text-foreground">
            Latest Media
          </h2>
          <p className="text-muted-foreground text-xs sm:text-sm">
            Recently added to your library ({latestMedia.length} items)
          </p>
        </div>
      </div>

      {/* Media Carousel */}
      <Carousel
        opts={{
          align: "start",
          slidesToScroll: "auto",
          containScroll: "trimSnaps",
        }}
        className="w-full max-w-full"
      >
        <CarouselContent className="-ml-2 md:-ml-4">
          {latestMedia.map((item) => (
            <CarouselItem key={item.id} className="pl-2 md:pl-4 basis-auto">
              <MediaCard 
                item={item} 
                size="md"
                onClick={(mediaItem) => {
                  // TODO: Add navigation to media details page
                  console.log('Clicked media item:', mediaItem.name, mediaItem.id);
                }}
              />
            </CarouselItem>
          ))}
        </CarouselContent>
        <CarouselPrevious className="hidden sm:flex -left-4 lg:-left-12" />
        <CarouselNext className="hidden sm:flex -right-4 lg:-right-12" />
      </Carousel>
    </div>
  );
}
