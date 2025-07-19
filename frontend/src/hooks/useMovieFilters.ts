import { useState, useCallback, useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

export interface FilterParams {
  genres: number[];
  yearFrom: number;
  yearTo: number;
  runtime: [number, number];
  includeAdult: boolean;
  voteCountMin: number;
  studios: Array<{id: number, name: string}>;
  streamingServices: number[];
  keywords: string;
  contentRating: string;
}

export function useMovieFilters() {
  const location = useLocation();
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(location.search);

  // Initialize filters from URL params
  const [movieFilters, setMovieFilters] = useState<FilterParams>({
    genres: searchParams.get('genres')?.split(',').map(Number).filter(Boolean) || [],
    yearFrom: Number(searchParams.get('yearFrom')) || 1900,
    yearTo: Number(searchParams.get('yearTo')) || new Date().getFullYear(),
    runtime: searchParams.get('runtime')?.split(',').map(Number) as [number, number] || [0, 300],
    includeAdult: searchParams.get('includeAdult') === 'true',
    voteCountMin: Number(searchParams.get('voteCountMin')) || 0,
    studios: [],
    streamingServices: searchParams.get('streamingServices')?.split(',').map(Number).filter(Boolean) || [],
    keywords: searchParams.get('keywords') || '',
    contentRating: searchParams.get('contentRating') || 'all',
  });

  // Update URL when filters change
  const updateURL = useCallback((newFilters?: FilterParams, newSort?: string) => {
    const params = new URLSearchParams(location.search);
    
    if (newSort) {
      params.set('sort', newSort);
    }
    
    if (newFilters) {
      if (newFilters.genres.length > 0) {
        params.set('genres', newFilters.genres.join(','));
      } else {
        params.delete('genres');
      }
      
      if (newFilters.yearFrom !== 1900) params.set('yearFrom', newFilters.yearFrom.toString());
      else params.delete('yearFrom');
      
      if (newFilters.yearTo !== new Date().getFullYear()) params.set('yearTo', newFilters.yearTo.toString());
      else params.delete('yearTo');
      
      if (newFilters.runtime[0] !== 0 || newFilters.runtime[1] !== 300) {
        params.set('runtime', newFilters.runtime.join(','));
      } else {
        params.delete('runtime');
      }
      
      if (newFilters.includeAdult) params.set('includeAdult', 'true');
      else params.delete('includeAdult');
      
      if (newFilters.voteCountMin !== 0) params.set('voteCountMin', newFilters.voteCountMin.toString());
      else params.delete('voteCountMin');
      
      if (newFilters.studios.length > 0) {
        const studioIds = newFilters.studios.map(s => s.id.toString()).join(',');
        params.set('studios', studioIds);
      } else {
        params.delete('studios');
      }
      
      if (newFilters.streamingServices.length > 0) params.set('streamingServices', newFilters.streamingServices.join(','));
      else params.delete('streamingServices');
      
      if (newFilters.keywords.trim()) params.set('keywords', newFilters.keywords.trim());
      else params.delete('keywords');
      
      if (newFilters.contentRating && newFilters.contentRating !== 'all') params.set('contentRating', newFilters.contentRating);
      else params.delete('contentRating');
    }
    
    navigate(`${location.pathname}?${params.toString()}`, { replace: true });
  }, [location, navigate]);

  const handleFiltersChange = useCallback((newFilters: FilterParams) => {
    setMovieFilters(newFilters);
    updateURL(newFilters);
  }, [updateURL]);

  const clearFilters = useCallback(() => {
    const defaultFilters: FilterParams = {
      genres: [],
      yearFrom: 1900,
      yearTo: new Date().getFullYear(),
      runtime: [0, 300],
      includeAdult: false,
      voteCountMin: 0,
      studios: [],
      streamingServices: [],
      keywords: '',
      contentRating: 'all',
    };
    handleFiltersChange(defaultFilters);
  }, [handleFiltersChange]);

  const hasActiveFilters = useMemo(() => {
    return movieFilters.genres.length > 0 ||
           movieFilters.yearFrom !== 1900 ||
           movieFilters.yearTo !== new Date().getFullYear() ||
           movieFilters.runtime[0] !== 0 ||
           movieFilters.runtime[1] !== 300 ||
           movieFilters.includeAdult ||
           movieFilters.voteCountMin !== 0 ||
           movieFilters.studios.length > 0 ||
           movieFilters.streamingServices.length > 0 ||
           movieFilters.keywords.trim().length > 0 ||
           (movieFilters.contentRating && movieFilters.contentRating !== 'all');
  }, [movieFilters]);

  const toggleGenre = useCallback((genreId: number) => {
    const newGenres = movieFilters.genres.includes(genreId)
      ? movieFilters.genres.filter(id => id !== genreId)
      : [...movieFilters.genres, genreId];
    
    const newFilters = { ...movieFilters, genres: newGenres };
    handleFiltersChange(newFilters);
  }, [movieFilters, handleFiltersChange]);

  const toggleStreamingService = useCallback((serviceId: number) => {
    const newServices = movieFilters.streamingServices.includes(serviceId)
      ? movieFilters.streamingServices.filter(id => id !== serviceId)
      : [...movieFilters.streamingServices, serviceId];
    
    const newFilters = { ...movieFilters, streamingServices: newServices };
    handleFiltersChange(newFilters);
  }, [movieFilters, handleFiltersChange]);

  return {
    movieFilters,
    setMovieFilters,
    handleFiltersChange,
    clearFilters,
    hasActiveFilters,
    toggleGenre,
    toggleStreamingService,
    updateURL,
  };
}