package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/mahcks/serra/pkg/structures"
)

type Service interface {
	GetTrendingMedia(page string) (structures.TMDBMediaResponse, error)

	SearchTV(query, page string) (structures.TMDBMediaResponse, error)
	GetTVPopular(page string) (structures.TMDBMediaResponse, error)
	GetTVWatchProviders(seriesID string) (structures.TMDBWatchProvidersResponse, error)
	GetTvRecommendations(seriesID string, page string) (structures.TMDBMediaResponse, error)

	SearchMovie(query, page string) (structures.TMDBMediaResponse, error)
	DiscoverMovie(params structures.DiscoverMovieParams) (structures.TMDBMediaResponse, error)
	GetMoviePopular(page string) (structures.TMDBMediaResponse, error)
	GetMovieWatchProviders(movieID string) (structures.TMDBWatchProvidersResponse, error)
	GetMovieRecommendations(movieID, page string) (structures.TMDBMediaResponse, error)
}

type tmdbService struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type Options struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration // optional: configurable timeout
}

func New(opts Options) (Service, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("TMDB API key is required")
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "https://api.themoviedb.org/3"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	return &tmdbService{
		baseURL: opts.BaseURL,
		apiKey:  opts.APIKey,
		client: &http.Client{
			Timeout: opts.Timeout,
		},
	}, nil
}


// GetTrendingMedia fetches trending media items from TMDB.
func (t *tmdbService) GetTrendingMedia(page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/trending/all/day", map[string]string{"page": page})
}

// SearchTV searches for TV series based on a query string and page number.
func (t *tmdbService) SearchTV(query, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/search/tv", map[string]string{
		"query":    query,
		"page":     page,
		"language": "en-US",
	})
}

func (t *tmdbService) GetTVPopular(page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/tv/popular", map[string]string{
		"page":          page,
		"language":      "en-US",
		"include_adult": "false",
		"sort_by":       "popularity.desc",
	})
}

func (t *tmdbService) GetTVWatchProviders(seriesID string) (structures.TMDBWatchProvidersResponse, error) {
	return t.makeWatchProvidersRequest("/tv/" + seriesID + "/watch/providers")
}

func (t *tmdbService) GetTvRecommendations(seriesID, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/tv/"+seriesID+"/recommendations", map[string]string{
		"page":     page,
		"language": "en-US",
	})
}

func (t *tmdbService) SearchMovie(query, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/search/movie", map[string]string{
		"query":    query,
		"page":     page,
		"language": "en-US",
	})
}

func (t *tmdbService) DiscoverMovie(params structures.DiscoverMovieParams) (structures.TMDBMediaResponse, error) {
	v, err := query.Values(params)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to encode query params: %w", err)
	}

	v.Set("api_key", t.apiKey)
	endpoint := t.baseURL + "/discover/movie?" + v.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to discover movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to discover movies: %s", resp.Status)
	}

	var result structures.TMDBMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetMoviePopular fetches popular movies from TMDB.
func (t *tmdbService) GetMoviePopular(page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/movie/popular", map[string]string{
		"page":          page,
		"language":      "en-US",
		"include_adult": "false",
		"sort_by":       "popularity.desc",
	})
}

// GetMovieWatchProviders fetches watch providers for a specific movie by its ID.
func (t *tmdbService) GetMovieWatchProviders(movieID string) (structures.TMDBWatchProvidersResponse, error) {
	return t.makeWatchProvidersRequest("/movie/" + movieID + "/watch/providers")
}

// GetMovieRecommendations fetches movie recommendations based on a given movie ID and page number.
func (t *tmdbService) GetMovieRecommendations(movieID, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/movie/"+movieID+"/recommendations", map[string]string{
		"page": page,
	})
}

func (t *tmdbService) makeRequest(endpoint string, params map[string]string) (structures.TMDBMediaResponse, error) {
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBMediaResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (t *tmdbService) makeWatchProvidersRequest(endpoint string) (structures.TMDBWatchProvidersResponse, error) {
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBWatchProvidersResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBWatchProvidersResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBWatchProvidersResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBWatchProvidersResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBWatchProvidersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBWatchProvidersResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
