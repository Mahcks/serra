import { useState, useEffect, useMemo } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { discoverApi } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ContentGrid } from "@/components/media/ContentGrid";
import { OnBehalfRequestDialog, SeasonSelectionDialog } from "@/components/media/RequestDialogs";
import { Search, Film, Tv, ArrowLeft, Loader2 } from "lucide-react";
import { type TMDBMediaItem } from "@/types";
import { useAdvancedRequestHandler } from "@/hooks/useAdvancedRequestHandler";

interface CombinedResults {
  combined: (TMDBMediaItem & { in_library?: boolean; requested?: boolean })[];
  movies: (TMDBMediaItem & { in_library?: boolean; requested?: boolean })[];
  tv: (TMDBMediaItem & { in_library?: boolean; requested?: boolean })[];
  totalResults: number;
}

// Function to calculate relevance score for search results
function calculateRelevanceScore(item: TMDBMediaItem, query: string): number {
  const title = (item.title || item.name || '').toLowerCase();
  const queryLower = query.toLowerCase();
  let score = 0;

  // Exact match gets highest score
  if (title === queryLower) {
    score += 100;
  }
  // Title starts with query
  else if (title.startsWith(queryLower)) {
    score += 80;
  }
  // Title includes query
  else if (title.includes(queryLower)) {
    score += 60;
  }

  // Boost score based on popularity (vote_average and vote_count)
  if (item.vote_average && item.vote_count) {
    score += (item.vote_average / 10) * 20; // Up to 20 points for rating
    score += Math.min(item.vote_count / 1000, 10); // Up to 10 points for vote count
  }

  // Boost for recent releases
  const releaseDate = item.release_date || item.first_air_date;
  if (releaseDate) {
    const year = new Date(releaseDate).getFullYear();
    const currentYear = new Date().getFullYear();
    const yearDiff = currentYear - year;
    if (yearDiff <= 5) {
      score += (5 - yearDiff) * 2; // Up to 10 points for recent releases
    }
  }

  return score;
}

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const [activeTab, setActiveTab] = useState('all');

  // Advanced request handling with dialogs and status checking
  const {
    showOnBehalfDialog,
    setShowOnBehalfDialog,
    selectedMedia,
    selectedUser,
    setSelectedUser,
    showSeasonDialog,
    setShowSeasonDialog,
    selectedSeasons,
    setSelectedSeasons,
    allUsers,
    handleRequest,
    handleSeasonRequestSubmit,
    handleOnBehalfSubmit,
    isRequestLoading,
  } = useAdvancedRequestHandler({
    queryKeysToInvalidate: [["searchAll"]],
  });

  // Sync query with URL params
  useEffect(() => {
    const urlQuery = searchParams.get('q') || '';
    if (urlQuery !== query) {
      setQuery(urlQuery);
    }
  }, [searchParams]);

  // Update URL when query changes (debounced to avoid too many updates)
  useEffect(() => {
    const timer = setTimeout(() => {
      if (query.trim()) {
        setSearchParams({ q: query.trim() }, { replace: true });
      } else if (!query.trim() && searchParams.get('q')) {
        setSearchParams({}, { replace: true });
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query, setSearchParams]);

  // Debounce the search query for API calls
  const [debouncedQuery, setDebouncedQuery] = useState(query.trim());
  
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query.trim());
    }, 300);

    return () => clearTimeout(timer);
  }, [query]);

  // Get search results
  const { data: searchResults, isLoading, error } = useQuery({
    queryKey: ["searchAll", debouncedQuery],
    queryFn: () => discoverApi.searchAll(debouncedQuery),
    enabled: debouncedQuery.length > 0,
    staleTime: 30000,
  });

  // Process and rank combined results
  const processedResults: CombinedResults = useMemo(() => {
    if (!searchResults || !debouncedQuery) {
      return { combined: [], movies: [], tv: [], totalResults: 0 };
    }

    const movies = searchResults.movies?.results || [];
    const tv = searchResults.tv?.results || [];

    // Add media_type to items and preserve status fields from enriched response
    const moviesWithType = movies.map((item: any) => {
      // If it's a TMDBFullMediaItem, extract the media item and preserve status
      if ('TMDBMediaItem' in item) {
        return {
          ...item.TMDBMediaItem,
          media_type: 'movie' as const,
          in_library: item.in_library,
          requested: item.requested,
        };
      }
      // Otherwise, it's already a basic TMDBMediaItem
      return {
        ...item,
        media_type: 'movie' as const
      };
    });
    
    const tvWithType = tv.map((item: any) => {
      // If it's a TMDBFullMediaItem, extract the media item and preserve status
      if ('TMDBMediaItem' in item) {
        return {
          ...item.TMDBMediaItem,
          media_type: 'tv' as const,
          in_library: item.in_library,
          requested: item.requested,
        };
      }
      // Otherwise, it's already a basic TMDBMediaItem
      return {
        ...item,
        media_type: 'tv' as const
      };
    });

    // Combine and sort by relevance
    const combined = [...moviesWithType, ...tvWithType]
      .map(item => ({
        ...item,
        relevanceScore: calculateRelevanceScore(item, debouncedQuery)
      }))
      .sort((a, b) => b.relevanceScore - a.relevanceScore);

    return {
      combined,
      movies: moviesWithType,
      tv: tvWithType,
      totalResults: combined.length
    };
  }, [searchResults, debouncedQuery]);

  const handleGoBack = () => {
    navigate(-1);
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      setSearchParams({ q: query.trim() });
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-40">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center gap-4 mb-4">
            <Button variant="ghost" onClick={handleGoBack}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
            
            <form onSubmit={handleSearch} className="flex-1 max-w-2xl">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search for movies and TV shows..."
                  className="pl-10 h-10"
                  autoFocus
                />
              </div>
            </form>
          </div>

          {query.trim() && (
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span>Search results for "{query.trim()}"</span>
              {processedResults.totalResults > 0 && (
                <span>({processedResults.totalResults} results)</span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="container mx-auto px-6 py-8">
        {!query.trim() && (
          <div className="text-center py-12">
            <div className="p-4 bg-muted rounded-full w-fit mx-auto mb-4">
              <Search className="w-8 h-8 text-muted-foreground" />
            </div>
            <h2 className="text-xl font-semibold mb-2">Search for Movies and TV Shows</h2>
            <p className="text-muted-foreground">
              Enter a title above to find movies and TV shows to request
            </p>
          </div>
        )}

        {query.trim() && (isLoading || (query.trim() !== debouncedQuery)) && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin mr-3" />
            <span className="text-lg">
              {query.trim() !== debouncedQuery ? "Typing..." : "Searching..."}
            </span>
          </div>
        )}

        {query.trim() && error && (
          <div className="text-center py-12">
            <div className="p-4 bg-destructive/10 rounded-full w-fit mx-auto mb-4">
              <Search className="w-8 h-8 text-destructive" />
            </div>
            <h2 className="text-xl font-semibold mb-2 text-destructive">Search Error</h2>
            <p className="text-muted-foreground">
              Something went wrong while searching. Please try again.
            </p>
          </div>
        )}

        {query.trim() && !isLoading && !error && processedResults.totalResults === 0 && query.trim() === debouncedQuery && (
          <div className="text-center py-12">
            <div className="p-4 bg-muted rounded-full w-fit mx-auto mb-4">
              <Search className="w-8 h-8 text-muted-foreground" />
            </div>
            <h2 className="text-xl font-semibold mb-2">No Results Found</h2>
            <p className="text-muted-foreground">
              No movies or TV shows found for "{query.trim()}". Try a different search term.
            </p>
          </div>
        )}

        {query.trim() && !isLoading && !error && processedResults.totalResults > 0 && query.trim() === debouncedQuery && (
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-3 mb-8">
              <TabsTrigger value="all" className="flex items-center gap-2">
                <Search className="w-4 h-4" />
                All ({processedResults.totalResults})
              </TabsTrigger>
              <TabsTrigger value="movies" className="flex items-center gap-2">
                <Film className="w-4 h-4" />
                Movies ({processedResults.movies.length})
              </TabsTrigger>
              <TabsTrigger value="tv" className="flex items-center gap-2">
                <Tv className="w-4 h-4" />
                TV Shows ({processedResults.tv.length})
              </TabsTrigger>
            </TabsList>

            <TabsContent value="all">
              <ContentGrid
                title="Search Results"
                data={processedResults.combined}
                isLoading={false}
                error={null}
                onRequest={handleRequest}
                isRequestLoading={isRequestLoading}
              />
            </TabsContent>

            <TabsContent value="movies">
              <ContentGrid
                title="Movies"
                data={processedResults.movies}
                isLoading={false}
                error={null}
                onRequest={handleRequest}
                isRequestLoading={isRequestLoading}
              />
            </TabsContent>

            <TabsContent value="tv">
              <ContentGrid
                title="TV Shows"
                data={processedResults.tv}
                isLoading={false}
                error={null}
                onRequest={handleRequest}
                isRequestLoading={isRequestLoading}
              />
            </TabsContent>
          </Tabs>
        )}
      </div>

      {/* Dialog Components */}
      <OnBehalfRequestDialog
        open={showOnBehalfDialog}
        onOpenChange={setShowOnBehalfDialog}
        selectedMedia={selectedMedia}
        selectedUser={selectedUser}
        onUserChange={setSelectedUser}
        allUsers={allUsers}
        onSubmit={handleOnBehalfSubmit}
        isLoading={isRequestLoading}
      />

      <SeasonSelectionDialog
        open={showSeasonDialog}
        onOpenChange={setShowSeasonDialog}
        selectedMedia={selectedMedia}
        selectedSeasons={selectedSeasons}
        onSeasonsChange={setSelectedSeasons}
        onSubmit={handleSeasonRequestSubmit}
        isLoading={isRequestLoading}
      />
    </div>
  );
}