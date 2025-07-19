import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Command } from "lucide-react";
import { cn } from "@/lib/utils";

interface HeaderSearchBarProps {
  onCommandPaletteOpen?: () => void;
  className?: string;
  placeholder?: string;
}

export function HeaderSearchBar({ 
  onCommandPaletteOpen, 
  className,
  placeholder = "Search..."
}: HeaderSearchBarProps) {
  const [query, setQuery] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      // Navigate to requests page for searching content to request
      navigate(`/requests?tab=discover&q=${encodeURIComponent(query.trim())}`);
      setQuery("");
      inputRef.current?.blur();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Open command palette on Ctrl+K when focused
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      onCommandPaletteOpen?.();
      inputRef.current?.blur();
    }
  };

  return (
    <div className={cn("relative", className)}>
      <form onSubmit={handleSearch} className="relative">
        <div className="relative group">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4 transition-colors group-hover:text-foreground" />
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className={cn(
              "pl-10 pr-16 h-9 bg-background/60 border border-border/40 transition-all duration-200",
              "hover:bg-background/80 hover:border-border/60",
              "focus:bg-background focus:border-border",
              isFocused && "ring-1 ring-ring"
            )}
          />
          
          {/* Command palette shortcut hint */}
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onCommandPaletteOpen}
            className={cn(
              "absolute right-1 top-1/2 transform -translate-y-1/2 h-7 px-2 text-xs text-muted-foreground",
              "hover:text-foreground transition-colors"
            )}
          >
            <Command className="w-3 h-3 mr-1" />
            <span className="hidden sm:inline">K</span>
          </Button>
        </div>
      </form>
    </div>
  );
}