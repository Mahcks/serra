package structures

// Request represents a media request made by a user
type Request struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	Username    string `json:"username,omitempty"`
	MediaType   string `json:"media_type"`
	TmdbID      *int64 `json:"tmdb_id,omitempty"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Notes       string `json:"notes,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	FulfilledAt string `json:"fulfilled_at,omitempty"`
	ApproverID  string `json:"approver_id,omitempty"`
	OnBehalfOf  string `json:"on_behalf_of,omitempty"`
	PosterURL   string `json:"poster_url,omitempty"`
}

// CreateRequestRequest represents a request to create a new media request
type CreateRequestRequest struct {
	MediaType   string  `json:"media_type" validate:"required,oneof=movie tv"`
	TmdbID      int64   `json:"tmdb_id" validate:"required,min=1"`
	Title       string  `json:"title" validate:"required,min=1"`
	Notes       *string `json:"notes,omitempty"`
	PosterURL   *string `json:"poster_url,omitempty"`
	OnBehalfOf  *string `json:"on_behalf_of,omitempty"`
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