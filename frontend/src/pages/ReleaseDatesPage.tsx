import { useQuery } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Calendar, Globe, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { discoverApi } from "@/lib/api";
import Loading from "@/components/Loading";

interface ReleaseDateInfo {
  certification: string;
  release_date: string;
  type: number;
  note?: string;
}

interface CountryReleaseDate {
  iso_3166_1: string;
  release_dates: ReleaseDateInfo[];
}

interface ReleaseDatesResponse {
  id: number;
  results: CountryReleaseDate[];
}

// Map release type numbers to readable names
const getReleaseTypeName = (type: number): string => {
  const typeMap: Record<number, string> = {
    1: "Premiere",
    2: "Theatrical (Limited)",
    3: "Theatrical",
    4: "Digital",
    5: "Physical",
    6: "TV",
  };
  return typeMap[type] || "Unknown";
};

// Map country codes to readable names
const getCountryName = (code: string): string => {
  const countryMap: Record<string, string> = {
    US: "United States",
    GB: "United Kingdom", 
    CA: "Canada",
    AU: "Australia",
    DE: "Germany",
    FR: "France",
    IT: "Italy",
    ES: "Spain",
    JP: "Japan",
    KR: "South Korea",
    CN: "China",
    IN: "India",
    BR: "Brazil",
    MX: "Mexico",
    NL: "Netherlands",
    SE: "Sweden",
    NO: "Norway",
    DK: "Denmark",
    FI: "Finland",
    BE: "Belgium",
    CH: "Switzerland",
    AT: "Austria",
    PL: "Poland",
    CZ: "Czech Republic",
    HU: "Hungary",
    RO: "Romania",
    BG: "Bulgaria",
    GR: "Greece",
    TR: "Turkey",
    RU: "Russia",
    UA: "Ukraine",
    EE: "Estonia",
    LV: "Latvia",
    LT: "Lithuania",
    SK: "Slovakia",
    SI: "Slovenia",
    HR: "Croatia",
    RS: "Serbia",
    BA: "Bosnia and Herzegovina",
    ME: "Montenegro",
    MK: "North Macedonia",
    AL: "Albania",
    XK: "Kosovo",
    MD: "Moldova",
    BY: "Belarus",
    IS: "Iceland",
    IE: "Ireland",
    PT: "Portugal",
    LU: "Luxembourg",
    MT: "Malta",
    CY: "Cyprus",
    LI: "Liechtenstein",
    AD: "Andorra",
    MC: "Monaco",
    SM: "San Marino",
    VA: "Vatican City",
  };
  return countryMap[code] || code;
};

// Get badge color for release type
const getReleaseTypeColor = (type: number): string => {
  const colorMap: Record<number, string> = {
    1: "bg-purple-500",    // Premiere
    2: "bg-orange-500",    // Theatrical (Limited)
    3: "bg-green-500",     // Theatrical
    4: "bg-blue-500",      // Digital
    5: "bg-gray-500",      // Physical
    6: "bg-yellow-500",    // TV
  };
  return colorMap[type] || "bg-gray-400";
};

export default function ReleaseDatesPage() {
  const { tmdb_id, media_type } = useParams();
  const navigate = useNavigate();

  console.log("🔥 ReleaseDatesPage component loaded!");
  console.log("ReleaseDatesPage params:", { tmdb_id, media_type });

  // Quick test - let's just render the params first
  if (!tmdb_id || !media_type) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Debug Info</h1>
          <p>tmdb_id: {tmdb_id || "undefined"}</p>
          <p>media_type: {media_type || "undefined"}</p>
          <button 
            onClick={() => navigate(-1)}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  const {
    data: releaseDates,
    isLoading,
    isError,
    error,
  } = useQuery<ReleaseDatesResponse>({
    queryKey: ["movieReleaseDates", tmdb_id],
    queryFn: async () => {
      console.log("Making API call for release dates", tmdb_id);
      if (!tmdb_id) throw new Error("No TMDB ID provided");
      if (media_type !== "movie") throw new Error("Release dates only available for movies");
      
      const result = await discoverApi.getMovieReleaseDates(tmdb_id);
      console.log("Release dates result:", result);
      return result;
    },
    enabled: !!tmdb_id && media_type === "movie",
  });

  console.log("Query state:", { isLoading, isError, error, enabled: !!tmdb_id && media_type === "movie" });

  const handleGoBack = () => {
    navigate(`/requests/${media_type}/${tmdb_id}/details`);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading />
      </div>
    );
  }

  if (isError || !releaseDates || media_type !== "movie") {
    console.log("Error state:", { isError, error, releaseDates, media_type });
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-destructive mb-4">
            {media_type !== "movie" ? "Not Available" : "Release Dates Not Found"}
          </h1>
          <p className="text-muted-foreground mb-4">
            {media_type !== "movie" 
              ? "Release dates are only available for movies"
              : "Unable to load release date information"
            }
          </p>
          {error && (
            <p className="text-sm text-destructive mb-4">
              Error: {error.message}
            </p>
          )}
          <Button onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  // Sort countries by release date (earliest first)
  const sortedResults = [...releaseDates.results].sort((a, b) => {
    const aEarliest = a.release_dates.reduce((earliest, current) => {
      return new Date(current.release_date) < new Date(earliest.release_date) ? current : earliest;
    });
    const bEarliest = b.release_dates.reduce((earliest, current) => {
      return new Date(current.release_date) < new Date(earliest.release_date) ? current : earliest;
    });
    return new Date(aEarliest.release_date).getTime() - new Date(bEarliest.release_date).getTime();
  });

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="sticky top-0 z-50 bg-background/95 backdrop-blur-sm border-b">
        <div className="container mx-auto px-4 py-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleGoBack}
            className="gap-2"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Details
          </Button>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2 flex items-center gap-2">
            <Calendar className="w-8 h-8" />
            Release Dates
          </h1>
          <p className="text-muted-foreground">
            International release information across different countries and formats
          </p>
        </div>

        {/* Release Type Legend */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Info className="w-5 h-5" />
              Release Types
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {[
                { type: 1, name: "Premiere" },
                { type: 2, name: "Theatrical (Limited)" },
                { type: 3, name: "Theatrical" },
                { type: 4, name: "Digital" },
                { type: 5, name: "Physical" },
                { type: 6, name: "TV" },
              ].map(({ type, name }) => (
                <Badge
                  key={type}
                  className={`${getReleaseTypeColor(type)} text-white`}
                >
                  {name}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Release Dates by Country */}
        <div className="grid gap-6">
          {sortedResults.map((country) => (
            <Card key={country.iso_3166_1}>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Globe className="w-5 h-5" />
                  {getCountryName(country.iso_3166_1)}
                  <Badge variant="outline">{country.iso_3166_1}</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {country.release_dates
                    .sort((a, b) => new Date(a.release_date).getTime() - new Date(b.release_date).getTime())
                    .map((release, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-3 bg-muted/30 rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <Badge
                            className={`${getReleaseTypeColor(release.type)} text-white`}
                          >
                            {getReleaseTypeName(release.type)}
                          </Badge>
                          <div>
                            <p className="font-medium">
                              {new Date(release.release_date).toLocaleDateString("en-US", {
                                year: "numeric",
                                month: "long",
                                day: "numeric",
                              })}
                            </p>
                            {release.note && (
                              <p className="text-sm text-muted-foreground">{release.note}</p>
                            )}
                          </div>
                        </div>
                        {release.certification && (
                          <Badge variant="secondary">{release.certification}</Badge>
                        )}
                      </div>
                    ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {sortedResults.length === 0 && (
          <Card>
            <CardContent className="text-center py-12">
              <Calendar className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">No Release Dates Available</h3>
              <p className="text-muted-foreground">
                Release date information is not available for this movie.
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}