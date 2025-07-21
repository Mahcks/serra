import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import {
  ArrowLeft,
  Star,
  ExternalLink,
  Calendar,
  Film,
  Tv,
  HardDrive,
  MonitorPlay,
  ThumbsUp,
  Users,
  Loader2,
  Folder,
  Plus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { MediaCard } from "@/components/ui/media-card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { discoverApi, requestsApi, backendApi } from "@/lib/api";
import Loading from "@/components/Loading";
import type {
  MovieDetails,
  TVDetails,
  TMDBMediaItem,
  CreateRequestRequest,
  UserWithPermissions,
  CastMember,
  PermissionInfo,
  Season,
  SeasonDetails,
  Episode,
  MediaRatingsResponse,
} from "@/types";
import { useAuth } from "@/lib/auth";
import { useState, useCallback, useMemo } from "react";
import { handleApiError, ERROR_CODES, getErrorCode } from "@/utils/errorHandling";
import Logo from "@/components/logos";
import MediaCarousel from "@/components/ui/media-carousel";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

interface ReleaseDateInfo {
  certification: string;
  release_date: string;
  type: number;
  note?: string;
}

export default function MediaDetailsPage() {
  const { tmdb_id, media_type } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [currentRequestItem, setCurrentRequestItem] =
    useState<TMDBMediaItem | null>(null);

  // On-behalf request state
  const [showOnBehalfDialog, setShowOnBehalfDialog] = useState(false);
  const [selectedMedia, setSelectedMedia] = useState<TMDBMediaItem | null>(
    null
  );
  const [selectedUser, setSelectedUser] = useState<string>("");

  // Season selection state
  const [showSeasonDialog, setShowSeasonDialog] = useState(false);
  const [selectedSeasons, setSelectedSeasons] = useState<number[]>([]);

  // Season details state
  const [seasonDetails, setSeasonDetails] = useState<
    Record<number, SeasonDetails>
  >({});
  const [loadingSeasons, setLoadingSeasons] = useState<Record<number, boolean>>(
    {}
  );

  // Get media type from URL params, validate it
  const mediaType =
    media_type === "movie" || media_type === "tv" ? media_type : "movie";

  const {
    data: mediaDetails,
    isLoading,
    isError,
  } = useQuery<MovieDetails | TVDetails>({
    queryKey: ["mediaDetails", tmdb_id, mediaType],
    queryFn: async () => {
      if (!tmdb_id) throw new Error("No TMDB ID provided");

      try {
        // Try the specified media type first
        return await discoverApi.getMediaDetails(tmdb_id, mediaType);
      } catch (error) {
        // If that fails and we tried movie, try TV show
        if (mediaType === "movie") {
          try {
            return await discoverApi.getMediaDetails(tmdb_id, "tv");
          } catch {
            throw new Error("Media not found");
          }
        }
        // If we tried TV show and it failed, try movie
        if (mediaType === "tv") {
          try {
            return await discoverApi.getMediaDetails(tmdb_id, "movie");
          } catch {
            throw new Error("Media not found");
          }
        }
        throw error;
      }
    },
    enabled: !!tmdb_id,
  });

  // Query season availability for TV shows - always call hook to maintain order
  const {
    data: seasonAvailability,
    isLoading: isLoadingAvailability,
    refetch: refetchAvailability,
  } = useQuery({
    queryKey: ["seasonAvailability", tmdb_id, mediaType],
    queryFn: () => {
      if (mediaType !== "tv" || !tmdb_id) {
        return Promise.resolve(null);
      }
      return discoverApi.getSeasonAvailability(Number(tmdb_id));
    },
    enabled: !!tmdb_id,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Get US release dates for movies
  const { data: releaseDates } = useQuery<ReleaseDateInfo[]>({
    queryKey: ["movieReleaseDates", tmdb_id],
    queryFn: async () => {
      if (!tmdb_id) throw new Error("No TMDB ID provided");
      const result = await discoverApi.getMovieReleaseDates(tmdb_id);
      console.log("🎬 Full release dates result:", result);
      // Filter for US release dates only
      const usReleases = result.results?.find(
        (country: { iso_3166_1: string }) => country.iso_3166_1 === "US"
      );
      console.log("🇺🇸 US release dates:", usReleases?.release_dates);
      return usReleases?.release_dates || [];
    },
    enabled: !!tmdb_id && mediaType === "movie",
  });

  // Get recommendations
  const { data: recommendations, isLoading: recommendationsLoading } = useQuery(
    {
      queryKey: ["recommendations", tmdb_id, mediaType],
      queryFn: async () => {
        if (!tmdb_id) throw new Error("No TMDB ID provided");
        if (mediaType === "movie") {
          return await discoverApi.getMovieRecommendations(tmdb_id);
        } else {
          return await discoverApi.getTVRecommendations(tmdb_id);
        }
      },
      enabled: !!tmdb_id,
    }
  );

  // Get similar content
  const { data: similar, isLoading: similarLoading } = useQuery({
    queryKey: ["similar", tmdb_id, mediaType],
    queryFn: async () => {
      if (!tmdb_id) throw new Error("No TMDB ID provided");
      if (mediaType === "movie") {
        return await discoverApi.getSimilarMovies(tmdb_id);
      } else {
        return await discoverApi.getSimilarTV(tmdb_id);
      }
    },
    enabled: !!tmdb_id,
  });

  // Get media ratings
  const { data: ratings, isLoading: ratingsLoading } =
    useQuery<MediaRatingsResponse>({
      queryKey: ["ratings", tmdb_id, mediaType],
      queryFn: async () => {
        if (!tmdb_id || !mediaDetails)
          throw new Error("No TMDB ID or media details provided");

        const title = isMovie
          ? (mediaDetails as MovieDetails).title
          : (mediaDetails as TVDetails).name;
        const releaseDate = isMovie
          ? (mediaDetails as MovieDetails).release_date
          : (mediaDetails as TVDetails).first_air_date;
        const year = releaseDate
          ? new Date(releaseDate).getFullYear()
          : undefined;

        return await discoverApi.getMediaRatings(
          tmdb_id,
          mediaType,
          title,
          year
        );
      },
      enabled: !!tmdb_id && !!mediaDetails,
      retry: false, // Don't retry on failure since ratings might not be available
      staleTime: 30 * 60 * 1000, // 30 minutes
    });

  // Fetch current user's detailed permissions
  const { data: currentUserPermissions } = useQuery({
    queryKey: ["current-user-permissions"],
    queryFn: backendApi.getCurrentUserPermissions,
    enabled: !!user,
  });

  // Check if user can request on behalf of others
  const canRequestOnBehalf = useMemo(() => {
    if (!user) return false;

    // Admin users can always request on behalf of others
    if (user.is_admin) return true;

    // Check if user has owner or requests.manage permission
    const userPermissions = currentUserPermissions?.permissions || [];
    return userPermissions.some(
      (perm: PermissionInfo) =>
        perm.id === "owner" || perm.id === "requests.manage"
    );
  }, [user, currentUserPermissions]);

  // Fetch all users for on-behalf requests (only if user has permission)
  const { data: allUsers } = useQuery<{ users: UserWithPermissions[] }>({
    queryKey: ["all-users"],
    queryFn: backendApi.getUsers,
    enabled: canRequestOnBehalf,
  });

  // Create request mutation
  const createRequestMutation = useMutation({
    mutationFn: (data: CreateRequestRequest) => {
      return requestsApi.createRequest(data);
    },
    onSuccess: (newRequest) => {
      const displayTitle =
        newRequest.title ||
        currentRequestItem?.title ||
        currentRequestItem?.name ||
        "the requested content";

      if (newRequest.status === "approved") {
        toast.success(`🎉 Request Approved!`, {
          description: `"${displayTitle}" was automatically approved and will be downloaded soon.`,
          duration: 5000,
        });
      } else if (newRequest.status === "fulfilled") {
        toast.success(`✅ Request Fulfilled!`, {
          description: `"${displayTitle}" is already available in your library.`,
          duration: 4000,
        });
      } else {
        toast.success(`📝 Request Submitted!`, {
          description: `Your request for "${displayTitle}" has been submitted for review.`,
          duration: 4000,
        });
      }

      // Invalidate recommendation queries to update request status
      queryClient.invalidateQueries({
        queryKey: ["recommendations", tmdb_id, mediaType],
      });
      queryClient.invalidateQueries({
        queryKey: ["similar", tmdb_id, mediaType],
      });

      setCurrentRequestItem(null);
    },
    onError: (error: unknown) => {
      console.error("Request creation failed:", error);
      
      const { message } = handleApiError(error);
      const errorCode = getErrorCode(error);

      if (errorCode === ERROR_CODES.DUPLICATE_REQUEST) {
        toast.error(`🔄 Already Requested`, {
          description: `You've already requested this content.`,
          duration: 4000,
        });
      } else {
        const title = errorCode ? `❌ Request Failed (${errorCode})` : `❌ Request Failed`;
        toast.error(title, {
          description: message,
          duration: 4000,
        });
      }

      setCurrentRequestItem(null);
    },
  });

  const handleRequest = useCallback(
    (item: TMDBMediaItem) => {
      if (!user) return;

      // If user can request on behalf of others, show dialog to choose
      if (canRequestOnBehalf && allUsers?.users && allUsers.users.length > 0) {
        setSelectedMedia(item);
        setSelectedUser("myself"); // Default to myself
        setShowOnBehalfDialog(true);
        return;
      }

      // Otherwise create request directly
      const mediaType =
        item.media_type || (item.first_air_date ? "tv" : "movie");
      const title = item.title || item.name || "Unknown Title";
      const posterUrl = item.poster_path
        ? `https://image.tmdb.org/t/p/w500${item.poster_path}`
        : undefined;

      const requestData = {
        media_type: mediaType,
        tmdb_id: item.id,
        title: title,
        poster_url: posterUrl,
        // Only include seasons for TV shows and when seasons are selected
        ...(mediaType === "tv" &&
          selectedSeasons.length > 0 && { seasons: selectedSeasons }),
      };

      setCurrentRequestItem(item);
      createRequestMutation.mutate(requestData);
    },
    [createRequestMutation, user, canRequestOnBehalf, allUsers, selectedSeasons]
  );

  // Handle on-behalf request submission
  const handleOnBehalfSubmit = () => {
    if (!selectedMedia) return;

    const mediaType =
      selectedMedia.media_type ||
      (selectedMedia.first_air_date ? "tv" : "movie");
    const title = selectedMedia.title || selectedMedia.name || "Unknown Title";
    const posterUrl = selectedMedia.poster_path
      ? `https://image.tmdb.org/t/p/w500${selectedMedia.poster_path}`
      : undefined;

    const requestData = {
      media_type: mediaType,
      tmdb_id: selectedMedia.id,
      title: title,
      poster_url: posterUrl,
      on_behalf_of:
        selectedUser && selectedUser !== "myself" ? selectedUser : undefined,
      // Only include seasons for TV shows and when seasons are selected
      ...(mediaType === "tv" &&
        selectedSeasons.length > 0 && { seasons: selectedSeasons }),
    };

    setCurrentRequestItem(selectedMedia);
    createRequestMutation.mutate(requestData);
    setShowOnBehalfDialog(false);
    setSelectedMedia(null);
    setSelectedUser("");
    setSelectedSeasons([]); // Clear seasons after request
  };

  // Collection data - extracted from belongs_to_collection field (movies only)
  const collectionData = useMemo(() => {
    if (!mediaDetails || mediaType !== "movie") {
      return null;
    }

    const movieData = mediaDetails as MovieDetails;
    if (!movieData.belongs_to_collection) {
      return null;
    }

    const collection = movieData.belongs_to_collection;
    if (
      typeof collection === "object" &&
      collection !== null &&
      "id" in collection
    ) {
      return collection as {
        id: number;
        name: string;
        poster_path?: string;
        backdrop_path?: string;
      };
    }

    return null;
  }, [mediaDetails, mediaType]);

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
          <h1 className="text-2xl font-bold text-destructive mb-4">
            Media Not Found
          </h1>
          <Button onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  const isMovie = mediaType === "movie";
  const movieData = mediaDetails as MovieDetails;
  const tvData = mediaDetails as TVDetails;

  const title = isMovie ? movieData.title : tvData.name;
  const originalTitle = isMovie
    ? movieData.original_title
    : tvData.original_name;
  const releaseDate = isMovie ? movieData.release_date : tvData.first_air_date;

  console.log("🎬 Debug info:", {
    mediaType,
    isMovie,
    tmdb_id,
    hasReleaseDate: !!releaseDate,
    releaseDatesCount: releaseDates?.length,
    releaseDates: releaseDates,
  });
  const lastAirDate = !isMovie ? tvData.last_air_date : undefined;
  const tagline = mediaDetails.tagline;
  const overview = mediaDetails.overview;
  const genres = mediaDetails.genres || [];
  const productionCompanies = mediaDetails.production_companies || [];
  const productionCountries = mediaDetails.production_countries || [];
  const spokenLanguages = mediaDetails.spoken_languages || [];
  const originalLanguage = mediaDetails.original_language;
  const voteAverage = mediaDetails.vote_average;
  const voteCount = mediaDetails.vote_count;
  const status = mediaDetails.status;
  const runtime = isMovie ? movieData.runtime : undefined;
  const numberOfSeasons = !isMovie ? tvData.number_of_seasons : undefined;
  const numberOfEpisodes = !isMovie ? tvData.number_of_episodes : undefined;
  const episodeRunTime = !isMovie ? tvData.episode_run_time?.[0] : undefined;
  const imdbId = isMovie ? movieData.imdb_id : undefined;
  const homepage = mediaDetails.homepage;
  const budget = isMovie ? movieData.budget : undefined;
  const revenue = isMovie ? movieData.revenue : undefined;
  const inProduction = !isMovie ? tvData.in_production : undefined;
  const networks = !isMovie ? tvData.networks : undefined;
  const createdBy = !isMovie ? tvData.created_by : undefined;

  const credits = mediaDetails.credits;
  const cast = credits?.cast || [];
  const crew = credits?.crew || [];

  // Find key crew members
  const directors = crew.filter((member) => member.job === "Director");
  const writers = crew.filter(
    (member) =>
      member.job === "Writer" ||
      member.job === "Screenplay" ||
      member.job === "Story" ||
      member.department === "Writing"
  );
  const editors = crew.filter((member) => member.job === "Editor");
  const producers = crew.filter(
    (member) => member.job === "Producer" || member.job === "Executive Producer"
  );
  const cinematographers = crew.filter(
    (member) =>
      member.job === "Director of Photography" ||
      member.job === "Cinematography"
  );

  const year = releaseDate ? new Date(releaseDate).getFullYear() : null;
  const lastYear = lastAirDate ? new Date(lastAirDate).getFullYear() : null;

  const posterUrl = mediaDetails.poster_path
    ? `https://image.tmdb.org/t/p/w500${mediaDetails.poster_path}`
    : null;

  const backdropUrl = mediaDetails.backdrop_path
    ? `https://image.tmdb.org/t/p/w1280${mediaDetails.backdrop_path}`
    : null;

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat("en-US").format(num);
  };

  const getLanguageName = (code: string) => {
    const language = spokenLanguages.find((lang) => lang.iso_639_1 === code);
    return language?.english_name || code.toUpperCase();
  };

  const formatRuntime = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return hours > 0 ? `${hours}h ${mins}m` : `${mins}m`;
  };

  // Helper functions for release dates
  const getReleaseTypeIcon = (type: number) => {
    const iconMap: Record<
      number,
      React.ComponentType<{ className?: string }>
    > = {
      1: Film, // Premiere
      2: Film, // Theatrical (Limited)
      3: Film, // Theatrical
      4: MonitorPlay, // Digital
      5: HardDrive, // Physical
      6: Tv, // TV
    };
    return iconMap[type] || Calendar;
  };

  const getReleaseTypeName = (type: number): string => {
    const typeMap: Record<number, string> = {
      1: "Premiere",
      2: "Limited",
      3: "Theatrical",
      4: "Digital",
      5: "Physical",
      6: "TV",
    };
    return typeMap[type] || "Unknown";
  };

  const formatReleaseDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  // Handle main media request button
  const handleMainRequest = () => {
    if (!mediaDetails || !user) return;

    // For TV series, show season selection dialog first
    if (mediaType === "tv" && tvData.seasons && tvData.seasons.length > 0) {
      setShowSeasonDialog(true);
      setSelectedSeasons([]); // Reset selection
      return;
    }

    // For movies, go directly to request (with on-behalf dialog if applicable)
    const mediaItem: TMDBMediaItem = {
      id: mediaDetails.id,
      title: isMovie ? movieData.title : tvData.name,
      name: isMovie ? movieData.title : tvData.name,
      media_type: mediaType,
      poster_path: mediaDetails.poster_path,
      overview: mediaDetails.overview,
      vote_average: mediaDetails.vote_average,
      vote_count: mediaDetails.vote_count,
      popularity: mediaDetails.popularity,
      original_language: mediaDetails.original_language,
      genre_ids: mediaDetails.genres?.map((g) => g.id) || [],
      adult: mediaDetails.adult,
      backdrop_path: mediaDetails.backdrop_path,
      release_date: isMovie ? movieData.release_date : undefined,
      first_air_date: !isMovie ? tvData.first_air_date : undefined,
      original_title: isMovie ? movieData.original_title : undefined,
      original_name: !isMovie ? tvData.original_name : undefined,
      video: isMovie ? movieData.video : undefined,
      origin_country: !isMovie ? tvData.origin_country : undefined,
    };

    handleRequest(mediaItem);
  };

  // Handle season selection submission - goes to on-behalf dialog if applicable
  const handleSeasonRequestSubmit = () => {
    if (!mediaDetails || selectedSeasons.length === 0) return;

    // Create the media item for the TV series
    const mediaItem: TMDBMediaItem = {
      id: mediaDetails.id,
      title: tvData.name,
      name: tvData.name,
      media_type: "tv",
      poster_path: mediaDetails.poster_path,
      overview: mediaDetails.overview,
      vote_average: mediaDetails.vote_average,
      vote_count: mediaDetails.vote_count,
      popularity: mediaDetails.popularity,
      original_language: mediaDetails.original_language,
      genre_ids: mediaDetails.genres?.map((g) => g.id) || [],
      adult: mediaDetails.adult,
      backdrop_path: mediaDetails.backdrop_path,
      first_air_date: tvData.first_air_date,
      original_name: tvData.original_name,
      origin_country: tvData.origin_country,
    };

    // Close season dialog first
    setShowSeasonDialog(false);

    // If user can request on behalf of others, show dialog to choose
    if (canRequestOnBehalf && allUsers?.users && allUsers.users.length > 0) {
      setSelectedMedia(mediaItem);
      setSelectedUser("myself"); // Default to myself
      setShowOnBehalfDialog(true);
      return;
    }

    // Otherwise create request directly with selected seasons
    const requestData = {
      media_type: "tv" as const,
      tmdb_id: mediaItem.id,
      title: mediaItem.name,
      poster_url: mediaItem.poster_path
        ? `https://image.tmdb.org/t/p/w500${mediaItem.poster_path}`
        : undefined,
      seasons: selectedSeasons,
    };

    setCurrentRequestItem(mediaItem);
    createRequestMutation.mutate(requestData);
    setSelectedSeasons([]);
  };

  // Handle season selection toggle
  const handleSeasonToggle = (seasonNumber: number) => {
    setSelectedSeasons((prev) =>
      prev.includes(seasonNumber)
        ? prev.filter((s) => s !== seasonNumber)
        : [...prev, seasonNumber]
    );
  };

  // Select all seasons
  const handleSelectAllSeasons = () => {
    if (!tvData.seasons) return;
    const allSeasonNumbers = tvData.seasons
      .filter((s) => s.season_number > 0)
      .map((s) => s.season_number);
    setSelectedSeasons(allSeasonNumbers);
  };

  // Clear season selection
  const handleClearSeasonSelection = () => {
    setSelectedSeasons([]);
  };

  // Fetch season details when accordion is expanded
  const fetchSeasonDetails = async (seasonNumber: number) => {
    if (
      !tmdb_id ||
      seasonDetails[seasonNumber] ||
      loadingSeasons[seasonNumber]
    ) {
      return;
    }

    setLoadingSeasons((prev) => ({ ...prev, [seasonNumber]: true }));

    try {
      const details = await discoverApi.getSeasonDetails(
        Number(tmdb_id),
        seasonNumber
      );
      setSeasonDetails((prev) => ({ ...prev, [seasonNumber]: details }));
    } catch (error) {
      console.error("Failed to fetch season details:", error);
      toast.error("Failed to load episodes", {
        description: "Could not load episode details for this season.",
      });
    } finally {
      setLoadingSeasons((prev) => ({ ...prev, [seasonNumber]: false }));
    }
  };

  // Helper functions for media carousels
  const handleMediaItemClick = (item: TMDBMediaItem) => {
    const mediaType = item.media_type || (item.title ? "movie" : "tv");
    navigate(`/requests/${mediaType}/${item.id}/details`);
  };

  const renderMediaItem = (item: TMDBMediaItem) => {
    const title = item.title || item.name || "Unknown Title";

    const embyItem = {
      id: item.id.toString(),
      name: title,
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
    };

    return (
      <MediaCard
        item={embyItem}
        size="md"
        onClick={() => handleMediaItemClick(item)}
        onRequest={user ? () => handleRequest(item) : undefined}
        className="w-full"
      />
    );
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header with Back Button */}
      <div className="sticky top-0 z-50 bg-background/95 backdrop-blur-sm border-b">
        <div className="container mx-auto px-4 py-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleGoBack}
            className="gap-2"
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </Button>
        </div>
      </div>

      {/* Hero Section with Backdrop */}
      <div className="relative">
        {backdropUrl && (
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url(${backdropUrl})` }}
          >
            <div className="absolute inset-0 bg-gradient-to-t from-background via-background/80 to-background/40" />
          </div>
        )}

        <div className="relative container mx-auto px-4 py-8">
          <div className="flex flex-col lg:flex-row gap-8">
            {/* Poster */}
            {posterUrl && (
              <div className="flex-shrink-0">
                <img
                  src={posterUrl}
                  alt={title}
                  className="w-64 h-96 object-cover rounded-lg shadow-2xl border"
                />
              </div>
            )}

            {/* Main Info */}
            <div className="flex-1 space-y-6">
              {/* Title and Year */}
              <div>
                <h1 className="text-4xl font-bold mb-2">{title}</h1>
                {originalTitle && originalTitle !== title && (
                  <p className="text-xl text-muted-foreground mb-2">
                    {originalTitle}
                  </p>
                )}
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>
                    {year}
                    {lastYear && lastYear !== year ? ` - ${lastYear}` : ""}
                  </span>
                  {runtime && <span>{formatRuntime(runtime)}</span>}
                  {numberOfSeasons && (
                    <span>
                      {numberOfSeasons} Season{numberOfSeasons > 1 ? "s" : ""}
                    </span>
                  )}
                  {numberOfEpisodes && <span>{numberOfEpisodes} Episodes</span>}
                </div>
              </div>

              {/* Tagline */}
              {tagline && (
                <p className="text-lg italic text-muted-foreground border-l-4 border-primary pl-4">
                  "{tagline}"
                </p>
              )}

              {/* Genres */}
              <div className="flex flex-wrap gap-2">
                {genres.map((genre) => (
                  <Badge key={genre.id} variant="secondary">
                    {genre.name}
                  </Badge>
                ))}
              </div>

              {/* Collection Banner */}
              {collectionData && (
                <div
                  className="p-3 bg-gradient-to-r from-amber-500/10 to-orange-500/10 border border-amber-500/20 rounded-lg cursor-pointer hover:from-amber-500/15 hover:to-orange-500/15 transition-colors duration-200"
                  onClick={() => navigate(`/collection/${collectionData.id}`)}
                >
                  <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400">
                    <Folder className="w-4 h-4" />
                    <span className="text-sm font-medium">
                      Part of {collectionData.name}
                    </span>
                    <ArrowLeft className="w-4 h-4 rotate-180 ml-auto" />
                  </div>
                </div>
              )}

              {/* Rating and Stats */}
              <div className="flex items-center gap-6">
                <div className="flex items-center gap-2">
                  <Star className="w-5 h-5 fill-yellow-400 text-yellow-400" />
                  <span className="font-semibold">
                    {voteAverage.toFixed(1)}
                  </span>
                  <span className="text-sm text-muted-foreground">
                    ({formatNumber(voteCount)} votes)
                  </span>
                </div>
              </div>

              {/* Request Button - Only show if user can request */}
              {user && (
                <div className="flex items-center gap-4">
                  <Button
                    onClick={handleMainRequest}
                    disabled={createRequestMutation.isPending}
                    className="flex items-center gap-2"
                  >
                    {createRequestMutation.isPending ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Requesting...
                      </>
                    ) : (
                      <>
                        <Plus className="w-4 h-4" />
                        Request {isMovie ? "Movie" : "Series"}
                      </>
                    )}
                  </Button>
                </div>
              )}

              {/* Overview */}
              {overview && (
                <div>
                  <h3 className="text-lg font-semibold mb-2">Overview</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {overview}
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8 space-y-8">
        {/* Combined Layout: Key People (2/3) + Details Table (1/3) */}
        <div className="flex flex-col lg:flex-row gap-8">
          {/* Key People */}
          <div className="flex-1">
            {(directors.length > 0 || writers.length > 0 || createdBy) && (
              <Card>
                <CardContent className="p-6">
                  <h3 className="text-lg font-semibold mb-4">Key People</h3>
                  <div className="grid sm:grid-cols-2 xl:grid-cols-3 gap-4">
                    {!isMovie && createdBy && createdBy.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Created By
                        </h4>
                        {createdBy.map((creator) => (
                          <p key={creator.id} className="text-sm">
                            {creator.name}
                          </p>
                        ))}
                      </div>
                    )}
                    {directors.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Director{directors.length > 1 ? "s" : ""}
                        </h4>
                        {directors.map((director) => (
                          <p key={director.id} className="text-sm">
                            {director.name}
                          </p>
                        ))}
                      </div>
                    )}
                    {writers.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Writer{writers.length > 1 ? "s" : ""}
                        </h4>
                        {writers.slice(0, 3).map((writer) => (
                          <p key={writer.id} className="text-sm">
                            {writer.name}
                          </p>
                        ))}
                        {writers.length > 3 && (
                          <p className="text-xs text-muted-foreground">
                            +{writers.length - 3} more
                          </p>
                        )}
                      </div>
                    )}
                    {editors.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Editor{editors.length > 1 ? "s" : ""}
                        </h4>
                        {editors.slice(0, 2).map((editor) => (
                          <p key={editor.id} className="text-sm">
                            {editor.name}
                          </p>
                        ))}
                      </div>
                    )}
                    {cinematographers.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Cinematography
                        </h4>
                        {cinematographers.slice(0, 2).map((cinematographer) => (
                          <p key={cinematographer.id} className="text-sm">
                            {cinematographer.name}
                          </p>
                        ))}
                      </div>
                    )}
                    {producers.length > 0 && (
                      <div>
                        <h4 className="font-medium text-sm text-muted-foreground mb-2">
                          Producer{producers.length > 1 ? "s" : ""}
                        </h4>
                        {producers.slice(0, 2).map((producer) => (
                          <p key={producer.id} className="text-sm">
                            {producer.name}
                          </p>
                        ))}
                        {producers.length > 2 && (
                          <p className="text-xs text-muted-foreground">
                            +{producers.length - 2} more
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>

          {/* Details Table */}
          <div className="max-w-xs ml-auto">
            <Card className="p-0">
              <CardContent className="p-0">
                <div className="overflow-hidden">
                  <table className="w-full">
                    <tbody className="divide-y divide-border">
                      {/* Ratings Section */}
                      {(ratings?.rotten_tomatoes || voteAverage > 0) && (
                        <tr>
                          <td
                            className="px-4 py-2 text-xs font-medium text-muted-foreground"
                            colSpan={2}
                          >
                            <div className="flex justify-center gap-2 w-full">
                              {/* TMDB Rating */}
                              <div className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs transition-colors">
                                <Logo name="tmdb" size={24} />
                                <div className="flex flex-col items-center">
                                  <span className="text-xs font-medium">
                                    {Math.round(voteAverage * 10)}%
                                  </span>
                                </div>
                              </div>

                              {/* Rotten Tomatoes Critics Rating */}
                              {ratings?.rotten_tomatoes && (
                                <a
                                  href={ratings.rotten_tomatoes.url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs transition-colors"
                                  title="Critics Score"
                                >
                                  <Logo name="tomatometer" />
                                  <div className="flex flex-col items-center">
                                    <span className="text-xs font-medium">
                                      {ratings.rotten_tomatoes.tomato_meter}%
                                    </span>
                                  </div>
                                </a>
                              )}

                              {/* Rotten Tomatoes Audience Rating */}
                              {ratings?.rotten_tomatoes && (
                                <a
                                  href={ratings.rotten_tomatoes.url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="inline-flex items-center gap-1 px-2 rounded-md text-xs transition-colors"
                                  title="Audience Score"
                                >
                                  <Logo name="popcornmeter" size={24} />
                                  <div className="flex flex-col items-center">
                                    <span className="text-xs font-medium">
                                      {ratings.rotten_tomatoes.audience_score}%
                                    </span>
                                  </div>
                                </a>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                      {status && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Status
                          </td>
                          <td className="px-4 py-2 text-xs">{status}</td>
                        </tr>
                      )}
                      <tr>
                        <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                          Release
                        </td>
                        <td className="px-4 py-2 text-xs">
                          {isMovie &&
                          releaseDates &&
                          releaseDates.length > 0 ? (
                            <div className="space-y-1">
                              {releaseDates
                                .sort(
                                  (a: ReleaseDateInfo, b: ReleaseDateInfo) =>
                                    new Date(a.release_date).getTime() -
                                    new Date(b.release_date).getTime()
                                )
                                .map(
                                  (release: ReleaseDateInfo, index: number) => {
                                    const IconComponent = getReleaseTypeIcon(
                                      release.type
                                    );
                                    return (
                                      <div
                                        key={index}
                                        className="flex items-center gap-2"
                                      >
                                        <IconComponent className="w-3 h-3 text-muted-foreground" />
                                        <span className="text-xs font-medium">
                                          {getReleaseTypeName(release.type)}:
                                        </span>
                                        <span className="text-xs">
                                          {formatReleaseDate(
                                            release.release_date
                                          )}
                                        </span>
                                      </div>
                                    );
                                  }
                                )}
                            </div>
                          ) : releaseDate ? (
                            new Date(releaseDate).toLocaleDateString()
                          ) : (
                            "N/A"
                          )}
                        </td>
                      </tr>
                      {lastAirDate && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Last Air
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {new Date(lastAirDate).toLocaleDateString()}
                          </td>
                        </tr>
                      )}
                      <tr>
                        <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                          Language
                        </td>
                        <td className="px-4 py-2 text-xs">
                          {getLanguageName(originalLanguage)}
                        </td>
                      </tr>
                      {runtime && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Runtime
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {formatRuntime(runtime)}
                          </td>
                        </tr>
                      )}
                      {episodeRunTime && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Episode
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {formatRuntime(episodeRunTime)}
                          </td>
                        </tr>
                      )}
                      {numberOfSeasons && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Seasons
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {numberOfSeasons}
                          </td>
                        </tr>
                      )}
                      {numberOfEpisodes && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Episodes
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {numberOfEpisodes}
                          </td>
                        </tr>
                      )}
                      {productionCompanies.length > 0 && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Studio
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {productionCompanies[0].name}
                            {productionCompanies.length > 1 &&
                              ` +${productionCompanies.length - 1}`}
                          </td>
                        </tr>
                      )}
                      {networks && networks.length > 0 && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Network
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {networks[0].name}
                          </td>
                        </tr>
                      )}
                      {isMovie && Number(budget) > 0 && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Budget
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {formatCurrency(Number(budget))}
                          </td>
                        </tr>
                      )}
                      {isMovie && Number(revenue) > 0 && (
                        <tr>
                          <td className="px-4 py-2 text-xs font-medium text-muted-foreground">
                            Revenue
                          </td>
                          <td className="px-4 py-2 text-xs">
                            {formatCurrency(revenue)}
                          </td>
                        </tr>
                      )}
                      <tr>
                        <td
                          className="px-4 py-2 text-xs font-medium text-muted-foreground"
                          colSpan={2}
                        >
                          <div className="flex justify-center gap-2 w-full">
                            {imdbId && (
                              <a
                                href={`https://www.imdb.com/title/${imdbId}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1 px-2 py-1 hover:border-primary/40 border border-transparent rounded-md transition-all duration-200 hover:scale-105"
                              >
                                <Logo name="imdb" size={30} />
                              </a>
                            )}
                            <a
                              href={`https://www.themoviedb.org/${
                                isMovie ? "movie" : "tv"
                              }/${tmdb_id}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="inline-flex items-center gap-1 px-2 py-1 hover:border-primary/40 border border-transparent rounded-md text-xs transition-all duration-200 hover:scale-105"
                            >
                              <Logo name="tmdb" size={30} />
                            </a>
                            {isMovie && (
                              <a
                                href={`https://letterboxd.com/tmdb/${tmdb_id}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1 px-2 py-1  hover:border-primary/40 border border-transparent rounded-md text-xs transition-all duration-200 hover:scale-105"
                              >
                                <Logo name="letterboxd" size={30} />
                              </a>
                            )}
                            {homepage?.startsWith("http") && (
                              <a
                                href={homepage}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1 px-2 py-1  hover:border-primary/40 border border-transparent rounded-md text-xs transition-all duration-200 hover:scale-105"
                              >
                                Website
                                <ExternalLink className="w-3 h-3" />
                              </a>
                            )}
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Seasons and Episodes Accordion - Only for TV Shows */}
        {mediaType === "tv" && tvData.seasons && tvData.seasons.length > 0 && (
          <Card>
            <CardContent className="p-6">
              <h3 className="text-lg font-semibold mb-4">Seasons & Episodes</h3>
              <Accordion type="multiple" className="w-full">
                {tvData.seasons
                  .filter((season) => season.season_number > 0)
                  .sort((a, b) => b.season_number - a.season_number)
                  .map((season: Season) => {
                    // Calculate season availability for badges
                    const seasonStatus = seasonAvailability?.seasons?.find(
                      (s: { season_number: number }) =>
                        s.season_number === season.season_number
                    );
                    const availableEpisodes =
                      seasonStatus?.available_episodes || 0;
                    const totalEpisodes = season.episode_count || 0;

                    const getStatusColor = () => {
                      if (totalEpisodes === 0) return "text-gray-500";
                      if (availableEpisodes === totalEpisodes)
                        return "text-green-500";
                      if (availableEpisodes > 0) return "text-yellow-500";
                      return "text-gray-500";
                    };

                    return (
                      <AccordionItem
                        key={season.id}
                        value={`season-${season.season_number}`}
                      >
                        <AccordionTrigger
                          className="hover:no-underline"
                          onClick={() =>
                            fetchSeasonDetails(season.season_number)
                          }
                        >
                          <div className="flex items-center justify-between w-full mr-4">
                            <div className="flex items-center gap-3">
                              <div className="text-left">
                                <div className="font-medium">{season.name}</div>
                                <div className="text-sm text-muted-foreground">
                                  {season.episode_count} episode
                                  {season.episode_count !== 1 ? "s" : ""}
                                  {season.air_date && (
                                    <span>
                                      {" "}
                                      •{" "}
                                      {new Date(season.air_date).getFullYear()}
                                    </span>
                                  )}
                                </div>
                              </div>
                            </div>
                            {totalEpisodes > 0 && (
                              <div
                                className={`text-sm font-medium ${getStatusColor()}`}
                              >
                                {availableEpisodes}/{totalEpisodes}
                              </div>
                            )}
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="space-y-3">
                            {season.overview && (
                              <p className="text-sm text-muted-foreground leading-relaxed">
                                {season.overview}
                              </p>
                            )}

                            {/* Episodes list */}
                            {loadingSeasons[season.season_number] ? (
                              <div className="flex items-center justify-center py-4">
                                <Loader2 className="w-4 h-4 animate-spin mr-2" />
                                <span className="text-sm text-muted-foreground">
                                  Loading episodes...
                                </span>
                              </div>
                            ) : seasonDetails[season.season_number]
                                ?.episodes ? (
                              <div className="space-y-2">
                                {seasonDetails[
                                  season.season_number
                                ].episodes.map((episode: Episode) => {
                                  const airDate = episode.air_date
                                    ? new Date(episode.air_date)
                                    : null;
                                  const today = new Date();
                                  const daysDiff = airDate
                                    ? Math.floor(
                                        (airDate.getTime() - today.getTime()) /
                                          (1000 * 60 * 60 * 24)
                                      )
                                    : null;

                                  const getRelativeDateBadge = () => {
                                    if (!daysDiff) return null;

                                    if (daysDiff > 0 && daysDiff <= 7) {
                                      return (
                                        <Badge variant="outline">
                                          Airing in {daysDiff} day
                                          {daysDiff !== 1 ? "s" : ""}
                                        </Badge>
                                      );
                                    } else if (daysDiff < 0 && daysDiff >= -7) {
                                      return (
                                        <Badge variant="outline">
                                          Aired {Math.abs(daysDiff)} day
                                          {Math.abs(daysDiff) !== 1
                                            ? "s"
                                            : ""}{" "}
                                          ago
                                        </Badge>
                                      );
                                    }
                                    return null;
                                  };

                                  return (
                                    <div
                                      key={episode.id}
                                      className="flex items-start gap-3 p-3 rounded-lg bg-muted/50"
                                    >
                                      <div className="flex-shrink-0 w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                                        <span className="text-xs font-medium">
                                          {episode.episode_number}
                                        </span>
                                      </div>
                                      {episode.still_path && (
                                        <div className="flex-shrink-0">
                                          <img
                                            src={`https://image.tmdb.org/t/p/w300${episode.still_path}`}
                                            alt={episode.name}
                                            className="w-28 h-16 object-cover rounded border"
                                          />
                                        </div>
                                      )}
                                      <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2 flex-wrap">
                                          <h4 className="font-medium text-sm">
                                            {episode.name}
                                          </h4>
                                          {episode.runtime &&
                                            episode.runtime > 0 && (
                                              <span className="text-xs text-muted-foreground">
                                                {formatRuntime(episode.runtime)}
                                              </span>
                                            )}
                                          {episode.air_date && (
                                            <Badge variant="outline">
                                              {airDate?.toLocaleDateString(
                                                "en-US",
                                                {
                                                  year: "numeric",
                                                  month: "long",
                                                  day: "numeric",
                                                }
                                              )}
                                            </Badge>
                                          )}
                                          {getRelativeDateBadge()}
                                        </div>

                                        {episode.overview && (
                                          <p
                                            className="text-xs text-muted-foreground mt-1"
                                            style={{
                                              display: "-webkit-box",
                                              WebkitLineClamp: 2,
                                              WebkitBoxOrient: "vertical",
                                              overflow: "hidden",
                                            }}
                                          >
                                            {episode.overview}
                                          </p>
                                        )}
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            ) : (
                              <div className="text-sm text-muted-foreground text-center py-4">
                                Click to load episode details
                              </div>
                            )}
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                    );
                  })}
              </Accordion>
            </CardContent>
          </Card>
        )}

        {/* Cast Carousel */}
        {cast.length > 0 && (
          <MediaCarousel
            title="Cast"
            icon={<Users className="w-5 h-5" />}
            data={cast}
            isLoading={false}
            error={null}
            renderItem={(member: CastMember) => (
              <div
                className="w-full cursor-pointer hover:opacity-80 transition-opacity duration-200"
                onClick={() => navigate(`/person/${member.id}`)}
              >
                <div className="aspect-[2/3] bg-muted rounded-lg overflow-hidden mb-2">
                  {member.profile_path ? (
                    <img
                      src={`https://image.tmdb.org/t/p/w185${member.profile_path}`}
                      alt={member.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center bg-muted">
                      <Users className="w-8 h-8 text-muted-foreground" />
                    </div>
                  )}
                </div>
                <div className="text-center">
                  <p className="font-medium text-sm leading-tight">
                    {member.name}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1 leading-tight">
                    {member.character}
                  </p>
                </div>
              </div>
            )}
            keyExtractor={(member: CastMember) => member.id}
            itemWidth="w-32"
            scrollAmount={300}
            maxItems={20}
            showViewAll={false}
          />
        )}

        {/* Recommendations Carousel */}
        {recommendations?.results?.length > 0 && (
          <MediaCarousel
            title="Recommended"
            icon={<ThumbsUp className="w-5 h-5" />}
            data={recommendations.results}
            isLoading={recommendationsLoading}
            error={null}
            renderItem={renderMediaItem}
            keyExtractor={(item) => item.id}
            itemWidth="w-44"
            scrollAmount={300}
            maxItems={20}
            showViewAll={false}
          />
        )}

        {/* Similar Content Carousel */}
        {similar?.results?.length > 0 && (
          <MediaCarousel
            title={`More Like This`}
            icon={
              isMovie ? (
                <Film className="w-5 h-5" />
              ) : (
                <Tv className="w-5 h-5" />
              )
            }
            data={similar.results}
            isLoading={similarLoading}
            error={null}
            renderItem={renderMediaItem}
            keyExtractor={(item) => item.id}
            itemWidth="w-44"
            scrollAmount={300}
            maxItems={20}
            showViewAll={false}
          />
        )}
      </div>

      {/* On-behalf request dialog */}
      <Dialog open={showOnBehalfDialog} onOpenChange={setShowOnBehalfDialog}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Request Media</DialogTitle>
            <DialogDescription>
              Choose who to request "
              {selectedMedia?.title || selectedMedia?.name}" for.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <Label htmlFor="user-select" className="text-sm font-medium">
              Request for:
            </Label>
            <Select value={selectedUser} onValueChange={setSelectedUser}>
              <SelectTrigger className="w-full mt-2">
                <SelectValue placeholder="Select a user or request for yourself" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="myself">Myself</SelectItem>
                {allUsers?.users?.map((user: UserWithPermissions) => (
                  <SelectItem key={user.id} value={user.id}>
                    <div className="flex items-center gap-2">
                      <Users className="w-4 h-4" />
                      <span>{user.username}</span>
                      {user.email && (
                        <span className="text-muted-foreground text-sm">
                          ({user.email})
                        </span>
                      )}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowOnBehalfDialog(false)}
              disabled={createRequestMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleOnBehalfSubmit}
              disabled={createRequestMutation.isPending}
            >
              {createRequestMutation.isPending && (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              )}
              Submit Request
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Season selection dialog */}
      <Dialog open={showSeasonDialog} onOpenChange={setShowSeasonDialog}>
        <DialogContent className="sm:max-w-[800px]">
          <DialogHeader>
            <DialogTitle>Select Seasons to Request</DialogTitle>
            <DialogDescription>
              Choose which seasons of "{tvData.name}" you'd like to request.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <div className="flex items-center gap-2 mb-4">
              <Button
                variant="outline"
                size="sm"
                onClick={handleSelectAllSeasons}
                disabled={
                  selectedSeasons.length ===
                  (tvData.seasons?.filter((s) => s.season_number > 0).length ||
                    0)
                }
              >
                Select All
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearSeasonSelection}
                disabled={selectedSeasons.length === 0}
              >
                Clear All
              </Button>
              <span className="text-sm text-muted-foreground ml-auto">
                {selectedSeasons.length} season
                {selectedSeasons.length !== 1 ? "s" : ""} selected
              </span>
            </div>

            <div className="border rounded-lg overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left p-3 font-medium">Season</th>
                    <th className="text-left p-3 font-medium">Episodes</th>
                    <th className="text-left p-3 font-medium">Select</th>
                  </tr>
                </thead>
                <tbody>
                  {tvData.seasons
                    ?.filter((season) => season.season_number > 0)
                    .map((season: Season) => {
                      // Calculate season status for availability checking
                      const seasonStatus = seasonAvailability?.seasons?.find(
                        (s: { season_number: number }) =>
                          s.season_number === season.season_number
                      );
                      const isComplete = seasonStatus?.is_complete || false;

                      return (
                        <tr
                          key={season.id}
                          className="border-b last:border-b-0 hover:bg-muted/20"
                        >
                          <td className="p-3">
                            <div className="font-medium">{season.name}</div>
                            {season.air_date && (
                              <div className="text-sm text-muted-foreground">
                                {new Date(season.air_date).getFullYear()}
                              </div>
                            )}
                          </td>
                          <td className="p-3">
                            <span className="text-sm">
                              {season.episode_count} episode
                              {season.episode_count !== 1 ? "s" : ""}
                            </span>
                          </td>
                          <td className="p-3">
                            <Checkbox
                              id={`season-${season.season_number}`}
                              checked={selectedSeasons.includes(
                                season.season_number
                              )}
                              onCheckedChange={() =>
                                handleSeasonToggle(season.season_number)
                              }
                              disabled={isComplete}
                            />
                          </td>
                        </tr>
                      );
                    })}
                </tbody>
              </table>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowSeasonDialog(false)}
              disabled={createRequestMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSeasonRequestSubmit}
              disabled={
                createRequestMutation.isPending || selectedSeasons.length === 0
              }
            >
              {createRequestMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Requesting...
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4 mr-2" />
                  Continue with {selectedSeasons.length} Season
                  {selectedSeasons.length !== 1 ? "s" : ""}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
