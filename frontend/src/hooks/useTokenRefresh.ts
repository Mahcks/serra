import { useEffect, useRef } from 'react';
import { proactiveTokenRefresh } from '@/lib/api';

/**
 * Hook to automatically refresh tokens in the background
 * @param intervalMinutes - How often to check for token refresh (default: 15 minutes)
 */
export const useTokenRefresh = (intervalMinutes: number = 15) => {
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Initial check
    proactiveTokenRefresh();

    // Set up periodic refresh checks
    intervalRef.current = setInterval(() => {
      proactiveTokenRefresh();
    }, intervalMinutes * 60 * 1000);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [intervalMinutes]);

  // Manual refresh function
  const refreshNow = async () => {
    return proactiveTokenRefresh();
  };

  return { refreshNow };
};

/**
 * Hook to refresh token when component mounts or becomes visible
 * Useful for critical components that need fresh tokens
 */
export const useTokenRefreshOnMount = () => {
  useEffect(() => {
    proactiveTokenRefresh();
  }, []);

  // Also refresh when page becomes visible (user returns to tab)
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        proactiveTokenRefresh();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);
};