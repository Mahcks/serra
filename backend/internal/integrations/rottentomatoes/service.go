package rottentomatoes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Service defines the interface for Rotten Tomatoes operations
type Service interface {
	GetRottenTomatoesScore(ctx context.Context, title string, year int, mediaType string) (*RottenTomatoesResponse, error)
}

// service implements the Service interface
type service struct {
	client *http.Client
}

// RottenTomatoesResponse represents the response structure for RT scores
type RottenTomatoesResponse struct {
	Title          string `json:"title"`
	Year           int    `json:"year"`
	Type           string `json:"type"`
	TomatoMeter    int    `json:"tomato_meter"`
	URL            string `json:"url"`
	CriticsRating  string `json:"critics_rating"`
	CriticsScore   int    `json:"critics_score"`
	AudienceRating string `json:"audience_rating"`
	AudienceScore  int    `json:"audience_score"`
}

// Internal structures for Algolia API response
type algoliaRTResponse struct {
	Results []struct {
		Hits []struct {
			Title          string `json:"title"`
			ReleaseYear    int    `json:"releaseYear"`
			Type           string `json:"type"`
			Vanity         string `json:"vanity"`
			RottenTomatoes *struct {
				CriticsScore   int  `json:"criticsScore"`
				AudienceScore  int  `json:"audienceScore"`
				CertifiedFresh bool `json:"certifiedFresh"`
			} `json:"rottenTomatoes"`
		} `json:"hits"`
		Index string `json:"index"`
	} `json:"results"`
}

// NewService creates a new Rotten Tomatoes service
func NewService() Service {
	return &service{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRottenTomatoesScore fetches RT scores for a given title and year
func (s *service) GetRottenTomatoesScore(ctx context.Context, title string, year int, mediaType string) (*RottenTomatoesResponse, error) {
	// Clean up the title for better search results
	query := cleanTitle(title)

	// Add year to query if provided
	if year > 0 {
		query = fmt.Sprintf("%s %d", query, year)
	}

	var filters string
	if mediaType == "tv" {
		filters = "isEmsSearchable=1 AND type:\"tv\""
	} else {
		filters = "isEmsSearchable=1 AND type:\"movie\""
	}

	payload := map[string]any{
		"requests": []map[string]any{
			{
				"indexName": "content_rt",
				"query":     query,
				"params":    fmt.Sprintf("filters=%s&hitsPerPage=20", url.QueryEscape(filters)),
			},
		},
	}

	// Convert to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://79frdp12pn-dsn.algolia.net/1/indexes/*/queries", strings.NewReader(string(jsonPayload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Algolia-Agent", "Algolia%20for%20JavaScript%20(4.14.3)%3B%20Browser%20(lite)")
	req.Header.Set("X-Algolia-API-Key", "175588f6e5f8319b27702e4cc4013561")
	req.Header.Set("X-Algolia-Application-Id", "79FRDP12PN")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.rottentomatoes.com/")

	// Make the request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var algoliaResp algoliaRTResponse
	if err := json.Unmarshal(body, &algoliaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find content_rt results
	var contentHits []struct {
		Title          string `json:"title"`
		ReleaseYear    int    `json:"releaseYear"`
		Type           string `json:"type"`
		Vanity         string `json:"vanity"`
		RottenTomatoes *struct {
			CriticsScore   int  `json:"criticsScore"`
			AudienceScore  int  `json:"audienceScore"`
			CertifiedFresh bool `json:"certifiedFresh"`
		} `json:"rottenTomatoes"`
	}

	for _, result := range algoliaResp.Results {
		if result.Index == "content_rt" {
			contentHits = result.Hits
			break
		}
	}

	if len(contentHits) == 0 {
		return nil, fmt.Errorf("no Rotten Tomatoes results found for %s", title)
	}

	// Find the best match
	bestMatch := findBestMatch(contentHits, query, year)
	if bestMatch == nil {
		return nil, fmt.Errorf("no matching results found for %s", title)
	}

	// Convert to response format
	tomatoMeter := 0
	audienceScore := 0
	criticsScore := 0
	if bestMatch.RottenTomatoes != nil {
		tomatoMeter = bestMatch.RottenTomatoes.CriticsScore
		audienceScore = bestMatch.RottenTomatoes.AudienceScore
		criticsScore = bestMatch.RottenTomatoes.CriticsScore
	}

	// Generate critics and audience ratings based on scores
	criticsRating := getCriticsRating(tomatoMeter, bestMatch.RottenTomatoes != nil && bestMatch.RottenTomatoes.CertifiedFresh)
	audienceRating := getAudienceRating(audienceScore)

	return &RottenTomatoesResponse{
		TomatoMeter:    tomatoMeter,
		AudienceScore:  audienceScore,
		URL:            fmt.Sprintf("https://www.rottentomatoes.com/m/%s", bestMatch.Vanity),
		CriticsRating:  criticsRating,
		CriticsScore:   criticsScore,
		AudienceRating: audienceRating,
		Title:          bestMatch.Title,
		Year:           bestMatch.ReleaseYear,
		Type:           bestMatch.Type,
	}, nil
}

// cleanTitle cleans up the title for better search results
func cleanTitle(title string) string {
	// Replace underscores with spaces
	cleaned := strings.ReplaceAll(title, "_", " ")

	// Capitalize first letter of each word
	words := strings.Fields(cleaned)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// findBestMatch finds the best matching result from the hits
func findBestMatch(hits []struct {
	Title          string `json:"title"`
	ReleaseYear    int    `json:"releaseYear"`
	Type           string `json:"type"`
	Vanity         string `json:"vanity"`
	RottenTomatoes *struct {
		CriticsScore   int  `json:"criticsScore"`
		AudienceScore  int  `json:"audienceScore"`
		CertifiedFresh bool `json:"certifiedFresh"`
	} `json:"rottenTomatoes"`
}, query string, searchYear int) *struct {
	Title          string `json:"title"`
	ReleaseYear    int    `json:"releaseYear"`
	Type           string `json:"type"`
	Vanity         string `json:"vanity"`
	RottenTomatoes *struct {
		CriticsScore   int  `json:"criticsScore"`
		AudienceScore  int  `json:"audienceScore"`
		CertifiedFresh bool `json:"certifiedFresh"`
	} `json:"rottenTomatoes"`
} {
	if len(hits) == 0 {
		return nil
	}

	// If year is provided, try to find exact year match first
	if searchYear > 0 {
		for _, hit := range hits {
			if hit.ReleaseYear == searchYear && hit.RottenTomatoes != nil {
				return &hit
			}
		}
	}

	// Return first match with ratings if no year match
	for _, hit := range hits {
		if hit.RottenTomatoes != nil {
			return &hit
		}
	}

	// Return first match even without ratings
	return &hits[0]
}

// getCriticsRating returns the critics rating based on tomato meter score
func getCriticsRating(tomatoMeter int, certifiedFresh bool) string {
	if certifiedFresh {
		return "Certified Fresh"
	}
	if tomatoMeter >= 60 {
		return "Fresh"
	}
	return "Rotten"
}

// getAudienceRating returns the audience rating based on audience score
func getAudienceRating(audienceScore int) string {
	if audienceScore >= 60 {
		return "Upright"
	}
	return "Spilled"
}
