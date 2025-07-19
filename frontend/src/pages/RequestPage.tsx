import { useLocation } from "react-router-dom";
import { useSettings } from "@/lib/settings";
import { useRequestHandler } from "@/hooks/useRequestHandler";
import { DiscoverySections } from "@/components/DiscoverySections";
import { MoviesTab } from "@/components/tabs/MoviesTab";
import { SeriesTab } from "@/components/tabs/SeriesTab";
import { RequestsTab } from "@/components/tabs/RequestsTab";
import { OnBehalfDialog } from "@/components/OnBehalfDialog";
import Loading from "@/components/Loading";
import { RequestSystemExternal } from "@/types";

export function RequestPage() {
  const { settings, isLoading } = useSettings();
  const location = useLocation();
  
  const searchParams = new URLSearchParams(location.search);
  const activeTab = searchParams.get('tab') || 'discover';

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