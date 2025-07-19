package cached

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mahcks/serra/internal/integrations/tmdb"
	"github.com/mahcks/serra/internal/services/cache"
	"github.com/mahcks/serra/pkg/structures"
)

// TMDBService wraps the TMDB service with caching capabilities
type TMDBService struct {
	tmdb  tmdb.Service
	cache *cache.TMDBCacheService
}

// NewTMDBService creates a new cached TMDB service
func NewTMDBService(tmdbService tmdb.Service, cacheService *cache.TMDBCacheService) tmdb.Service {
	return &TMDBService{
		tmdb:  tmdbService,
		cache: cacheService,
	}
}

// Helper function to get cached data or fetch from API
func (c *TMDBService) getCachedOrFetch(endpoint string, params map[string]interface{}, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Generate cache key
	cacheKey := c.cache.GenerateCacheKey(endpoint, params)

	// Try to get from cache first
	if cachedData, found, err := c.cache.GetCachedData(cacheKey); err == nil && found {
		// Unmarshal based on endpoint type
		var result interface{}
		switch {
		case endpoint == "trending/all/day" || endpoint == "movie/popular" || endpoint == "tv/popular" || endpoint == "search/movie" || endpoint == "search/tv" || endpoint == "discover/movie" || endpoint == "discover/tv":
			result = &structures.TMDBMediaResponse{}
		case endpoint == "search/company":
			result = &structures.TMDBCompanySearchResponse{}
		case endpoint == "watch/providers/movie" || endpoint == "watch/providers/tv":
			result = &structures.TMDBWatchProvidersListResponse{}
		case endpoint == "watch/providers/regions":
			result = &structures.TMDBWatchProviderRegionsResponse{}
		case strings.Contains(endpoint, "/release_dates"):
			result = &structures.TMDBReleaseDatesResponse{}
		case strings.Contains(endpoint, "collection/"):
			result = &structures.TMDBCollectionResponse{}
		case strings.Contains(endpoint, "person/"):
			result = &structures.TMDBPersonResponse{}
		case strings.Contains(endpoint, "/season/"):
			result = &structures.SeasonDetails{}
		default:
			// For movie/TV details and other specific endpoints
			result = &structures.TMDBWatchProvidersResponse{}
		}

		if err := json.Unmarshal(cachedData, result); err == nil {
			return result, nil
		}
		// If unmarshal fails, continue to fetch fresh data
	}

	// Cache miss or error - fetch from API
	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if jsonData, err := json.Marshal(data); err == nil {
		ttl := c.cache.GetTTLForEndpoint(endpoint)
		c.cache.SetCachedData(cacheKey, endpoint, jsonData, ttl)
	}

	// Track API usage
	c.cache.TrackAPIUsage(endpoint)

	return data, nil
}

// Implement all Service interface methods with caching

func (c *TMDBService) GetTrendingMedia(page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"page": page}

	result, err := c.getCachedOrFetch("trending/all/day", params, func() (interface{}, error) {
		return c.tmdb.GetTrendingMedia(page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	// Fallback to direct API call if type assertion fails
	return c.tmdb.GetTrendingMedia(page)
}

func (c *TMDBService) SearchTV(query, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"query": query, "page": page}

	result, err := c.getCachedOrFetch("search/tv", params, func() (interface{}, error) {
		return c.tmdb.SearchTV(query, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.SearchTV(query, page)
}

func (c *TMDBService) GetTVPopular(page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"page": page}

	result, err := c.getCachedOrFetch("tv/popular", params, func() (interface{}, error) {
		return c.tmdb.GetTVPopular(page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetTVPopular(page)
}

func (c *TMDBService) GetTVUpcoming(page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"page": page}

	result, err := c.getCachedOrFetch("tv/upcoming", params, func() (interface{}, error) {
		return c.tmdb.GetTVUpcoming(page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetTVUpcoming(page)
}

func (c *TMDBService) GetTVWatchProviders(seriesID string) (structures.TMDBWatchProvidersResponse, error) {
	params := map[string]interface{}{"series_id": seriesID}

	result, err := c.getCachedOrFetch(fmt.Sprintf("tv/%s/watch/providers", seriesID), params, func() (interface{}, error) {
		return c.tmdb.GetTVWatchProviders(seriesID)
	})
	if err != nil {
		return structures.TMDBWatchProvidersResponse{}, err
	}

	if response, ok := result.(*structures.TMDBWatchProvidersResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetTVWatchProviders(seriesID)
}

func (c *TMDBService) GetTvRecommendations(seriesID, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"series_id": seriesID, "page": page}

	result, err := c.getCachedOrFetch(fmt.Sprintf("tv/%s/recommendations", seriesID), params, func() (interface{}, error) {
		return c.tmdb.GetTvRecommendations(seriesID, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetTvRecommendations(seriesID, page)
}

func (c *TMDBService) GetTvSimilar(seriesID, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"series_id": seriesID, "page": page}

	result, err := c.getCachedOrFetch(fmt.Sprintf("tv/%s/similar", seriesID), params, func() (interface{}, error) {
		return c.tmdb.GetTvSimilar(seriesID, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetTvSimilar(seriesID, page)
}

func (c *TMDBService) SearchMovie(query, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"query": query, "page": page}

	result, err := c.getCachedOrFetch("search/movie", params, func() (interface{}, error) {
		return c.tmdb.SearchMovie(query, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.SearchMovie(query, page)
}

func (c *TMDBService) DiscoverMovie(params structures.DiscoverMovieParams) (structures.TMDBMediaResponse, error) {
	cacheParams := map[string]interface{}{
		"params": params,
	}

	result, err := c.getCachedOrFetch("discover/movie", cacheParams, func() (interface{}, error) {
		return c.tmdb.DiscoverMovie(params)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.DiscoverMovie(params)
}

func (c *TMDBService) GetMoviePopular(page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"page": page}

	result, err := c.getCachedOrFetch("movie/popular", params, func() (interface{}, error) {
		return c.tmdb.GetMoviePopular(page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMoviePopular(page)
}

func (c *TMDBService) GetMovieUpcoming(page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"page": page}

	result, err := c.getCachedOrFetch("movie/upcoming", params, func() (interface{}, error) {
		return c.tmdb.GetMovieUpcoming(page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMovieUpcoming(page)
}

func (c *TMDBService) GetMovieWatchProviders(movieID string) (structures.TMDBWatchProvidersResponse, error) {
	params := map[string]interface{}{"movie_id": movieID}

	result, err := c.getCachedOrFetch(fmt.Sprintf("movie/%s/watch/providers", movieID), params, func() (interface{}, error) {
		return c.tmdb.GetMovieWatchProviders(movieID)
	})
	if err != nil {
		return structures.TMDBWatchProvidersResponse{}, err
	}

	if response, ok := result.(*structures.TMDBWatchProvidersResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMovieWatchProviders(movieID)
}

func (c *TMDBService) GetMovieRecommendations(movieID, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"movie_id": movieID, "page": page}

	result, err := c.getCachedOrFetch(fmt.Sprintf("movie/%s/recommendations", movieID), params, func() (interface{}, error) {
		return c.tmdb.GetMovieRecommendations(movieID, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMovieRecommendations(movieID, page)
}

func (c *TMDBService) GetMovieSimilar(movieID, page string) (structures.TMDBMediaResponse, error) {
	params := map[string]interface{}{"movie_id": movieID, "page": page}

	result, err := c.getCachedOrFetch(fmt.Sprintf("movie/%s/similar", movieID), params, func() (interface{}, error) {
		return c.tmdb.GetMovieSimilar(movieID, page)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMovieSimilar(movieID, page)
}

func (c *TMDBService) GetWatchProviders(mediaType string) (structures.TMDBWatchProvidersListResponse, error) {
	params := map[string]interface{}{"media_type": mediaType}

	result, err := c.getCachedOrFetch(fmt.Sprintf("watch/providers/%s", mediaType), params, func() (interface{}, error) {
		return c.tmdb.GetWatchProviders(mediaType)
	})
	if err != nil {
		return structures.TMDBWatchProvidersListResponse{}, err
	}

	if response, ok := result.(*structures.TMDBWatchProvidersListResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetWatchProviders(mediaType)
}

func (c *TMDBService) GetWatchProviderRegions() (structures.TMDBWatchProviderRegionsResponse, error) {
	params := map[string]interface{}{}

	result, err := c.getCachedOrFetch("watch/providers/regions", params, func() (interface{}, error) {
		return c.tmdb.GetWatchProviderRegions()
	})
	if err != nil {
		return structures.TMDBWatchProviderRegionsResponse{}, err
	}

	if response, ok := result.(*structures.TMDBWatchProviderRegionsResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetWatchProviderRegions()
}

func (c *TMDBService) SearchCompanies(query, page string) (structures.TMDBCompanySearchResponse, error) {
	params := map[string]interface{}{"query": query, "page": page}

	result, err := c.getCachedOrFetch("search/company", params, func() (interface{}, error) {
		return c.tmdb.SearchCompanies(query, page)
	})
	if err != nil {
		return structures.TMDBCompanySearchResponse{}, err
	}

	if response, ok := result.(*structures.TMDBCompanySearchResponse); ok {
		return *response, nil
	}

	return c.tmdb.SearchCompanies(query, page)
}

func (c *TMDBService) GetMovieReleaseDates(movieID string) (structures.TMDBReleaseDatesResponse, error) {
	params := map[string]interface{}{"movie_id": movieID}

	result, err := c.getCachedOrFetch(fmt.Sprintf("movie/%s/release_dates", movieID), params, func() (interface{}, error) {
		return c.tmdb.GetMovieReleaseDates(movieID)
	})
	if err != nil {
		return structures.TMDBReleaseDatesResponse{}, err
	}

	if response, ok := result.(*structures.TMDBReleaseDatesResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetMovieReleaseDates(movieID)
}

func (c *TMDBService) GetCollection(collectionID string) (structures.TMDBCollectionResponse, error) {
	params := map[string]interface{}{"collection_id": collectionID}

	result, err := c.getCachedOrFetch(fmt.Sprintf("collection/%s", collectionID), params, func() (interface{}, error) {
		return c.tmdb.GetCollection(collectionID)
	})
	if err != nil {
		return structures.TMDBCollectionResponse{}, err
	}

	if response, ok := result.(*structures.TMDBCollectionResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetCollection(collectionID)
}

func (c *TMDBService) GetPerson(personID string) (structures.TMDBPersonResponse, error) {
	params := map[string]interface{}{"person_id": personID}

	result, err := c.getCachedOrFetch(fmt.Sprintf("person/%s", personID), params, func() (interface{}, error) {
		return c.tmdb.GetPerson(personID)
	})
	if err != nil {
		return structures.TMDBPersonResponse{}, err
	}

	if response, ok := result.(*structures.TMDBPersonResponse); ok {
		return *response, nil
	}

	return c.tmdb.GetPerson(personID)
}

func (c *TMDBService) GetSeasonDetails(seriesID string, seasonNumber string) (structures.SeasonDetails, error) {
	params := map[string]interface{}{"series_id": seriesID, "season_number": seasonNumber}

	result, err := c.getCachedOrFetch(fmt.Sprintf("tv/%s/season/%s", seriesID, seasonNumber), params, func() (interface{}, error) {
		return c.tmdb.GetSeasonDetails(seriesID, seasonNumber)
	})
	if err != nil {
		return structures.SeasonDetails{}, err
	}

	if response, ok := result.(*structures.SeasonDetails); ok {
		return *response, nil
	}

	return c.tmdb.GetSeasonDetails(seriesID, seasonNumber)
}

func (c *TMDBService) DiscoverTV(params structures.DiscoverTVParams) (structures.TMDBMediaResponse, error) {
	cacheParams := map[string]interface{}{
		"params": params,
	}

	result, err := c.getCachedOrFetch("discover/tv", cacheParams, func() (interface{}, error) {
		return c.tmdb.DiscoverTV(params)
	})
	if err != nil {
		return structures.TMDBMediaResponse{}, err
	}

	if response, ok := result.(*structures.TMDBMediaResponse); ok {
		return *response, nil
	}

	return c.tmdb.DiscoverTV(params)
}
