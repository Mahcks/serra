import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Search, ArrowRight, Film, Tv, Clock, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface SearchSuggestion {
  type: 'action' | 'recent' | 'popular';
  label: string;
  description?: string;
  icon: React.ReactNode;
  action: () => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const location = useLocation();

  // Sample suggestions - can be enhanced with real data
  const suggestions: SearchSuggestion[] = [
    {
      type: 'action',
      label: 'Search Movies',
      description: 'Find movies to add to your library',
      icon: <Film className="w-4 h-4" />,
      action: () => {
        navigate('/requests?tab=movies');
        onOpenChange(false);
      }
    },
    {
      type: 'action',
      label: 'Search TV Shows',
      description: 'Find TV series to add to your library',
      icon: <Tv className="w-4 h-4" />,
      action: () => {
        navigate('/requests?tab=series');
        onOpenChange(false);
      }
    },
    {
      type: 'action',
      label: 'Browse Trending',
      description: 'See what\'s popular right now',
      icon: <TrendingUp className="w-4 h-4" />,
      action: () => {
        navigate('/requests?tab=discover');
        onOpenChange(false);
      }
    },
    {
      type: 'action',
      label: 'My Requests',
      description: 'View your request history',
      icon: <Clock className="w-4 h-4" />,
      action: () => {
        navigate('/requests?tab=requests');
        onOpenChange(false);
      }
    }
  ];

  const filteredSuggestions = query.trim() 
    ? suggestions.filter(s => 
        s.label.toLowerCase().includes(query.toLowerCase()) ||
        s.description?.toLowerCase().includes(query.toLowerCase())
      )
    : suggestions;

  // Focus input when dialog opens
  useEffect(() => {
    if (open && inputRef.current) {
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [open]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setQuery("");
      setSelectedIndex(0);
    }
  }, [open]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!open) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev < filteredSuggestions.length - 1 ? prev + 1 : 0
          );
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev > 0 ? prev - 1 : filteredSuggestions.length - 1
          );
          break;
        case 'Enter':
          e.preventDefault();
          if (query.trim()) {
            // Direct search - go to requests page for content discovery
            navigate(`/requests?tab=discover&q=${encodeURIComponent(query.trim())}`);
            onOpenChange(false);
          } else if (filteredSuggestions[selectedIndex]) {
            // Execute selected suggestion
            filteredSuggestions[selectedIndex].action();
          }
          break;
        case 'Escape':
          onOpenChange(false);
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, query, selectedIndex, filteredSuggestions, navigate, onOpenChange]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      // Navigate to requests page for content discovery
      navigate(`/requests?tab=discover&q=${encodeURIComponent(query.trim())}`);
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] p-0">
        <DialogHeader className="px-6 pt-6 pb-2">
          <DialogTitle className="text-sm font-medium text-muted-foreground">
            Search Serra
          </DialogTitle>
        </DialogHeader>
        
        <div className="px-6">
          <form onSubmit={handleSearch} className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
            <Input
              ref={inputRef}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search for movies, TV shows, or use commands..."
              className="pl-10 h-12 text-base border-0 bg-muted/30 focus-visible:ring-0 focus-visible:ring-offset-0"
            />
          </form>
        </div>

        <div className="max-h-[400px] overflow-y-auto">
          {filteredSuggestions.length > 0 ? (
            <div className="px-6 pb-6">
              <div className="space-y-1">
                {filteredSuggestions.map((suggestion, index) => (
                  <button
                    key={`${suggestion.type}-${suggestion.label}`}
                    onClick={suggestion.action}
                    className={cn(
                      "w-full flex items-center gap-3 p-3 rounded-lg text-left transition-colors",
                      "hover:bg-muted/50",
                      selectedIndex === index && "bg-muted/70"
                    )}
                  >
                    <div className="flex-shrink-0 text-muted-foreground">
                      {suggestion.icon}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium">{suggestion.label}</div>
                      {suggestion.description && (
                        <div className="text-sm text-muted-foreground truncate">
                          {suggestion.description}
                        </div>
                      )}
                    </div>
                    <div className="flex-shrink-0">
                      <ArrowRight className="w-4 h-4 text-muted-foreground" />
                    </div>
                  </button>
                ))}
              </div>
            </div>
          ) : query.trim() ? (
            <div className="px-6 pb-6 text-center">
              <div className="text-muted-foreground">
                Press Enter to search for "{query}"
              </div>
            </div>
          ) : null}
        </div>

        <div className="px-6 py-3 border-t bg-muted/20">
          <div className="flex justify-between text-xs text-muted-foreground">
            <div className="flex gap-4">
              <span><kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium">↑↓</kbd> navigate</span>
              <span><kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium">↵</kbd> select</span>
            </div>
            <span><kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium">esc</kbd> close</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}