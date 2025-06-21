import CalendarWidget from "@/components/widgets/CalendarWidget";
import DownloadWidget from "@/components/widgets/DownloadWidget";
import LatestMediaWidget from "@/components/widgets/LatestMediaWidget";

export function Dashboard() {
  return (
    <>
      {/* Dashboard Grid */}
      <div className="px-4 sm:px-0">
        {/*
          To control widget widths individually, it's best to use a more granular grid, like a 12-column grid.
          You can then use Tailwind's `col-span-*` utilities to specify how many columns each widget should occupy
          on different screen sizes (e.g., `md:col-span-6`, `lg:col-span-4`).
        */}
        <div className="grid grid-cols-1 md:grid-cols-12 lg:grid-cols-12 gap-6">
          {/* --- Row 1 --- */}

          {/* Example: This widget takes up half the width on medium screens, and 2/3 on large screens. */}
          <div className="md:col-span-6 lg:col-span-8">
            <LatestMediaWidget />
          </div>

          {/* Example: This widget takes up the other half on medium, and 1/3 on large screens. */}
          <div className="md:col-span-6 lg:col-span-4">
            <CalendarWidget />
          </div>

          {/* --- Row 2 --- */}

          {/* Example: This widget takes up the full width on all screen sizes. */}
          <div className="md:col-span-12">
            <DownloadWidget />
          </div>
        </div>
      </div>
    </>
  );
}
