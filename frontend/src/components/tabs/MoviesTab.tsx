import { useState, useCallback, useMemo } from "react";
import { useLocation } from "react-router-dom";
import { Film, Filter, Loader2 } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ContentGrid } from "@/components/ContentGrid";
import { useInfiniteScroll } from "@/hooks/useInfiniteScroll";
import { useMovieFilters } from "@/hooks/useMovieFilters";
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

interface MoviesTabProps {
  onRequest: (item: TMDBMediaItem) => void;
  isRequestLoading?: boolean;
}

export function MoviesTab({ onRequest, isRequestLoading }: MoviesTabProps) {
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const [movieSort, setMovieSort] = useState(searchParams.get('sort_by') || searchParams.get('sort') || 'popularity.desc');
  
  const { movieFilters, updateURL } = useMovieFilters();

  // Get URL parameters for additional filtering (like upcoming movies)
  const urlReleaseDateGte = searchParams.get('release_date.gte');
  
  // Data fetching for movies
  const moviesScroll = useInfiniteScroll({
    queryKey: ["movies", movieSort, JSON.stringify(movieFilters), urlReleaseDateGte],
    queryFn: (page) => discoverApi.discoverMovies({
      page,
      sort_by: movieSort,
      with_genres: movieFilters.genres.length > 0 ? movieFilters.genres.join(',') : undefined,
      // Use URL release_date.gte if provided, otherwise use filter yearFrom
      release_date_gte: urlReleaseDateGte || (movieFilters.yearFrom !== 1900 ? `${movieFilters.yearFrom}-01-01` : undefined),
      release_date_lte: movieFilters.yearTo !== new Date().getFullYear() ? `${movieFilters.yearTo}-12-31` : undefined,
      with_runtime_gte: movieFilters.runtime[0] !== 0 ? movieFilters.runtime[0] : undefined,
      with_runtime_lte: movieFilters.runtime[1] !== 300 ? movieFilters.runtime[1] : undefined,
      include_adult: movieFilters.includeAdult || undefined,
      vote_count_gte: movieFilters.voteCountMin !== 0 ? movieFilters.voteCountMin : undefined,
      with_companies: movieFilters.studios.length > 0 ? movieFilters.studios.map(s => s.id).join(',') : undefined,
      with_watch_providers: movieFilters.streamingServices.length > 0 ? movieFilters.streamingServices.join(',') : undefined,
      watch_region: movieFilters.streamingServices.length > 0 ? 'US' : undefined,
      with_keywords: movieFilters.keywords.trim() || undefined,
    }),
  });

  const handleSortChange = useCallback((newSort: string) => {
    setMovieSort(newSort);
    updateURL(undefined, newSort);
  }, [updateURL]);

  // Smart client-side content rating estimation
  const getContentRating = useCallback((movie: TMDBMediaItem): string => {
    const genreIds = movie.genre_ids || [];
    const adult = movie.adult || false;
    const voteAverage = movie.vote_average || 0;
    
    // Adult content is automatically NC-17
    if (adult) return 'NC-17';
    
    // Genre-based classification
    const horrorThrillerIds = [27, 53]; // Horror, Thriller
    const crimeActionIds = [80, 28]; // Crime, Action
    const familyAnimationIds = [10751, 16]; // Family, Animation
    const comedyRomanceIds = [35, 10749]; // Comedy, Romance
    
    const hasHorrorThriller = genreIds.some(id => horrorThrillerIds.includes(id));
    const hasCrimeAction = genreIds.some(id => crimeActionIds.includes(id));
    const hasFamilyAnimation = genreIds.some(id => familyAnimationIds.includes(id));
    const hasComedyRomance = genreIds.some(id => comedyRomanceIds.includes(id));
    
    // G - Family/Animation with high ratings
    if (hasFamilyAnimation && voteAverage >= 7.0) return 'G';
    
    // PG - Family content or light comedy/romance
    if (hasFamilyAnimation || (hasComedyRomance && voteAverage >= 6.0)) return 'PG';
    
    // R - Horror/Thriller or Crime with lower ratings  
    if (hasHorrorThriller || (hasCrimeAction && voteAverage < 7.0)) return 'R';
    
    // NC-17 - More inclusive: any horror/crime content with low ratings OR adult flag
    if ((hasHorrorThriller || hasCrimeAction) && voteAverage < 6.0) return 'NC-17';
    
    // PG-13 - Default for most other content
    return 'PG-13';
  }, []);

  // Filter movies by content rating
  const filteredMovies = useMemo(() => {
    if (movieFilters.contentRating === 'all') {
      return moviesScroll.items;
    }
    
    const filtered = moviesScroll.items.filter(movie => {
      const rating = getContentRating(movie);
      return rating === movieFilters.contentRating;
    });
    
    // If we have very few results, log it to help debug
    if (filtered.length < 5 && moviesScroll.items.length > 20) {
      console.log(`Content rating filter "${movieFilters.contentRating}" found ${filtered.length} movies out of ${moviesScroll.items.length} total`);
    }
    
    return filtered;
  }, [moviesScroll.items, movieFilters.contentRating, getContentRating]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary/20 rounded-lg border">
            <Film className="w-6 h-6 text-primary" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-foreground">Movies</h2>
            <p className="text-muted-foreground">Discover and request movies</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {/* TODO: Add MovieFiltersSheet component here */}
          
          {/* Sort Dropdown */}
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-muted-foreground" />
            <Select value={movieSort} onValueChange={handleSortChange}>
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
      </div>
      
      <ContentGrid
        title="Movies"
        data={filteredMovies}
        isLoading={moviesScroll.isLoading}
        error={moviesScroll.isError}
        onRequest={onRequest}
        isRequestLoading={isRequestLoading}
      />
      
      {/* Only show infinite scroll loading when content rating is 'all' */}
      {movieFilters.contentRating === 'all' && (
        <div ref={moviesScroll.loadingRef} className="flex justify-center py-8">
          {moviesScroll.isLoadingMore && (
            <div className="flex items-center gap-3 text-muted-foreground">
              <Loader2 className="w-5 h-5 animate-spin" />
              <span>Loading more movies...</span>
            </div>
          )}
        </div>
      )}
      
      {/* Show message when content rating filter is active */}
      {movieFilters.contentRating !== 'all' && (
        <div className="flex justify-center py-8">
          <div className="text-center text-muted-foreground">
            <p className="text-sm">
              Showing {filteredMovies.length} movies with {movieFilters.contentRating} rating
            </p>
            <p className="text-xs mt-1">
              Content rating filters show results from loaded pages only
            </p>
          </div>
        </div>
      )}
    </div>
  );
}