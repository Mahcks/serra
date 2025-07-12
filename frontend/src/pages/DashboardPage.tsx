import CalendarWidget from "@/components/widgets/CalendarWidget";
import DownloadWidget from "@/components/widgets/DownloadWidget";
import LatestMediaWidget from "@/components/widgets/LatestMediaWidget";

export function Dashboard() {
  return (
    <div className="space-y-6">
      {/* Dashboard Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 md:gap-6">
        {/* Latest Media Widget - Large on desktop, full width on mobile */}
        <div className="lg:col-span-8">
          <LatestMediaWidget />
        </div>

        {/* Calendar Widget - Smaller on desktop, full width on mobile */}
        <div className="lg:col-span-4">
          <CalendarWidget />
        </div>
      </div>

      {/* Download Widget - Full width */}
      <DownloadWidget />
    </div>
  );
}
