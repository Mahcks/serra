package structures

// Request represents a media request made by a user
type Request struct {
	ID             int64                 `json:"id"`
	UserID         string                `json:"user_id"`
	Username       string                `json:"username,omitempty"`
	MediaType      string                `json:"media_type"`
	TmdbID         *int64                `json:"tmdb_id,omitempty"`
	Title          string                `json:"title"`
	Status         string                `json:"status"`
	Notes          string                `json:"notes,omitempty"`
	CreatedAt      string                `json:"created_at"`
	UpdatedAt      string                `json:"updated_at"`
	FulfilledAt    string                `json:"fulfilled_at,omitempty"`
	ApproverID     string                `json:"approver_id,omitempty"`
	OnBehalfOf     string                `json:"on_behalf_of,omitempty"`
	PosterURL      string                `json:"poster_url,omitempty"`
	Seasons        []int                 `json:"seasons,omitempty"`        // For TV shows - which seasons were requested
	SeasonStatuses map[string]SeasonInfo `json:"season_statuses,omitempty"` // Status of each season
}

// CreateRequestRequest represents a request to create a new media request
type CreateRequestRequest struct {
	MediaType   string  `json:"media_type" validate:"required,oneof=movie tv"`
	TmdbID      int64   `json:"tmdb_id" validate:"required,min=1"`
	Title       string  `json:"title" validate:"required,min=1"`
	Notes       *string `json:"notes,omitempty"`
	PosterURL   *string `json:"poster_url,omitempty"`
	OnBehalfOf  *string `json:"on_behalf_of,omitempty"`
	Seasons     []int   `json:"seasons,omitempty"`     // For TV shows - which seasons to request
}

// UpdateRequestRequest represents a request to update an existing media request
type UpdateRequestRequest struct {
	Status string  `json:"status" validate:"required,oneof=pending approved denied fulfilled"`
	Notes  *string `json:"notes,omitempty"`
}

// RequestStatistics represents statistics about requests in the system
type RequestStatistics struct {
	TotalRequests    int64 `json:"total_requests"`
	PendingRequests  int64 `json:"pending_requests"`
	ApprovedRequests int64 `json:"approved_requests"`
	DeniedRequests   int64 `json:"denied_requests"`
	FulfilledRequests int64 `json:"fulfilled_requests"`
}

// GetAllRequestsResponse represents the response for getting all requests
type GetAllRequestsResponse struct {
	Total    int64     `json:"total"`
	Requests []Request `json:"requests"`
}

// SeasonInfo represents the status of a specific season
type SeasonInfo struct {
	Status           string `json:"status"`            // "pending", "approved", "fulfilled", "partial"
	Episodes         string `json:"episodes"`          // "available/total" e.g., "10/10" or "5/12"
	AvailableEpisodes int   `json:"available_episodes"` // Number of episodes available
	TotalEpisodes     int   `json:"total_episodes"`     // Total episodes in season
	LastUpdated       string `json:"last_updated"`      // When status was last updated
}

// SeasonAvailability represents what's available in the media server
type SeasonAvailability struct {
	ID                int    `json:"id"`
	TmdbID            int    `json:"tmdb_id"`
	SeasonNumber      int    `json:"season_number"`
	EpisodeCount      int    `json:"episode_count"`
	AvailableEpisodes int    `json:"available_episodes"`
	IsComplete        bool   `json:"is_complete"`
	LastUpdated       string `json:"last_updated"`
}

// ShowAvailability represents the overall availability of a TV show
type ShowAvailability struct {
	TmdbID        int                  `json:"tmdb_id"`
	Title         string               `json:"title"`
	TotalSeasons  int                  `json:"total_seasons"`
	Seasons       []SeasonAvailability `json:"seasons"`
	OverallStatus string               `json:"overall_status"` // "not_available", "partial", "complete"
}

// SeasonRequest represents a request for specific seasons
type SeasonRequest struct {
	TmdbID      int   `json:"tmdb_id"`
	SeasonNumbers []int `json:"season_numbers"`
}

// SeasonStatusUpdate represents an update to season status
type SeasonStatusUpdate struct {
	TmdbID            int    `json:"tmdb_id"`
	SeasonNumber      int    `json:"season_number"`
	Status            string `json:"status"`
	AvailableEpisodes int    `json:"available_episodes"`
	TotalEpisodes     int    `json:"total_episodes"`
}