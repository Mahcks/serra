import { useLocation, useNavigate } from "react-router-dom";
import { useSettings } from "@/lib/settings";
import { useRequestHandler } from "@/hooks/useRequestHandler";
import { DiscoverySections } from "@/components/media/DiscoverySections";
import { MoviesTab } from "@/components/media/MoviesTab";
import { SeriesTab } from "@/components/media/SeriesTab";
import { RequestsTab } from "@/components/user/RequestsTab";
import { OnBehalfDialog } from "@/components/shared/OnBehalfDialog";
import Loading from "@/components/shared/Loading";
import { RequestSystemExternal } from "@/types";
import { useEffect } from "react";

export function RequestPage() {
  const { settings, isLoading } = useSettings();
  const location = useLocation();
  const navigate = useNavigate();
  
  const searchParams = new URLSearchParams(location.search);
  const activeTab = searchParams.get('tab') || 'discover';
  const searchQuery = searchParams.get('q');

  // Redirect to search page if there's a search query
  useEffect(() => {
    if (searchQuery) {
      navigate(`/search?q=${encodeURIComponent(searchQuery)}`, { replace: true });
    }
  }, [searchQuery, navigate]);

  const {
    showOnBehalfDialog,
    setShowOnBehalfDialog,
    selectedMedia,
    selectedUser,
    setSelectedUser,
    allUsers,
    createRequestMutation,
    handleRequest,
    handleOnBehalfSubmit,
  } = useRequestHandler();

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
      <div className="container mx-auto px-6 py-8">
        {/* Content based on active tab */}
        {activeTab === 'discover' && (
          <DiscoverySections onRequest={handleRequest} />
        )}

        {activeTab === 'movies' && (
          <MoviesTab 
            onRequest={handleRequest} 
            isRequestLoading={createRequestMutation.isPending}
          />
        )}

        {activeTab === 'series' && (
          <SeriesTab 
            onRequest={handleRequest} 
            isRequestLoading={createRequestMutation.isPending}
          />
        )}

        {activeTab === 'requests' && (
          <RequestsTab />
        )}
      </div>

      {/* On-behalf request dialog */}
      <OnBehalfDialog
        open={showOnBehalfDialog}
        onOpenChange={setShowOnBehalfDialog}
        selectedMedia={selectedMedia}
        selectedUser={selectedUser}
        onSelectedUserChange={setSelectedUser}
        allUsers={allUsers}
        onSubmit={handleOnBehalfSubmit}
        isLoading={createRequestMutation.isPending}
      />
    </div>
  );
}