import { XCircle } from "lucide-react";
import { RequestableMediaCard } from "@/components/RequestableMediaCard";
import { type TMDBMediaItem } from "@/types";

interface ContentGridProps {
  title: string;
  data: TMDBMediaItem[];
  isLoading: boolean;
  error: unknown;
  onRequest: (item: TMDBMediaItem) => void;
  isRequestLoading?: boolean;
}

export function ContentGrid({ 
  title, 
  data, 
  isLoading, 
  error, 
  onRequest, 
  isRequestLoading 
}: ContentGridProps) {
  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <XCircle className="w-12 h-12 text-destructive mx-auto mb-4" />
          <p className="text-destructive">Failed to load {title.toLowerCase()}</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-6">
        {[...Array(24)].map((_, i) => (
          <div key={i} className="aspect-[2/3] bg-muted rounded-lg animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-6">
      {data.map((item, index) => (
        <RequestableMediaCard
          key={`${title}-${item.id}-${index}`}
          item={item}
          onRequest={onRequest}
          size="md"
          isRequestLoading={isRequestLoading}
        />
      ))}
    </div>
  );
}