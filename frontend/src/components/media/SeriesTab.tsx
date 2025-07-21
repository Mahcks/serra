import { useState } from "react";
import { useLocation } from "react-router-dom";
import { Tv, Filter, Loader2 } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ContentGrid } from "@/components/media/ContentGrid";
import { useInfiniteScroll } from "@/hooks/useInfiniteScroll";
import { discoverApi } from "@/lib/api";
import { type TMDBMediaItem } from "@/types";

const sortOptions = [
  { value: 'popularity.desc', label: 'Most Popular' },
  { value: 'popularity.asc', label: 'Least Popular' },
  { value: 'vote_average.desc', label: 'Highest Rated' },
  { value: 'vote_average.asc', label: 'Lowest Rated' },
  { value: 'release_date.desc', label: 'Newest First' },
  { value: 'release_date.asc', label: 'Oldest First' },
  { value: 'title.asc', label: 'A-Z' },
  { value: 'title.desc', label: 'Z-A' },
];

interface SeriesTabProps {
  onRequest: (item: TMDBMediaItem) => void;
  isRequestLoading?: boolean;
}

export function SeriesTab({ onRequest, isRequestLoading }: SeriesTabProps) {
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  
  // Get URL parameters for filtering
  const urlFilters = {
    'first_air_date.gte': searchParams.get('first_air_date.gte'),
    'air_date.gte': searchParams.get('air_date.gte'),
    'air_date.lte': searchParams.get('air_date.lte'),
    'first_air_date.lte': searchParams.get('first_air_date.lte'),
    first_air_date_year: searchParams.get('first_air_date_year') ? parseInt(searchParams.get('first_air_date_year')!) : undefined,
    with_genres: searchParams.get('with_genres'),
    with_networks: searchParams.get('with_networks') ? parseInt(searchParams.get('with_networks')!) : undefined,
    with_origin_country: searchParams.get('with_origin_country'),
    'vote_average.gte': searchParams.get('vote_average.gte') ? parseFloat(searchParams.get('vote_average.gte')!) : undefined,
    'vote_average.lte': searchParams.get('vote_average.lte') ? parseFloat(searchParams.get('vote_average.lte')!) : undefined,
  };

  // Use URL sort_by parameter if available, otherwise default to popularity.desc
  const defaultSort = searchParams.get('sort_by') || 'popularity.desc';
  const [seriesSort, setSeriesSort] = useState(defaultSort);

  const seriesScroll = useInfiniteScroll({
    queryKey: ["series", seriesSort, urlFilters],
    queryFn: (page) => {
      // Filter out undefined values and create params object
      const params = Object.fromEntries(
        Object.entries({
          page,
          sort_by: seriesSort,
          ...urlFilters
        }).filter(([_, value]) => value !== null && value !== undefined)
      );
      
      
      return discoverApi.discoverTV(params);
    },
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-secondary/50 rounded-lg border">
            <Tv className="w-6 h-6 text-secondary-foreground" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-foreground">Series</h2>
            <p className="text-muted-foreground">Discover and request TV shows</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-muted-foreground" />
          <Select value={seriesSort} onValueChange={setSeriesSort}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Sort by..." />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      
      <ContentGrid
        title="Series"
        data={seriesScroll.items}
        isLoading={seriesScroll.isLoading}
        error={seriesScroll.isError}
        onRequest={onRequest}
        isRequestLoading={isRequestLoading}
      />
      
      <div ref={seriesScroll.loadingRef} className="flex justify-center py-8">
        {seriesScroll.isLoadingMore && (
          <div className="flex items-center gap-3 text-muted-foreground">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>Loading more series...</span>
          </div>
        )}
      </div>
    </div>
  );
}