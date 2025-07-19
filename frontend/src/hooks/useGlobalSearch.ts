import { useEffect, useCallback, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

interface UseGlobalSearchOptions {
  onCommandPaletteOpen?: () => void;
  enableInstantSearch?: boolean;
  excludePaths?: string[];
}

export function useGlobalSearch({
  onCommandPaletteOpen,
  enableInstantSearch = true,
  excludePaths = ['/search', '/login', '/setup']
}: UseGlobalSearchOptions = {}) {
  const [searchBuffer, setSearchBuffer] = useState('');
  const [bufferTimeout, setBufferTimeout] = useState<NodeJS.Timeout | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  const isExcludedPath = useCallback(() => {
    return excludePaths.some(path => location.pathname.startsWith(path));
  }, [location.pathname, excludePaths]);

  const isInputFocused = useCallback(() => {
    const activeElement = document.activeElement;
    return activeElement && (
      activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      (activeElement as HTMLElement).contentEditable === 'true' ||
      activeElement.getAttribute('role') === 'textbox'
    );
  }, []);

  const clearSearchBuffer = useCallback(() => {
    setSearchBuffer('');
    if (bufferTimeout) {
      clearTimeout(bufferTimeout);
      setBufferTimeout(null);
    }
  }, [bufferTimeout]);

  const handleInstantSearch = useCallback((query: string) => {
    if (query.trim()) {
      // Navigate to requests page with search query - this is where users search for content to request
      navigate(`/requests?tab=discover&q=${encodeURIComponent(query.trim())}`);
      clearSearchBuffer();
    }
  }, [navigate, clearSearchBuffer]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Handle Ctrl+K for command palette
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        onCommandPaletteOpen?.();
        clearSearchBuffer();
        return;
      }

      // Skip if instant search is disabled, on excluded paths, or input is focused
      if (!enableInstantSearch || isExcludedPath() || isInputFocused()) {
        return;
      }

      // Handle Escape to clear search buffer
      if (e.key === 'Escape') {
        clearSearchBuffer();
        return;
      }

      // Handle printable characters for instant search
      if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
        e.preventDefault();

        const newBuffer = searchBuffer + e.key;
        setSearchBuffer(newBuffer);

        // Clear existing timeout
        if (bufferTimeout) {
          clearTimeout(bufferTimeout);
        }

        // Set new timeout to trigger search after user stops typing
        const timeout = setTimeout(() => {
          handleInstantSearch(newBuffer);
        }, 1000); // 1 second delay

        setBufferTimeout(timeout);
      }

      // Handle backspace
      if (e.key === 'Backspace' && searchBuffer.length > 0) {
        e.preventDefault();
        const newBuffer = searchBuffer.slice(0, -1);
        setSearchBuffer(newBuffer);

        if (bufferTimeout) {
          clearTimeout(bufferTimeout);
          setBufferTimeout(null);
        }
      }

      // Handle Enter to immediately trigger search
      if (e.key === 'Enter' && searchBuffer.trim()) {
        e.preventDefault();
        handleInstantSearch(searchBuffer);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      if (bufferTimeout) {
        clearTimeout(bufferTimeout);
      }
    };
  }, [
    searchBuffer,
    bufferTimeout,
    onCommandPaletteOpen,
    enableInstantSearch,
    isExcludedPath,
    isInputFocused,
    clearSearchBuffer,
    handleInstantSearch
  ]);

  // Clear buffer when navigating away
  useEffect(() => {
    clearSearchBuffer();
  }, [location.pathname, clearSearchBuffer]);

  return {
    searchBuffer,
    clearSearchBuffer,
    isSearching: searchBuffer.length > 0
  };
}