import React from 'react';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { ChevronLeft, ChevronRight, AlertTriangle } from 'lucide-react';

interface MediaCarouselProps<T = any> {
  title: string;
  icon?: React.ReactNode;
  data?: T[];
  isLoading?: boolean;
  error?: Error | null;
  onViewAll?: () => void;
  onItemClick?: (item: T) => void;
  renderItem: (item: T, index: number) => React.ReactNode;
  keyExtractor?: (item: T, index: number) => string | number;
  itemWidth?: string; // e.g., "w-44", "w-32", etc.
  scrollAmount?: number; // pixels to scroll
  maxItems?: number; // maximum items to show
  showScrollButtons?: boolean;
  showViewAll?: boolean;
  className?: string;
}

export default function MediaCarousel<T = any>({
  title,
  icon,
  data = [],
  isLoading = false,
  error = null,
  onViewAll,
  onItemClick,
  renderItem,
  keyExtractor = (item: T, index: number) => index,
  itemWidth = "w-44",
  scrollAmount = 300,
  maxItems = 20,
  showScrollButtons = true,
  showViewAll = true,
  className = "",
}: MediaCarouselProps<T>) {
  const sectionId = title.replace(/\s+/g, '-').toLowerCase();

  const scrollLeft = () => {
    const container = document.getElementById(`scroll-${sectionId}`);
    if (container) {
      container.scrollBy({ left: -scrollAmount, behavior: 'smooth' });
    }
  };

  const scrollRight = () => {
    const container = document.getElementById(`scroll-${sectionId}`);
    if (container) {
      container.scrollBy({ left: scrollAmount, behavior: 'smooth' });
    }
  };

  if (error) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-xl font-semibold">
            {icon}
            {title}
          </h2>
        </div>
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <AlertTriangle className="w-8 h-8 text-destructive mx-auto mb-2" />
            <p className="text-sm text-muted-foreground">Failed to load content</p>
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-xl font-semibold">
            {icon}
            {title}
          </h2>
        </div>
        <div className="flex gap-4 overflow-hidden">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className={`flex-shrink-0 ${itemWidth} space-y-2`}>
              <Skeleton className="aspect-[2/3] w-full rounded-lg" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-3 w-2/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return null;
  }

  return (
    <div className={`space-y-4 w-full max-w-[calc(100vw-4rem)] sm:max-w-[calc(100vw-8rem)] md:max-w-[calc(100vw-12rem)] lg:max-w-[calc(100vw-16rem)] xl:max-w-[calc(100vw-20rem)] ${className}`}>
      {/* Section Header */}
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xl font-semibold">
          {icon}
          {title}
        </h2>
        <div className="flex items-center gap-2 flex-shrink-0">
          {showScrollButtons && (
            <>
              <Button
                variant="outline"
                size="sm"
                className="h-8 w-8 p-0"
                onClick={scrollLeft}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="h-8 w-8 p-0"
                onClick={scrollRight}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </>
          )}
          {showViewAll && onViewAll && (
            <Button variant="ghost" size="sm" onClick={onViewAll}>
              View All
              <ChevronRight className="w-4 h-4 ml-1" />
            </Button>
          )}
        </div>
      </div>

      {/* Media Carousel */}
      <div className="relative overflow-hidden">
        <div 
          id={`scroll-${sectionId}`}
          className="flex gap-2 overflow-x-auto pb-2 scrollbar-hidden"
          style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
        >
          {data.slice(0, maxItems).map((item, index) => (
            <div 
              key={keyExtractor(item, index)} 
              className={`flex-none ${itemWidth}`}
              onClick={() => onItemClick?.(item)}
            >
              {renderItem(item, index)}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}