import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Folder, Film } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import Loading from "@/components/shared/Loading";
import { discoverApi } from "@/lib/api";
import { ContentGrid } from "@/components/media/ContentGrid";
import { OnBehalfRequestDialog, SeasonSelectionDialog } from "@/components/media/RequestDialogs";
import { useAdvancedRequestHandler } from "@/hooks/useAdvancedRequestHandler";
import type { TMDBCollectionResponse } from "@/types";

export default function CollectionPage() {
  const { collection_id } = useParams();
  const navigate = useNavigate();
  
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
    queryKeysToInvalidate: [["collection", collection_id]],
  });

  const handleGoBack = () => {
    navigate(-1);
  };

  // Fetch collection details
  const {
    data: collectionData,
    isLoading,
    isError,
  } = useQuery<TMDBCollectionResponse>({
    queryKey: ["collection", collection_id],
    queryFn: () => discoverApi.getCollection(collection_id!),
    enabled: !!collection_id,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading />
      </div>
    );
  }

  if (isError || !collectionData) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-destructive mb-4">
            Collection Not Found
          </h1>
          <Button onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  const backdropUrl = collectionData.backdrop_path
    ? `https://image.tmdb.org/t/p/w1280${collectionData.backdrop_path}`
    : null;

  return (
    <div className="min-h-screen bg-background">
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

        <div className="relative container mx-auto px-4 py-12">
          <div className="text-center space-y-6">
            <div className="flex justify-start mb-6">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleGoBack}
                className="gap-2 bg-background/20 hover:bg-background/30 backdrop-blur-sm"
              >
                <ArrowLeft className="w-4 h-4" />
                Back
              </Button>
            </div>
            
            <div className="inline-flex items-center justify-center w-20 h-20 bg-amber-500/10 rounded-full">
              <Folder className="w-10 h-10 text-amber-500" />
            </div>

            <div className="space-y-3">
              <h1 className="text-4xl font-bold">{collectionData.name}</h1>
              {collectionData.overview && (
                <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
                  {collectionData.overview}
                </p>
              )}
              <div className="flex items-center justify-center gap-4 text-sm text-muted-foreground">
                <span>{collectionData.parts?.length || 0} Movies</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Movies Grid */}
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div className="flex items-center gap-2">
            <Film className="w-5 h-5 text-blue-500" />
            <h2 className="text-2xl font-semibold">Movies in Collection</h2>
            {collectionData?.parts && (
              <Badge variant="outline" className="ml-2">
                {collectionData.parts.length}{" "}
                {collectionData.parts.length === 1 ? "Movie" : "Movies"}
              </Badge>
            )}
          </div>

          <ContentGrid
            title="Movies in Collection"
            data={
              collectionData?.parts?.sort(
                (a: TMDBMediaItem, b: TMDBMediaItem) =>
                  new Date(a.release_date || "").getTime() -
                  new Date(b.release_date || "").getTime()
              ) || []
            }
            isLoading={isLoading}
            error={isError}
            onRequest={handleRequest}
            isRequestLoading={isRequestLoading}
          />
        </div>
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
