import { Button } from "@/components/ui/button";
import { useSettings } from "@/lib/settings";
import Loading from "@/components/Loading";
import { RequestSystemExternal } from "@/types";

export function RequestPage() {
  const { settings, isLoading } = useSettings();
  if (isLoading) return <Loading />;

  if (settings?.request_system == RequestSystemExternal) {
    return (
      <iframe
        src={settings.request_system_url}
        className="border-0 bg-gray-900 w-full"
        title="Jellyseerr"
        sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-top-navigation-by-user-activation allow-storage-access-by-user-activation"
        referrerPolicy="strict-origin-when-cross-origin"
        allow="camera; microphone; geolocation; storage-access"
        style={{
          colorScheme: "dark",
          minHeight: "100%",
          display: "block",
          width: "100%",
          height: "100vh",}}
      />
    );
  }

  return (
    <>
      {/* Page Header */}
      <div className="px-4 py-6 sm:px-0">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Requests</h2>
        <p className="text-gray-600">Submit and manage media requests</p>
      </div>

      {/* Request Content */}
      <div className="px-4 sm:px-0">
        {/* Sample Requests List */}
        <div className="space-y-4">
          <div className="border rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="font-medium text-gray-900">The Matrix (1999)</h4>
                <p className="text-sm text-gray-500">
                  Movie • Requested by John Doe
                </p>
                <p className="text-sm text-gray-500">Requested 2 days ago</p>
              </div>
              <div className="flex items-center space-x-2">
                <span className="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 rounded-full">
                  Pending
                </span>
                <Button variant="outline" size="sm">
                  View Details
                </Button>
              </div>
            </div>
          </div>

          <div className="border rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="font-medium text-gray-900">Breaking Bad</h4>
                <p className="text-sm text-gray-500">
                  TV Series • Requested by Jane Smith
                </p>
                <p className="text-sm text-gray-500">Requested 1 week ago</p>
              </div>
              <div className="flex items-center space-x-2">
                <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full">
                  Approved
                </span>
                <Button variant="outline" size="sm">
                  View Details
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
