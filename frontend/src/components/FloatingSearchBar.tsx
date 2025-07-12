import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, X } from "lucide-react";

interface FloatingSearchBarProps {
  onClose?: () => void;
}

export function FloatingSearchBar({ onClose }: FloatingSearchBarProps) {
  const [query, setQuery] = useState("");
  const [isVisible, setIsVisible] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const location = useLocation();

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        handleClose();
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setIsVisible(true);
        setTimeout(() => inputRef.current?.focus(), 100);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  // Close search when route changes
  useEffect(() => {
    handleClose();
  }, [location.pathname]);

  const handleClose = () => {
    setIsVisible(false);
    setQuery("");
    onClose?.();
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      navigate(`/search?q=${encodeURIComponent(query.trim())}`);
      handleClose();
    }
  };

  const handleInputFocus = () => {
    setIsVisible(true);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newQuery = e.target.value;
    setQuery(newQuery);
    
    // Navigate to search page immediately when user starts typing
    if (newQuery.trim() && location.pathname !== '/search') {
      navigate(`/search?q=${encodeURIComponent(newQuery.trim())}`);
    } else if (newQuery.trim() && location.pathname === '/search') {
      // Update URL params if already on search page
      navigate(`/search?q=${encodeURIComponent(newQuery.trim())}`, { replace: true });
    }
  };

  // Don't show on search page itself
  if (location.pathname === '/search') {
    return null;
  }

  return (
    <>
      <div className="fixed top-4 left-1/2 transform -translate-x-1/2 w-full max-w-2xl mx-auto px-4 z-50">
        <form onSubmit={handleSearch} className="relative">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
            <Input
              ref={inputRef}
              value={query}
              onChange={handleInputChange}
              onFocus={handleInputFocus}
              placeholder="Search for movies and TV shows... (Ctrl+K)"
              className="pl-10 pr-10 h-12 text-base bg-background/95 backdrop-blur-sm border-2 shadow-lg"
            />
            {(query || isVisible) && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleClose}
                className="absolute right-1 top-1/2 transform -translate-y-1/2 h-8 w-8 p-0"
              >
                <X className="w-4 h-4" />
              </Button>
            )}
          </div>
        </form>

        {/* Quick hint */}
        {isVisible && !query && (
          <div className="absolute top-full mt-2 w-full">
            <div className="bg-background/95 backdrop-blur-sm border rounded-lg p-4 shadow-lg">
              <p className="text-sm text-muted-foreground text-center">
                Start typing to search for movies and TV shows, or press Enter to go to the search page
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Backdrop */}
      {isVisible && (
        <div
          className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40"
          onClick={handleClose}
        />
      )}
    </>
  );
}