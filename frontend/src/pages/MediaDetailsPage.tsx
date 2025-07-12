import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ArrowLeft, Star, Calendar, Clock, Film, Tv, Plus, CheckCircle } from "lucide-react";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import Loading from "@/components/Loading";
import { discoverApi } from "@/lib/api";
import { useState, useMemo } from "react";

export function MediaDetailsPage() {
  const { tmdb_id } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [isRequested, setIsRequested] = useState(false);

  // Get media type from URL params, default to movie if not specified
  const mediaType = (searchParams.get('type') as 'movie' | 'tv') || 'movie';

  // Mock library status - replace with actual API call later
  const isInLibrary = useMemo(() => Math.random() > 0.7, []);

  const { data: mediaDetails, isLoading, isError } = useQuery({
    queryKey: ["mediaDetails", tmdb_id, mediaType],
    queryFn: async () => {
      if (!tmdb_id) throw new Error("No TMDB ID provided");
      
      try {
        // Try the specified media type first
        return await discoverApi.getMediaDetails(tmdb_id, mediaType);
      } catch (error) {
        // If that fails and we tried movie, try TV show
        if (mediaType === 'movie') {
          try {
            return await discoverApi.getMediaDetails(tmdb_id, 'tv');
          } catch {
            throw new Error("Media not found");
          }
        }
        // If we tried TV show and it failed, try movie
        if (mediaType === 'tv') {
          try {
            return await discoverApi.getMediaDetails(tmdb_id, 'movie');
          } catch {
            throw new Error("Media not found");
          }
        }
        throw error;
      }
    },
    enabled: !!tmdb_id,
  });

  // Get recommendations (only for movies for now)
  const { data: recommendations } = useQuery({
    queryKey: ["movieRecommendations", tmdb_id],
    queryFn: () => discoverApi.getMovieRecommendations(tmdb_id!),
    enabled: !!tmdb_id && mediaType === 'movie',
  });

  // Get watch providers (only for movies for now)  
  const { data: watchProviders } = useQuery({
    queryKey: ["movieWatchProviders", tmdb_id],
    queryFn: () => discoverApi.getMovieWatchProviders(tmdb_id!),
    enabled: !!tmdb_id && mediaType === 'movie',
  });

  const handleRequest = () => {
    setIsRequested(true);
    // TODO: Implement actual request functionality
    alert(`Request submitted for: ${mediaDetails?.title || mediaDetails?.name}`);
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading />
      </div>
    );
  }

  if (isError || !mediaDetails) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-500 mb-4">Media Not Found</h1>
          <Button onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  const isMovie = mediaDetails.media_type === 'movie';
  const title = mediaDetails.title || mediaDetails.name;
  const releaseDate = mediaDetails.release_date || mediaDetails.first_air_date;
  const year = releaseDate ? new Date(releaseDate).getFullYear() : null;
  const posterUrl = mediaDetails.poster_path 
    ? `https://image.tmdb.org/t/p/w500${mediaDetails.poster_path}`
    : null;
  const backdropUrl = mediaDetails.backdrop_path
    ? `https://image.tmdb.org/t/p/w1280${mediaDetails.backdrop_path}`
    : null;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50">
        <div className="container mx-auto px-6 py-4">
          <Button variant="ghost" onClick={handleGoBack} className="mb-2">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Requests
          </Button>
        </div>
      </div>

      {/* Hero Section with Backdrop */}
      {backdropUrl && (
        <div 
          className="relative h-64 md:h-80 bg-cover bg-center"
          style={{ backgroundImage: `url(${backdropUrl})` }}
        >
          <div className="absolute inset-0 bg-black/60" />
          <div className="absolute bottom-6 left-6 right-6 text-white">
            <h1 className="text-3xl md:text-4xl font-bold mb-2">{title}</h1>
            {year && (
              <p className="text-lg opacity-90">{year}</p>
            )}
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="container mx-auto px-6 py-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {/* Poster */}
          <div className="md:col-span-1">
            <Card className="overflow-hidden">
              <CardContent className="p-0">
                {posterUrl ? (
                  <img 
                    src={posterUrl} 
                    alt={title}
                    className="w-full h-auto aspect-[2/3] object-cover"
                  />
                ) : (
                  <div className="w-full aspect-[2/3] bg-muted flex items-center justify-center">
                    {isMovie ? (
                      <Film className="w-16 h-16 text-muted-foreground" />
                    ) : (
                      <Tv className="w-16 h-16 text-muted-foreground" />
                    )}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Action Buttons */}
            <div className="mt-4 space-y-2">
              {isInLibrary ? (
                <Button disabled className="w-full">
                  <CheckCircle className="w-4 h-4 mr-2" />
                  In Library
                </Button>
              ) : isRequested ? (
                <Button disabled variant="secondary" className="w-full">
                  <Clock className="w-4 h-4 mr-2" />
                  Requested
                </Button>
              ) : (
                <Button onClick={handleRequest} className="w-full">
                  <Plus className="w-4 h-4 mr-2" />
                  Request
                </Button>
              )}
            </div>
          </div>

          {/* Details */}
          <div className="md:col-span-2 space-y-6">
            {/* Title and Year (for mobile when no backdrop) */}
            {!backdropUrl && (
              <div>
                <h1 className="text-3xl font-bold mb-2">{title}</h1>
                {year && <p className="text-lg text-muted-foreground">{year}</p>}
              </div>
            )}

            {/* Status Badges */}
            <div className="flex gap-2 flex-wrap">
              <Badge variant="outline">
                {isMovie ? (
                  <>
                    <Film className="w-3 h-3 mr-1" />
                    Movie
                  </>
                ) : (
                  <>
                    <Tv className="w-3 h-3 mr-1" />
                    TV Series
                  </>
                )}
              </Badge>
              
              {mediaDetails.vote_average && (
                <Badge variant="outline">
                  <Star className="w-3 h-3 mr-1 fill-yellow-400 text-yellow-400" />
                  {mediaDetails.vote_average.toFixed(1)}
                </Badge>
              )}

              {releaseDate && (
                <Badge variant="outline">
                  <Calendar className="w-3 h-3 mr-1" />
                  {new Date(releaseDate).toLocaleDateString()}
                </Badge>
              )}

              {mediaDetails.runtime && (
                <Badge variant="outline">
                  <Clock className="w-3 h-3 mr-1" />
                  {mediaDetails.runtime} min
                </Badge>
              )}

              {mediaDetails.number_of_seasons && (
                <Badge variant="outline">
                  {mediaDetails.number_of_seasons} Season{mediaDetails.number_of_seasons !== 1 ? 's' : ''}
                </Badge>
              )}
            </div>

            {/* Overview */}
            {mediaDetails.overview && (
              <div>
                <h2 className="text-xl font-semibold mb-3">Overview</h2>
                <p className="text-muted-foreground leading-relaxed">
                  {mediaDetails.overview}
                </p>
              </div>
            )}

            {/* Genres */}
            {mediaDetails.genres && mediaDetails.genres.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-3">Genres</h2>
                <div className="flex gap-2 flex-wrap">
                  {mediaDetails.genres.map((genre: { id: number; name: string }) => (
                    <Badge key={genre.id} variant="secondary">
                      {genre.name}
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Additional Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {mediaDetails.original_language && (
                <div>
                  <h3 className="font-medium text-sm text-muted-foreground mb-1">Original Language</h3>
                  <p className="text-sm">{mediaDetails.original_language.toUpperCase()}</p>
                </div>
              )}

              {mediaDetails.status && (
                <div>
                  <h3 className="font-medium text-sm text-muted-foreground mb-1">Status</h3>
                  <p className="text-sm">{mediaDetails.status}</p>
                </div>
              )}

              {mediaDetails.budget && mediaDetails.budget > 0 && (
                <div>
                  <h3 className="font-medium text-sm text-muted-foreground mb-1">Budget</h3>
                  <p className="text-sm">${mediaDetails.budget.toLocaleString()}</p>
                </div>
              )}

              {mediaDetails.revenue && mediaDetails.revenue > 0 && (
                <div>
                  <h3 className="font-medium text-sm text-muted-foreground mb-1">Revenue</h3>
                  <p className="text-sm">${mediaDetails.revenue.toLocaleString()}</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Watch Providers - US Only */}
        {watchProviders && watchProviders.results && watchProviders.results.US && (
          <div className="mt-12">
            <h2 className="text-2xl font-semibold mb-6">Where to Watch</h2>
            
            <div className="space-y-6">
              {/* Streaming providers */}
              {watchProviders.results.US.flatrate && watchProviders.results.US.flatrate.length > 0 && (
                <div>
                  <h3 className="text-lg font-medium mb-3">Stream</h3>
                  <div className="flex flex-wrap gap-3">
                    {watchProviders.results.US.flatrate.map((provider: { provider_id: number; provider_name: string; logo_path: string }) => (
                      <div 
                        key={provider.provider_id} 
                        className="group relative"
                        title={provider.provider_name}
                      >
                        <img 
                          src={`https://image.tmdb.org/t/p/w92${provider.logo_path}`}
                          alt={provider.provider_name}
                          className="w-12 h-12 rounded-lg transition-transform duration-200 hover:scale-105 cursor-pointer"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Buy providers */}
              {watchProviders.results.US.buy && watchProviders.results.US.buy.length > 0 && (
                <div>
                  <h3 className="text-lg font-medium mb-3">Buy</h3>
                  <div className="flex flex-wrap gap-3">
                    {watchProviders.results.US.buy.slice(0, 8).map((provider: { provider_id: number; provider_name: string; logo_path: string }) => (
                      <div 
                        key={provider.provider_id} 
                        className="group relative"
                        title={provider.provider_name}
                      >
                        <img 
                          src={`https://image.tmdb.org/t/p/w92${provider.logo_path}`}
                          alt={provider.provider_name}
                          className="w-12 h-12 rounded-lg transition-transform duration-200 hover:scale-105 cursor-pointer"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Rent providers */}
              {watchProviders.results.US.rent && watchProviders.results.US.rent.length > 0 && (
                <div>
                  <h3 className="text-lg font-medium mb-3">Rent</h3>
                  <div className="flex flex-wrap gap-3">
                    {watchProviders.results.US.rent.slice(0, 8).map((provider: { provider_id: number; provider_name: string; logo_path: string }) => (
                      <div 
                        key={provider.provider_id} 
                        className="group relative"
                        title={provider.provider_name}
                      >
                        <img 
                          src={`https://image.tmdb.org/t/p/w92${provider.logo_path}`}
                          alt={provider.provider_name}
                          className="w-12 h-12 rounded-lg transition-transform duration-200 hover:scale-105 cursor-pointer"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Recommendations Carousel */}
        {recommendations && recommendations.results && recommendations.results.length > 0 && (
          <div className="mt-12">
            <h2 className="text-2xl font-semibold mb-6">Recommended Movies</h2>
            
            <Carousel
              opts={{
                align: "start",
                slidesToScroll: "auto",
                containScroll: "trimSnaps",
              }}
              className="w-full max-w-full"
            >
              <CarouselContent className="-ml-2 md:-ml-4">
                {recommendations.results.map((movie: { id: number; title: string; poster_path: string; vote_average?: number }) => (
                  <CarouselItem key={movie.id} className="pl-2 md:pl-4 basis-auto">
                    <Card 
                      className="group cursor-pointer transition-all duration-300 hover:shadow-lg hover:scale-105 w-40"
                      onClick={() => navigate(`/requests/${movie.id}/details?type=movie`)}
                    >
                      <CardContent className="p-0">
                        <div className="relative aspect-[2/3] bg-muted rounded-lg overflow-hidden">
                          {movie.poster_path ? (
                            <img
                              src={`https://image.tmdb.org/t/p/w342${movie.poster_path}`}
                              alt={movie.title}
                              className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
                            />
                          ) : (
                            <div className="w-full h-full flex items-center justify-center">
                              <Film className="w-8 h-8 text-muted-foreground" />
                            </div>
                          )}
                          
                          {/* Hover overlay */}
                          <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-3">
                            <h3 className="text-white text-sm font-medium line-clamp-2 mb-1">
                              {movie.title}
                            </h3>
                            {movie.vote_average && movie.vote_average > 0 && (
                              <div className="flex items-center gap-1 text-yellow-400 text-xs">
                                <Star className="w-3 h-3 fill-current" />
                                <span>{movie.vote_average.toFixed(1)}</span>
                              </div>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </CarouselItem>
                ))}
              </CarouselContent>
              <CarouselPrevious className="hidden sm:flex -left-4 lg:-left-12" />
              <CarouselNext className="hidden sm:flex -right-4 lg:-right-12" />
            </Carousel>
          </div>
        )}
      </div>
    </div>
  );
}