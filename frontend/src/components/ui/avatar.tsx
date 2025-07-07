import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";

interface AvatarProps {
  src?: string;
  alt?: string;
  fallback?: string;
  className?: string;
  size?: "sm" | "md" | "lg" | "xl";
}

const sizeClasses = {
  sm: "h-8 w-8 text-xs",
  md: "h-10 w-10 text-sm", 
  lg: "h-12 w-12 text-base",
  xl: "h-16 w-16 text-lg"
};

export function Avatar({ 
  src, 
  alt = "Avatar", 
  fallback = "U", 
  className,
  size = "md" 
}: AvatarProps) {
  const [hasError, setHasError] = useState(false);
  const [isLoading, setIsLoading] = useState(!!src);

  // Reset error state when src changes
  useEffect(() => {
    setHasError(false);
    setIsLoading(!!src);
  }, [src]);

  const handleError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    console.log("Avatar image failed to load:", src, e);
    setHasError(true);
    setIsLoading(false);
  };

  const handleLoad = () => {
    console.log("Avatar image loaded successfully:", src);
    setIsLoading(false);
  };

  const showFallback = !src || hasError || isLoading;

  return (
    <div
      className={cn(
        "relative flex shrink-0 overflow-hidden rounded-full",
        sizeClasses[size],
        className
      )}
    >
      {src && !hasError && (
        <img
          src={src}
          alt={alt}
          onError={handleError}
          onLoad={handleLoad}
          className={cn(
            "aspect-square h-full w-full object-cover",
            isLoading ? "opacity-0" : "opacity-100"
          )}
        />
      )}
      {showFallback && (
        <div className="flex h-full w-full items-center justify-center bg-muted">
          <span className="font-medium text-muted-foreground">
            {fallback.charAt(0).toUpperCase()}
          </span>
        </div>
      )}
    </div>
  );
}