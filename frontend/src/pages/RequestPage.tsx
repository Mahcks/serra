import { memo, useCallback, useMemo, useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Slider } from "@/components/ui/slider";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { MediaCard } from "@/components/ui/media-card";
import { DiscoverySections } from "@/components/DiscoverySections";
import { useSettings } from "@/lib/settings";
import { discoverApi } from "@/lib/api";
import { useInfiniteScroll } from "@/hooks/useInfiniteScroll";
import Loading from "@/components/Loading";
import {
  RequestSystemExternal,
  type TMDBMediaItem,
  type EmbyMediaItem,
} from "@/types";
import {
  Search,
  Users,
  Plus,
  Film,
  Tv,
  XCircle,
  Loader2,
  Filter,
  ChevronDown,
  SlidersHorizontal,
  X,
  Calendar,
  Clock,
  Shield,
  Users2,
  Building2,
  Play,
  Hash,
} from "lucide-react";

interface RequestableMediaCardProps {
  item: TMDBMediaItem;
  onRequest?: (item: TMDBMediaItem) => void;
  size?: "sm" | "md" | "lg";
}

const RequestableMediaCard = memo(function RequestableMediaCard({
  item,
  onRequest,
  size = "md",
}: RequestableMediaCardProps) {
  const navigate = useNavigate();
  const isInLibrary = useMemo(() => Math.random() > 0.7, []);
  const isRequested = useMemo(() => Math.random() > 0.8, []);

  const embyItem = useMemo(
    (): EmbyMediaItem & {
      vote_average?: number;
      release_date?: string;
      first_air_date?: string;
      media_type?: string;
      overview?: string;
    } => ({
      id: item.id.toString(),
      name: item.title || item.name || "Unknown Title",
      type:
        item.media_type === "tv" || item.first_air_date ? "Series" : "Movie",
      poster: item.poster_path
        ? `https://image.tmdb.org/t/p/w500${item.poster_path}`
        : "",
      vote_average: item.vote_average,
      release_date: item.release_date,
      first_air_date: item.first_air_date,
      media_type: item.media_type,
      overview: item.overview,
    }),
    [item]
  );

  const handleCardClick = useCallback(() => {
    const mediaType = item.media_type || (item.first_air_date ? 'tv' : 'movie');
    navigate(`/requests/${item.id}/details?type=${mediaType}`);
  }, [navigate, item.id, item.media_type, item.first_air_date]);

  return (
    <div className="group relative">
      <MediaCard
        item={embyItem}
        size={size}
        onClick={handleCardClick}
        onRequest={() => onRequest?.(item)}
        className="h-full transition-all duration-200 group-hover:scale-105 group-hover:shadow-xl"
        status={{
          isInLibrary,
          isRequested,
        }}
      />
    </div>
  );
});


function ContentGrid({ title, data, isLoading, error, onRequest }: {
  title: string;
  data: TMDBMediaItem[];
  isLoading: boolean;
  error: unknown;
  onRequest: (item: TMDBMediaItem) => void;
}) {
  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <XCircle className="w-12 h-12 text-destructive mx-auto mb-4" />
          <p className="text-destructive">Failed to load {title.toLowerCase()}</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-6">
        {[...Array(24)].map((_, i) => (
          <div key={i} className="aspect-[2/3] bg-muted rounded-lg animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-6">
      {data.map((item, index) => (
        <RequestableMediaCard
          key={`${title}-${item.id}-${index}`}
          item={item}
          onRequest={onRequest}
          size="md"
        />
      ))}
    </div>
  );
}

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

// US Content Ratings for movies
const contentRatings = [
  { value: 'all', label: 'All Ratings' },
  { value: 'G', label: 'G - General Audiences' },
  { value: 'PG', label: 'PG - Parental Guidance' },
  { value: 'PG-13', label: 'PG-13 - Parents Cautioned' },
  { value: 'R', label: 'R - Restricted' },
  { value: 'NC-17', label: 'NC-17 - Adults Only' },
];

// TMDB Genre IDs for movies
const movieGenres = [
  { id: 28, name: 'Action' },
  { id: 12, name: 'Adventure' },
  { id: 16, name: 'Animation' },
  { id: 35, name: 'Comedy' },
  { id: 80, name: 'Crime' },
  { id: 99, name: 'Documentary' },
  { id: 18, name: 'Drama' },
  { id: 10751, name: 'Family' },
  { id: 14, name: 'Fantasy' },
  { id: 36, name: 'History' },
  { id: 27, name: 'Horror' },
  { id: 10402, name: 'Music' },
  { id: 9648, name: 'Mystery' },
  { id: 10749, name: 'Romance' },
  { id: 878, name: 'Science Fiction' },
  { id: 10770, name: 'TV Movie' },
  { id: 53, name: 'Thriller' },
  { id: 10752, name: 'War' },
  { id: 37, name: 'Western' },
];



interface FilterParams {
  genres: number[];
  yearFrom: number;
  yearTo: number;
  runtime: [number, number]; // Changed to range slider
  includeAdult: boolean;
  voteCountMin: number;
  studios: Array<{id: number, name: string}>; // Changed to array of selected studios
  streamingServices: number[];
  keywords: string;
  contentRating: string; // Content rating filter (G, PG, PG-13, R, etc.)
}

interface TMDBProvider {
  provider_id: number;
  provider_name: string;
  logo_path?: string;
  display_priority?: number;
}

interface TMDBCompany {
  id: number;
  name: string;
  logo_path?: string;
  origin_country?: string;
}

export function RequestPage() {
  const { settings, isLoading } = useSettings();
  const location = useLocation();
  const navigate = useNavigate();
  
  const searchParams = new URLSearchParams(location.search);
  const activeTab = searchParams.get('tab') || 'discover';

  // Sorting state
  const [movieSort, setMovieSort] = useState(searchParams.get('sort') || 'popularity.desc');
  const [seriesSort, setSeriesSort] = useState('popularity.desc');

  // Streaming services state
  const [streamingServices, setStreamingServices] = useState<Array<{id: number, name: string, logo?: string}>>([]);
  const [showAllServices, setShowAllServices] = useState(false);
  
  // Studio search state
  const [studioSearchQuery, setStudioSearchQuery] = useState('');
  const [studioSearchResults, setStudioSearchResults] = useState<TMDBCompany[]>([]);
  const [isSearchingStudios, setIsSearchingStudios] = useState(false);

  // Fetch streaming services on component mount
  useEffect(() => {
    const fetchStreamingServices = async () => {
      try {
        const response = await discoverApi.getWatchProviders('movie');
        
        // Define priority order for most common services
        const priorityServices = [
          { id: 8, name: 'Netflix', logo: '🎬' },
          { id: 337, name: 'Disney Plus', logo: '🏰' },
          { id: 15, name: 'Hulu', logo: '📺' },
          { id: 9, name: 'Amazon Prime Video', logo: '📦' },
          { id: 384, name: 'HBO Max', logo: '🎭' },
          { id: 386, name: 'Peacock', logo: '🦚' },
          { id: 387, name: 'Paramount Plus', logo: '⭐' },
          { id: 350, name: 'Apple TV Plus', logo: '🍎' },
          { id: 2, name: 'Apple iTunes', logo: '🎵' },
          { id: 3, name: 'Google Play Movies & TV', logo: '▶️' },
        ];
        
        // Map TMDB response to our format
        const allServices = response.results.map((provider: TMDBProvider) => ({
          id: provider.provider_id,
          name: provider.provider_name,
          logo: provider.logo_path ? `https://image.tmdb.org/t/p/original${provider.logo_path}` : '📺'
        }));
        
        // Sort services: priority services first, then alphabetically
        const sortedServices = allServices.sort((a: {id: number, name: string, logo: string}, b: {id: number, name: string, logo: string}) => {
          const aPriority = priorityServices.findIndex(p => p.id === a.id);
          const bPriority = priorityServices.findIndex(p => p.id === b.id);
          
          // If both are priority services, sort by priority order
          if (aPriority !== -1 && bPriority !== -1) {
            return aPriority - bPriority;
          }
          
          // If only one is priority, priority goes first
          if (aPriority !== -1) return -1;
          if (bPriority !== -1) return 1;
          
          // If neither is priority, sort alphabetically
          return a.name.localeCompare(b.name);
        });
        
        setStreamingServices(sortedServices);
      } catch (error) {
        console.error('Failed to fetch streaming services:', error);
        // Fallback to priority services if API fails
        setStreamingServices([
          { id: 8, name: 'Netflix', logo: '🎬' },
          { id: 337, name: 'Disney Plus', logo: '🏰' },
          { id: 15, name: 'Hulu', logo: '📺' },
          { id: 9, name: 'Amazon Prime Video', logo: '📦' },
          { id: 384, name: 'HBO Max', logo: '🎭' },
          { id: 386, name: 'Peacock', logo: '🦚' },
          { id: 387, name: 'Paramount Plus', logo: '⭐' },
          { id: 350, name: 'Apple TV Plus', logo: '🍎' },
        ]);
      }
    };

    fetchStreamingServices();
  }, []);

  // Filter state
  const [movieFilters, setMovieFilters] = useState<FilterParams>({
    genres: searchParams.get('genres')?.split(',').map(Number).filter(Boolean) || [],
    yearFrom: Number(searchParams.get('yearFrom')) || 1900,
    yearTo: Number(searchParams.get('yearTo')) || new Date().getFullYear(),
    runtime: searchParams.get('runtime')?.split(',').map(Number) as [number, number] || [0, 300],
    includeAdult: searchParams.get('includeAdult') === 'true',
    voteCountMin: Number(searchParams.get('voteCountMin')) || 0,
    studios: [], // Initialize as empty array, will be populated from URL if needed
    streamingServices: searchParams.get('streamingServices')?.split(',').map(Number).filter(Boolean) || [],
    keywords: searchParams.get('keywords') || '',
    contentRating: searchParams.get('contentRating') || 'all',
  });

  // Update URL when filters change
  const updateURL = useCallback((newFilters?: FilterParams, newSort?: string) => {
    const params = new URLSearchParams(location.search);
    
    if (newSort) {
      params.set('sort', newSort);
    }
    
    if (newFilters) {
      if (newFilters.genres.length > 0) {
        params.set('genres', newFilters.genres.join(','));
      } else {
        params.delete('genres');
      }
      
      if (newFilters.yearFrom !== 1900) params.set('yearFrom', newFilters.yearFrom.toString());
      else params.delete('yearFrom');
      
      if (newFilters.yearTo !== new Date().getFullYear()) params.set('yearTo', newFilters.yearTo.toString());
      else params.delete('yearTo');
      
      if (newFilters.runtime[0] !== 0 || newFilters.runtime[1] !== 300) {
        params.set('runtime', newFilters.runtime.join(','));
      } else {
        params.delete('runtime');
      }
      
      if (newFilters.includeAdult) params.set('includeAdult', 'true');
      else params.delete('includeAdult');
      
      if (newFilters.voteCountMin !== 0) params.set('voteCountMin', newFilters.voteCountMin.toString());
      else params.delete('voteCountMin');
      
      if (newFilters.studios.length > 0) {
        const studioIds = newFilters.studios.map(s => s.id.toString()).join(',');
        params.set('studios', studioIds);
      } else {
        params.delete('studios');
      }
      
      if (newFilters.streamingServices.length > 0) params.set('streamingServices', newFilters.streamingServices.join(','));
      else params.delete('streamingServices');
      
      if (newFilters.keywords.trim()) params.set('keywords', newFilters.keywords.trim());
      else params.delete('keywords');
      
      if (newFilters.contentRating && newFilters.contentRating !== 'all') params.set('contentRating', newFilters.contentRating);
      else params.delete('contentRating');
    }
    
    navigate(`${location.pathname}?${params.toString()}`, { replace: true });
  }, [location, navigate]);

  // Data fetching for different content types
  const moviesScroll = useInfiniteScroll({
    queryKey: ["movies", movieSort, JSON.stringify(movieFilters)],
    queryFn: (page) => discoverApi.discoverMovies({
      page,
      sort_by: movieSort,
      with_genres: movieFilters.genres.length > 0 ? movieFilters.genres.join(',') : undefined,
      release_date_gte: movieFilters.yearFrom !== 1900 ? `${movieFilters.yearFrom}-01-01` : undefined,
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

  const seriesScroll = useInfiniteScroll({
    queryKey: ["series", seriesSort],
    queryFn: (page) => discoverApi.getTVWithSort(page, seriesSort),
  });


  const handleRequest = useCallback((item: TMDBMediaItem) => {
    console.log("Requesting:", item);
    alert(`Request submitted for: ${item.title || item.name}`);
  }, []);

  // Filter handlers
  const handleSortChange = useCallback((newSort: string) => {
    setMovieSort(newSort);
    updateURL(undefined, newSort);
  }, [updateURL]);

  const handleFiltersChange = useCallback((newFilters: FilterParams) => {
    setMovieFilters(newFilters);
    updateURL(newFilters);
  }, [updateURL]);

  const toggleGenre = useCallback((genreId: number) => {
    const newGenres = movieFilters.genres.includes(genreId)
      ? movieFilters.genres.filter(id => id !== genreId)
      : [...movieFilters.genres, genreId];
    
    const newFilters = { ...movieFilters, genres: newGenres };
    handleFiltersChange(newFilters);
  }, [movieFilters, handleFiltersChange]);


  const toggleStreamingService = useCallback((serviceId: number) => {
    const newServices = movieFilters.streamingServices.includes(serviceId)
      ? movieFilters.streamingServices.filter(id => id !== serviceId)
      : [...movieFilters.streamingServices, serviceId];
    
    const newFilters = { ...movieFilters, streamingServices: newServices };
    handleFiltersChange(newFilters);
  }, [movieFilters, handleFiltersChange]);

  // Studio search functionality
  const searchStudios = useCallback(async (query: string) => {
    if (query.length < 2) {
      setStudioSearchResults([]);
      return;
    }
    
    setIsSearchingStudios(true);
    try {
      const response = await discoverApi.searchCompanies(query);
      setStudioSearchResults(response.results || []);
    } catch (error) {
      console.error('Failed to search studios:', error);
      setStudioSearchResults([]);
    } finally {
      setIsSearchingStudios(false);
    }
  }, []);

  const addStudio = useCallback((studio: TMDBCompany) => {
    // Check if studio is already added
    if (movieFilters.studios.some(s => s.id === studio.id)) {
      return;
    }
    
    const newStudios = [...movieFilters.studios, { id: studio.id, name: studio.name }];
    const newFilters = { ...movieFilters, studios: newStudios };
    handleFiltersChange(newFilters);
    setStudioSearchQuery('');
    setStudioSearchResults([]);
  }, [movieFilters, handleFiltersChange]);

  const removeStudio = useCallback((studioId: number) => {
    const newStudios = movieFilters.studios.filter(s => s.id !== studioId);
    const newFilters = { ...movieFilters, studios: newStudios };
    handleFiltersChange(newFilters);
  }, [movieFilters, handleFiltersChange]);

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

  const clearFilters = useCallback(() => {
    const defaultFilters: FilterParams = {
      genres: [],
      yearFrom: 1900,
      yearTo: new Date().getFullYear(),
      runtime: [0, 300],
      includeAdult: false,
      voteCountMin: 0,
      studios: [],
      streamingServices: [],
      keywords: '',
      contentRating: 'all',
    };
    handleFiltersChange(defaultFilters);
  }, [handleFiltersChange]);

  const hasActiveFilters = useMemo(() => {
    return movieFilters.genres.length > 0 ||
           movieFilters.yearFrom !== 1900 ||
           movieFilters.yearTo !== new Date().getFullYear() ||
           movieFilters.runtime[0] !== 0 ||
           movieFilters.runtime[1] !== 300 ||
           movieFilters.includeAdult ||
           movieFilters.voteCountMin !== 0 ||
           movieFilters.studios.length > 0 ||
           movieFilters.streamingServices.length > 0 ||
           movieFilters.keywords.trim().length > 0 ||
           (movieFilters.contentRating && movieFilters.contentRating !== 'all');
  }, [movieFilters]);

  if (isLoading) return <Loading />;

  if (settings?.request_system === RequestSystemExternal) {
    return (
      <iframe
        src={settings.request_system_url}
        className="border-0 bg-gray-900 w-full h-full"
        title="External Request System"
        sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-top-navigation-by-user-activation allow-storage-access-by-user-activation"
        referrerPolicy="strict-origin-when-cross-origin"
        allow="camera; microphone; geolocation; storage-access"
        style={{
          colorScheme: "dark",
          minHeight: "100vh",
          display: "block",
          width: "100%",
        }}
      />
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-6 py-8">
        {/* Content based on sidebar navigation */}
        {activeTab === 'discover' && (
          <div className="space-y-8">
            <div className="space-y-6">
              {/* Search Bar */}
              <div className="relative max-w-2xl mx-auto">
                <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-muted-foreground h-5 w-5" />
                <Input
                  placeholder="Search for movies, TV shows, and more..."
                  className="pl-12 py-3 text-lg rounded-xl"
                />
              </div>
            </div>
            
            {/* Discovery Carousels */}
            <DiscoverySections />
          </div>
        )}

        {activeTab === 'movies' && (
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
                {/* Advanced Filters */}
                <Sheet>
                  <SheetTrigger asChild >
                    <Button variant="outline" size="sm" className="gap-2">
                      <SlidersHorizontal className="w-4 h-4" />
                      Filters
                      {hasActiveFilters && (
                        <Badge variant="secondary" className="ml-1 px-1.5 py-0.5 text-xs">
                          {movieFilters.genres.length + 
                           (movieFilters.yearFrom !== 1900 ? 1 : 0) +
                           (movieFilters.yearTo !== new Date().getFullYear() ? 1 : 0) +
                           (movieFilters.runtime[0] !== 0 || movieFilters.runtime[1] !== 300 ? 1 : 0) +
                           (movieFilters.includeAdult ? 1 : 0) +
                           (movieFilters.voteCountMin !== 0 ? 1 : 0) +
                           movieFilters.studios.length +
                           movieFilters.streamingServices.length +
                           (movieFilters.keywords.trim() ? 1 : 0) +
                           (movieFilters.contentRating && movieFilters.contentRating !== 'all' ? 1 : 0)}
                        </Badge>
                      )}
                    </Button>
                  </SheetTrigger>
                  <SheetContent className="overflow-y-auto px-5">
                    <SheetHeader>
                      <SheetTitle>Filter Movies</SheetTitle>
                      <SheetDescription>
                        Narrow down your search using these filters
                      </SheetDescription>
                    </SheetHeader>
                    
                    <div className="space-y-6 mt-6 pb-6">
                      {/* Genres */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Film className="w-4 h-4" />
                          Genres
                        </h4>
                        <div className="flex flex-wrap gap-2">
                          {movieGenres.map((genre) => (
                            <Badge
                              key={genre.id}
                              variant={movieFilters.genres.includes(genre.id) ? "default" : "outline"}
                              className="cursor-pointer"
                              onClick={() => toggleGenre(genre.id)}
                            >
                              {genre.name}
                            </Badge>
                          ))}
                        </div>
                      </div>

                      {/* Content Rating */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Shield className="w-4 h-4" />
                          Content Rating
                        </h4>
                        <Select
                          value={movieFilters.contentRating}
                          onValueChange={(value) => handleFiltersChange({ ...movieFilters, contentRating: value })}
                        >
                          <SelectTrigger className="w-full">
                            <SelectValue placeholder="Select content rating" />
                          </SelectTrigger>
                          <SelectContent>
                            {contentRatings.map((rating) => (
                              <SelectItem key={rating.value} value={rating.value}>
                                {rating.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <p className="text-xs text-muted-foreground mt-1">
                          Content ratings are estimated based on genres and movie data
                        </p>
                      </div>

                      {/* Year Range */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Calendar className="w-4 h-4" />
                          Release Year
                        </h4>
                        <div className="space-y-4">
                          <div>
                            <label className="text-sm text-muted-foreground">From: {movieFilters.yearFrom}</label>
                            <Slider
                              value={[movieFilters.yearFrom]}
                              onValueChange={(values) => handleFiltersChange({ ...movieFilters, yearFrom: values[0] })}
                              min={1900}
                              max={new Date().getFullYear()}
                              step={1}
                              className="mt-2"
                            />
                          </div>
                          <div>
                            <label className="text-sm text-muted-foreground">To: {movieFilters.yearTo}</label>
                            <Slider
                              value={[movieFilters.yearTo]}
                              onValueChange={(values) => handleFiltersChange({ ...movieFilters, yearTo: values[0] })}
                              min={1900}
                              max={new Date().getFullYear()}
                              step={1}
                              className="mt-2"
                            />
                          </div>
                        </div>
                      </div>

                      {/* Runtime Range - Double Slider */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Clock className="w-4 h-4" />
                          Runtime
                        </h4>
                        <div className="space-y-3">
                          <div className="flex justify-between text-sm text-muted-foreground">
                            <span>{movieFilters.runtime[0]} min</span>
                            <span>{movieFilters.runtime[1]} min</span>
                          </div>
                          <Slider
                            value={movieFilters.runtime}
                            onValueChange={(values) => handleFiltersChange({ ...movieFilters, runtime: values as [number, number] })}
                            min={0}
                            max={300}
                            step={5}
                            className="mt-2"
                          />
                        </div>
                      </div>

                      {/* Studios */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Building2 className="w-4 h-4" />
                          Studios
                        </h4>
                        
                        {/* Selected Studios */}
                        {movieFilters.studios.length > 0 && (
                          <div className="flex flex-wrap gap-2 mb-3">
                            {movieFilters.studios.map((studio) => (
                              <Badge
                                key={studio.id}
                                variant="default"
                                className="cursor-pointer flex items-center gap-1"
                                onClick={() => removeStudio(studio.id)}
                              >
                                {studio.name}
                                <X className="w-3 h-3" />
                              </Badge>
                            ))}
                          </div>
                        )}
                        
                        {/* Search Input */}
                        <div className="relative">
                          <Input
                            placeholder="Search for studios..."
                            value={studioSearchQuery}
                            onChange={(e) => {
                              setStudioSearchQuery(e.target.value);
                              searchStudios(e.target.value);
                            }}
                            className="text-sm"
                          />
                          {isSearchingStudios && (
                            <div className="absolute right-2 top-1/2 -translate-y-1/2">
                              <Loader2 className="w-4 h-4 animate-spin" />
                            </div>
                          )}
                        </div>
                        
                        {/* Search Results */}
                        {studioSearchResults.length > 0 && (
                          <div className="mt-2 border rounded-md bg-popover max-h-32 overflow-y-auto">
                            {studioSearchResults.map((studio) => (
                              <div
                                key={studio.id}
                                className="px-3 py-2 hover:bg-accent cursor-pointer text-sm border-b last:border-b-0"
                                onClick={() => addStudio(studio)}
                              >
                                <div className="font-medium">{studio.name}</div>
                                {studio.origin_country && (
                                  <div className="text-xs text-muted-foreground">{studio.origin_country}</div>
                                )}
                              </div>
                            ))}
                          </div>
                        )}
                        
                        <p className="text-xs text-muted-foreground mt-1">
                          Search and click to add studios
                        </p>
                      </div>

                      {/* Streaming Services */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Play className="w-4 h-4" />
                          Streaming Services
                        </h4>
                        
                        {/* Most Common Services */}
                        <div className="flex flex-wrap gap-3">
                          {(showAllServices ? streamingServices : streamingServices.slice(0, 8)).map((service) => (
                            <div
                              key={service.id}
                              className={`cursor-pointer rounded-lg border-2 transition-all duration-200 hover:scale-105 w-16 h-16 flex items-center justify-center p-2 ${
                                movieFilters.streamingServices.includes(service.id) 
                                  ? 'border-primary bg-primary/10' 
                                  : 'border-border hover:border-primary/50'
                              }`}
                              onClick={() => toggleStreamingService(service.id)}
                              title={service.name}
                            >
                              {service.logo && service.logo.startsWith('http') ? (
                                <img 
                                  src={service.logo} 
                                  alt={service.name}
                                  className="w-full h-full object-contain"
                                  onError={(e) => {
                                    // Fallback to emoji if image fails to load
                                    e.currentTarget.style.display = 'none';
                                    e.currentTarget.nextElementSibling?.classList.remove('hidden');
                                  }}
                                />
                              ) : (
                                <span className="text-2xl">{service.logo || '📺'}</span>
                              )}
                              <span className="hidden text-2xl">📺</span>
                            </div>
                          ))}
                        </div>
                        
                        {/* Show More/Less Button */}
                        {streamingServices.length > 8 && (
                          <div className="mt-3">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setShowAllServices(!showAllServices)}
                              className="text-xs text-muted-foreground hover:text-foreground"
                            >
                              {showAllServices ? (
                                <>
                                  <ChevronDown className="w-3 h-3 mr-1 rotate-180" />
                                  Show Less
                                </>
                              ) : (
                                <>
                                  <ChevronDown className="w-3 h-3 mr-1" />
                                  Show More ({streamingServices.length - 8} more)
                                </>
                              )}
                            </Button>
                          </div>
                        )}
                      </div>

                      {/* Keywords */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Hash className="w-4 h-4" />
                          Keywords
                        </h4>
                        <Input
                          placeholder="e.g. superhero, space, comedy..."
                          value={movieFilters.keywords}
                          onChange={(e) => handleFiltersChange({ ...movieFilters, keywords: e.target.value })}
                          className="text-sm"
                        />
                        <p className="text-xs text-muted-foreground mt-1">
                          Separate multiple keywords with commas
                        </p>
                      </div>

                      {/* Popularity & Content Filters */}
                      <div>
                        <h4 className="font-medium mb-3 flex items-center gap-2">
                          <Users2 className="w-4 h-4" />
                          Popularity & Content
                        </h4>
                        <div className="space-y-4">
                          <div>
                            <label className="text-sm text-muted-foreground">Minimum Vote Count: {movieFilters.voteCountMin}</label>
                            <Slider
                              value={[movieFilters.voteCountMin]}
                              onValueChange={(values) => handleFiltersChange({ ...movieFilters, voteCountMin: values[0] })}
                              min={0}
                              max={10000}
                              step={100}
                              className="mt-2"
                            />
                          </div>
                          <div className="flex items-center space-x-2">
                            <Checkbox
                              id="includeAdult"
                              checked={movieFilters.includeAdult}
                              onCheckedChange={(checked) => 
                                handleFiltersChange({ ...movieFilters, includeAdult: checked as boolean })
                              }
                            />
                            <Label htmlFor="includeAdult" className="text-sm">
                              Include adult content
                            </Label>
                          </div>
                        </div>
                      </div>

                      {/* Clear Filters */}
                      {hasActiveFilters && (
                        <Button variant="outline" onClick={clearFilters} className="w-full">
                          <X className="w-4 h-4 mr-2" />
                          Clear All Filters
                        </Button>
                      )}
                    </div>
                  </SheetContent>
                </Sheet>

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
              onRequest={handleRequest}
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
        )}

        {activeTab === 'series' && (
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
              onRequest={handleRequest}
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
        )}

        {activeTab === 'requests' && (
          <div className="space-y-6">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-2 bg-muted/50 rounded-lg border">
                <Users className="w-6 h-6 text-muted-foreground" />
              </div>
              <div>
                <h2 className="text-2xl font-bold text-foreground">My Requests</h2>
                <p className="text-muted-foreground">Track your content requests</p>
              </div>
            </div>
            <div className="bg-muted/50 rounded-xl p-8 text-center border">
              <Users className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-foreground mb-2">No Requests Yet</h3>
              <p className="text-muted-foreground mb-6">Start by requesting some content you'd like to watch</p>
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                Make Your First Request
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}