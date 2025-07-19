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
	DiscoverTV(params structures.DiscoverTVParams) (structures.TMDBMediaResponse, error)
	GetTVPopular(page string) (structures.TMDBMediaResponse, error)
	GetTVUpcoming(page string) (structures.TMDBMediaResponse, error)
	GetTVWatchProviders(seriesID string) (structures.TMDBWatchProvidersResponse, error)
	GetTvRecommendations(seriesID string, page string) (structures.TMDBMediaResponse, error)
	GetTvSimilar(seriesID string, page string) (structures.TMDBMediaResponse, error)
	GetSeasonDetails(seriesID string, seasonNumber string) (structures.SeasonDetails, error)

	SearchMovie(query, page string) (structures.TMDBMediaResponse, error)
	DiscoverMovie(params structures.DiscoverMovieParams) (structures.TMDBMediaResponse, error)
	GetMoviePopular(page string) (structures.TMDBMediaResponse, error)
	GetMovieUpcoming(page string) (structures.TMDBMediaResponse, error)
	GetMovieWatchProviders(movieID string) (structures.TMDBWatchProvidersResponse, error)
	GetMovieRecommendations(movieID, page string) (structures.TMDBMediaResponse, error)
	GetMovieSimilar(movieID, page string) (structures.TMDBMediaResponse, error)
	GetMovieReleaseDates(movieID string) (structures.TMDBReleaseDatesResponse, error)

	// Watch providers
	GetWatchProviders(mediaType string) (structures.TMDBWatchProvidersListResponse, error)
	GetWatchProviderRegions() (structures.TMDBWatchProviderRegionsResponse, error)

	// Company search
	SearchCompanies(query, page string) (structures.TMDBCompanySearchResponse, error)

	// Collection details
	GetCollection(collectionID string) (structures.TMDBCollectionResponse, error)

	// Person details
	GetPerson(personID string) (structures.TMDBPersonResponse, error)
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

// GetTVUpcoming fetches TV shows on the air (upcoming episodes) from TMDB.
func (t *tmdbService) GetTVUpcoming(page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/tv/on_the_air", map[string]string{
		"page":     page,
		"language": "en-US",
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

func (t *tmdbService) GetTvSimilar(seriesID, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/tv/"+seriesID+"/similar", map[string]string{
		"page":     page,
		"language": "en-US",
	})
}

func (t *tmdbService) GetSeasonDetails(seriesID string, seasonNumber string) (structures.SeasonDetails, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/tv/%s/season/%s?language=en-US&api_key=%s", seriesID, seasonNumber, t.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return structures.SeasonDetails{}, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return structures.SeasonDetails{}, fmt.Errorf("failed to fetch season details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.SeasonDetails{}, fmt.Errorf("TMDB API returned %d", resp.StatusCode)
	}

	var result structures.SeasonDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.SeasonDetails{}, fmt.Errorf("failed to decode TMDB season response: %w", err)
	}

	return result, nil
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

func (t *tmdbService) DiscoverTV(params structures.DiscoverTVParams) (structures.TMDBMediaResponse, error) {
	v, err := query.Values(params)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to encode query params: %w", err)
	}

	v.Set("api_key", t.apiKey)
	endpoint := t.baseURL + "/discover/tv?" + v.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to discover TV shows: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBMediaResponse{}, fmt.Errorf("failed to discover TV shows: %s", resp.Status)
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

// GetMovieUpcoming fetches upcoming movies from TMDB.
func (t *tmdbService) GetMovieUpcoming(page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/movie/upcoming", map[string]string{
		"page":     page,
		"language": "en-US",
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

// GetMovieSimilar fetches movies similar to a given movie ID and page number.
func (t *tmdbService) GetMovieSimilar(movieID, page string) (structures.TMDBMediaResponse, error) {
	return t.makeRequest("/movie/"+movieID+"/similar", map[string]string{
		"page": page,
	})
}

// GetMovieReleaseDates fetches release dates for a movie across different countries.
func (t *tmdbService) GetMovieReleaseDates(movieID string) (structures.TMDBReleaseDatesResponse, error) {
	return t.makeReleaseDatesRequest("/movie/" + movieID + "/release_dates")
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

func (t *tmdbService) makeReleaseDatesRequest(endpoint string) (structures.TMDBReleaseDatesResponse, error) {
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBReleaseDatesResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBReleaseDatesResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBReleaseDatesResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBReleaseDatesResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBReleaseDatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBReleaseDatesResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetWatchProviders fetches the list of watch providers for movies or TV shows.
func (t *tmdbService) GetWatchProviders(mediaType string) (structures.TMDBWatchProvidersListResponse, error) {
	endpoint := fmt.Sprintf("/watch/providers/%s", mediaType)
	
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBWatchProvidersListResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBWatchProvidersListResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBWatchProvidersListResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBWatchProvidersListResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBWatchProvidersListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBWatchProvidersListResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetWatchProviderRegions fetches the list of regions where watch providers are available.
func (t *tmdbService) GetWatchProviderRegions() (structures.TMDBWatchProviderRegionsResponse, error) {
	endpoint := "/watch/providers/regions"
	
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBWatchProviderRegionsResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBWatchProviderRegionsResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBWatchProviderRegionsResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBWatchProviderRegionsResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBWatchProviderRegionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBWatchProviderRegionsResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// SearchCompanies searches for companies based on a query string and page number.
func (t *tmdbService) SearchCompanies(query, page string) (structures.TMDBCompanySearchResponse, error) {
	u, err := url.Parse(t.baseURL + "/search/company")
	if err != nil {
		return structures.TMDBCompanySearchResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	q.Set("query", query)
	q.Set("page", page)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBCompanySearchResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBCompanySearchResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBCompanySearchResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBCompanySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBCompanySearchResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetCollection fetches collection details and movies from TMDB.
func (t *tmdbService) GetCollection(collectionID string) (structures.TMDBCollectionResponse, error) {
	endpoint := "/collection/" + collectionID
	
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBCollectionResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	q.Set("language", "en-US")
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBCollectionResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBCollectionResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBCollectionResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBCollectionResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPerson fetches person details, movie credits, and TV credits from TMDB.
func (t *tmdbService) GetPerson(personID string) (structures.TMDBPersonResponse, error) {
	endpoint := "/person/" + personID
	
	u, err := url.Parse(t.baseURL + endpoint)
	if err != nil {
		return structures.TMDBPersonResponse{}, fmt.Errorf("invalid endpoint: %w", err)
	}

	q := u.Query()
	q.Set("api_key", t.apiKey)
	q.Set("language", "en-US")
	q.Set("append_to_response", "movie_credits,tv_credits")
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), t.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return structures.TMDBPersonResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return structures.TMDBPersonResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return structures.TMDBPersonResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result structures.TMDBPersonResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return structures.TMDBPersonResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
