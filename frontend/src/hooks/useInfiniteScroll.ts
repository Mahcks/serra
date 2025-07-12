import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { type TMDBMediaResponse } from '@/types';

interface UseInfiniteScrollOptions {
  queryKey: string[];
  queryFn: (pageParam: number) => Promise<TMDBMediaResponse>;
  threshold?: number; // Distance from bottom to trigger load (in pixels)
  enabled?: boolean;
}

export function useInfiniteScroll({
  queryKey,
  queryFn,
  threshold = 1000,
  enabled = true
}: UseInfiniteScrollOptions) {
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const loadingRef = useRef<HTMLDivElement>(null);

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
    error
  } = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam }) => queryFn(pageParam as number),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      // The TMDBPageResults fields are embedded at the root level, not nested
      const hasNext = lastPage?.page && lastPage?.total_pages && 
          lastPage.page < lastPage.total_pages;
      const nextPage = hasNext ? lastPage.page + 1 : undefined;
      
      
      return nextPage;
    },
    enabled,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Flatten all pages into a single array - memoize to prevent re-computation
  const allItems = useMemo(() => 
    data?.pages.flatMap(page => page.results) ?? [], 
    [data?.pages]
  );
  const totalResults = data?.pages[0]?.total_results ?? 0;
  const currentPage = data?.pages[data.pages.length - 1]?.page ?? 1;
  const totalPages = data?.pages[0]?.total_pages ?? 1;
  



  // Intersection Observer for infinite scroll - check whenever element availability changes
  useEffect(() => {
    if (!hasNextPage) return;
    
    let observer: IntersectionObserver | null = null;
    
    const setupObserver = () => {
      const loadingElement = loadingRef.current;
      
      if (loadingElement && !observer) {
        observer = new IntersectionObserver(
          (entries) => {
            const [entry] = entries;
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
              setIsLoadingMore(true);
              fetchNextPage().finally(() => setIsLoadingMore(false));
            }
          },
          {
            root: null, // Use viewport as root
            rootMargin: `${threshold}px`,
            threshold: [0, 0.1, 0.5, 1.0], // Multiple thresholds for better detection
          }
        );
        
        observer.observe(loadingElement);
      } else if (!loadingElement && observer) {
        observer.disconnect();
        observer = null;
      }
    };
    
    // Check immediately
    setupObserver();
    
    // Check periodically for element availability changes
    const interval = setInterval(setupObserver, 100);

    return () => {
      clearInterval(interval);
      if (observer) {
        observer.disconnect();
      }
    };
  }, [fetchNextPage, hasNextPage, threshold, queryKey, isFetchingNextPage]);

  // Manual load more function
  const loadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      setIsLoadingMore(true);
      fetchNextPage().finally(() => setIsLoadingMore(false));
    }
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  return {
    // Data
    items: allItems,
    totalResults,
    currentPage,
    totalPages,
    
    // States
    isLoading,
    isLoadingMore: isLoadingMore || isFetchingNextPage,
    hasNextPage,
    isError,
    error,
    
    // Actions
    loadMore,
    
    // Refs
    loadingRef,
    
    // Stats for debugging/display
    loadedCount: allItems.length,
    hasMore: hasNextPage,
    pagesLoaded: data?.pages?.length || 0,
  };
}