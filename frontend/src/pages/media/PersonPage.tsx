import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, User, Calendar, MapPin, ExternalLink, Film, Tv } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import Loading from "@/components/shared/Loading";
import { discoverApi } from "@/lib/api";
import { ContentGrid } from "@/components/media/ContentGrid";
import type { TMDBPersonResponse, TMDBPersonMovieCast, TMDBPersonTVCast, TMDBPersonMovieCrew, TMDBPersonTVCrew, TMDBMediaItem } from "@/types";
import { useState } from "react";

export default function PersonPage() {
  const { person_id } = useParams();
  const navigate = useNavigate();
  const [showFullBiography, setShowFullBiography] = useState(false);

  const handleGoBack = () => {
    navigate(-1);
  };

  const handleRequest = (item: TMDBMediaItem) => {
    // Person pages might not have request functionality, but keep the interface consistent
    console.log("Request functionality not implemented for person pages:", item);
  };

  // Fetch person details
  const {
    data: personData,
    isLoading,
    isError,
  } = useQuery<TMDBPersonResponse>({
    queryKey: ["person", person_id],
    queryFn: () => discoverApi.getPerson(person_id!),
    enabled: !!person_id,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading />
      </div>
    );
  }

  if (isError || !personData) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-destructive mb-4">Person Not Found</h1>
          <Button onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  const profileImageUrl = personData.profile_path
    ? `https://image.tmdb.org/t/p/w500${personData.profile_path}`
    : null;

  const formatDate = (dateString: string) => {
    if (!dateString) return null;
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const getGenderLabel = (gender: number) => {
    switch (gender) {
      case 1:
        return "Female";
      case 2:
        return "Male";
      case 3:
        return "Non-binary";
      default:
        return "Not specified";
    }
  };

  const calculateAge = (birthday: string, deathday?: string) => {
    if (!birthday) return null;
    const birthDate = new Date(birthday);
    const endDate = deathday ? new Date(deathday) : new Date();
    const age = endDate.getFullYear() - birthDate.getFullYear();
    const monthDiff = endDate.getMonth() - birthDate.getMonth();
    if (monthDiff < 0 || (monthDiff === 0 && endDate.getDate() < birthDate.getDate())) {
      return age - 1;
    }
    return age;
  };

  const age = calculateAge(personData.birthday, personData.deathday);

  // Sort and prepare movie credits
  const movieCredits: (TMDBMediaItem & { character?: string })[] = personData.movie_credits?.cast
    ?.sort((a: TMDBPersonMovieCast, b: TMDBPersonMovieCast) => 
      new Date(b.release_date || "").getTime() - new Date(a.release_date || "").getTime()
    )
    .map((credit: TMDBPersonMovieCast): TMDBMediaItem & { character?: string } => ({
      id: credit.id,
      original_language: credit.original_language,
      title: credit.title,
      original_title: credit.original_title,
      overview: credit.overview,
      poster_path: credit.poster_path,
      release_date: credit.release_date,
      vote_average: credit.vote_average,
      vote_count: credit.vote_count,
      popularity: credit.popularity,
      adult: credit.adult,
      backdrop_path: credit.backdrop_path,
      genre_ids: credit.genre_ids,
      video: credit.video,
      media_type: "movie",
      character: credit.character,
    })) || [];

  // Sort and prepare TV credits
  const tvCredits: (TMDBMediaItem & { character?: string })[] = personData.tv_credits?.cast
    ?.sort((a: TMDBPersonTVCast, b: TMDBPersonTVCast) => 
      new Date(b.first_air_date || "").getTime() - new Date(a.first_air_date || "").getTime()
    )
    .map((credit: TMDBPersonTVCast): TMDBMediaItem & { character?: string } => ({
      id: credit.id,
      original_language: credit.original_language,
      title: credit.name,
      name: credit.name,
      original_name: credit.original_name,
      overview: credit.overview,
      poster_path: credit.poster_path,
      first_air_date: credit.first_air_date,
      vote_average: credit.vote_average,
      vote_count: credit.vote_count,
      popularity: credit.popularity,
      adult: credit.adult,
      backdrop_path: credit.backdrop_path,
      genre_ids: credit.genre_ids,
      origin_country: credit.origin_country,
      media_type: "tv",
      character: credit.character,
    })) || [];

  // Combine and sort crew credits from movies and TV
  const allCrewCredits = [
    ...(personData.movie_credits?.crew || []).map((credit: TMDBPersonMovieCrew): TMDBMediaItem & { job?: string; department?: string } => ({
      id: credit.id,
      original_language: credit.original_language,
      title: credit.title,
      original_title: credit.original_title,
      overview: credit.overview,
      poster_path: credit.poster_path,
      release_date: credit.release_date,
      vote_average: credit.vote_average,
      vote_count: credit.vote_count,
      popularity: credit.popularity,
      adult: credit.adult,
      backdrop_path: credit.backdrop_path,
      genre_ids: credit.genre_ids,
      video: credit.video,
      media_type: "movie",
      job: credit.job,
      department: credit.department,
    })),
    ...(personData.tv_credits?.crew || []).map((credit: TMDBPersonTVCrew): TMDBMediaItem & { job?: string; department?: string } => ({
      id: credit.id,
      original_language: credit.original_language,
      title: credit.name,
      name: credit.name,
      original_name: credit.original_name,
      overview: credit.overview,
      poster_path: credit.poster_path,
      first_air_date: credit.first_air_date,
      vote_average: credit.vote_average,
      vote_count: credit.vote_count,
      popularity: credit.popularity,
      adult: credit.adult,
      backdrop_path: credit.backdrop_path,
      genre_ids: credit.genre_ids,
      origin_country: credit.origin_country,
      media_type: "tv",
      job: credit.job,
      department: credit.department,
    })),
  ].sort((a, b) => {
    const dateA = new Date(a.release_date || a.first_air_date || "").getTime();
    const dateB = new Date(b.release_date || b.first_air_date || "").getTime();
    return dateB - dateA;
  });


  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <div className="relative">
        <div className="container mx-auto px-4 py-12">
          <div className="flex justify-start mb-6">
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

          <div className="flex flex-col md:flex-row gap-8 items-start">
            {/* Profile Image */}
            <div className="flex-shrink-0 mx-auto md:mx-0">
              <div className="w-64 h-80 bg-muted rounded-lg overflow-hidden">
                {profileImageUrl ? (
                  <img
                    src={profileImageUrl}
                    alt={personData.name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <User className="w-16 h-16 text-muted-foreground" />
                  </div>
                )}
              </div>
              <div className="mt-3 text-center">
                <h2 className="text-lg font-semibold">{personData.name}</h2>
              </div>
            </div>

            {/* Person Details */}
            <div className="flex-1 space-y-6">
              <div className="space-y-3">
                <h1 className="text-4xl font-bold">{personData.name}</h1>
                <div className="flex flex-wrap gap-2">
                  <Badge variant="secondary">
                    {personData.known_for_department}
                  </Badge>
                  <Badge variant="outline">
                    {getGenderLabel(personData.gender)}
                  </Badge>
                  {age && (
                    <Badge variant="outline">
                      {age} years old{personData.deathday && " (deceased)"}
                    </Badge>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {personData.birthday && (
                  <div className="flex items-center gap-2">
                    <Calendar className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm">
                      Born: {formatDate(personData.birthday)}
                    </span>
                  </div>
                )}
                {personData.deathday && (
                  <div className="flex items-center gap-2">
                    <Calendar className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm">
                      Died: {formatDate(personData.deathday)}
                    </span>
                  </div>
                )}
                {personData.place_of_birth && (
                  <div className="flex items-center gap-2">
                    <MapPin className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm">{personData.place_of_birth}</span>
                  </div>
                )}
                {personData.imdb_id && (
                  <div className="flex items-center gap-2">
                    <ExternalLink className="w-4 h-4 text-muted-foreground" />
                    <a
                      href={`https://www.imdb.com/name/${personData.imdb_id}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-500 hover:underline"
                    >
                      IMDb Profile
                    </a>
                  </div>
                )}
              </div>

              {/* Biography */}
              {personData.biography && (
                <div className="space-y-2">
                  <h3 className="text-lg font-semibold">Biography</h3>
                  <div className="text-sm text-muted-foreground leading-relaxed">
                    {personData.biography.length > 300 ? (
                      <>
                        <p>
                          {showFullBiography
                            ? personData.biography
                            : `${personData.biography.substring(0, 300)}...`}
                        </p>
                        <Button
                          variant="link"
                          size="sm"
                          onClick={() => setShowFullBiography(!showFullBiography)}
                          className="p-0 h-auto text-primary hover:text-primary/80 mt-2"
                        >
                          {showFullBiography ? "Show Less" : "Show More"}
                        </Button>
                      </>
                    ) : (
                      <p>{personData.biography}</p>
                    )}
                  </div>
                </div>
              )}

              {/* Also Known As */}
              {personData.also_known_as && personData.also_known_as.length > 0 && (
                <div className="space-y-2">
                  <h3 className="text-lg font-semibold">Also Known As</h3>
                  <div className="flex flex-wrap gap-2">
                    {personData.also_known_as.map((alias: string, index: number) => (
                      <Badge key={index} variant="outline" className="text-xs">
                        {alias}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Credits Section */}
      <div className="container mx-auto px-4 py-8">
        <Tabs defaultValue="movies" className="space-y-6">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="movies" className="flex items-center gap-2">
              <Film className="w-4 h-4" />
              Movies ({movieCredits.length})
            </TabsTrigger>
            <TabsTrigger value="tv" className="flex items-center gap-2">
              <Tv className="w-4 h-4" />
              TV Shows ({tvCredits.length})
            </TabsTrigger>
            <TabsTrigger value="crew" className="flex items-center gap-2">
              <User className="w-4 h-4" />
              Crew ({allCrewCredits.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="movies" className="space-y-4">
            <ContentGrid
              title="Movie Credits"
              data={movieCredits}
              isLoading={false}
              error={false}
              onRequest={handleRequest}
            />
          </TabsContent>

          <TabsContent value="tv" className="space-y-4">
            <ContentGrid
              title="TV Credits"
              data={tvCredits}
              isLoading={false}
              error={false}
              onRequest={handleRequest}
            />
          </TabsContent>

          <TabsContent value="crew" className="space-y-4">
            <ContentGrid
              title="Crew Credits"
              data={allCrewCredits}
              isLoading={false}
              error={false}
              onRequest={handleRequest}
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}