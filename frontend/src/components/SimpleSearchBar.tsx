import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface SimpleSearchBarProps {
  placeholder?: string;
  className?: string;
  onSearch?: (query: string) => void;
  autoFocus?: boolean;
  size?: "sm" | "md" | "lg";
}

export function SimpleSearchBar({ 
  placeholder = "Search for movies and TV shows...",
  className,
  onSearch,
  autoFocus = false,
  size = "md"
}: SimpleSearchBarProps) {
  const [query, setQuery] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedQuery = query.trim();
    
    if (trimmedQuery) {
      if (onSearch) {
        onSearch(trimmedQuery);
      } else {
        // Default behavior: navigate to requests page for content discovery
        navigate(`/requests?tab=discover&q=${encodeURIComponent(trimmedQuery)}`);
      }
    }
  };

  const handleClear = () => {
    setQuery("");
    inputRef.current?.focus();
    if (onSearch) {
      onSearch("");
    }
  };

  const sizeClasses = {
    sm: "h-8 text-sm",
    md: "h-10 text-base",
    lg: "h-12 text-lg"
  };

  const iconSizes = {
    sm: "w-3 h-3",
    md: "w-4 h-4", 
    lg: "w-5 h-5"
  };

  return (
    <form onSubmit={handleSearch} className={cn("relative w-full", className)}>
      <div className="relative">
        <Search className={cn(
          "absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground",
          iconSizes[size]
        )} />
        <Input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={placeholder}
          autoFocus={autoFocus}
          className={cn(
            "pl-10 pr-10 border-border/60 transition-all duration-200",
            "focus:border-border focus:ring-1 focus:ring-ring",
            sizeClasses[size]
          )}
        />
        {query && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleClear}
            className={cn(
              "absolute right-1 top-1/2 transform -translate-y-1/2 p-0",
              size === "sm" ? "h-6 w-6" : size === "md" ? "h-8 w-8" : "h-10 w-10"
            )}
          >
            <X className={iconSizes[size]} />
          </Button>
        )}
      </div>
    </form>
  );
}