package structures

import "time"

// Analytics API Response Types
// These types are used for API responses and use the SQLC-generated types as their foundation

// RequestAnalyticsResponse represents analytics data for request performance
type RequestAnalyticsResponse struct {
	ID                       int64     `json:"id"`
	TmdbID                   int64     `json:"tmdb_id"`
	MediaType                string    `json:"media_type"`
	Title                    string    `json:"title"`
	RequestCount             *int64    `json:"request_count"`
	LastRequested            *time.Time `json:"last_requested"`
	FirstRequested           *time.Time `json:"first_requested"`
	AvgProcessingTimeSeconds *int64    `json:"avg_processing_time_seconds"`
	SuccessRate              *float64  `json:"success_rate"`
	PopularityScore          *float64  `json:"popularity_score"`
	CreatedAt                *time.Time `json:"created_at"`
	UpdatedAt                *time.Time `json:"updated_at"`
}

// RequestMetricResponse represents metrics for individual request tracking
type RequestMetricResponse struct {
	ID                    int64     `json:"id"`
	RequestID             int64     `json:"request_id"`
	StatusChange          string    `json:"status_change"`
	PreviousStatus        *string   `json:"previous_status"`
	NewStatus             string    `json:"new_status"`
	ProcessingTimeSeconds *int64    `json:"processing_time_seconds"`
	ErrorCode             *int64    `json:"error_code"`
	ErrorMessage          *string   `json:"error_message"`
	UserID                string    `json:"user_id"`
	Timestamp             *time.Time `json:"timestamp"`
}

// DriveUsageHistoryResponse represents drive usage tracking data
type DriveUsageHistoryResponse struct {
	ID                 int64     `json:"id"`
	DriveID            string    `json:"drive_id"`
	TotalSize          int64     `json:"total_size"`
	UsedSize           int64     `json:"used_size"`
	AvailableSize      int64     `json:"available_size"`
	UsagePercentage    float64   `json:"usage_percentage"`
	GrowthRateGbPerDay *float64  `json:"growth_rate_gb_per_day"`
	ProjectedFullDate  *time.Time `json:"projected_full_date"`
	RecordedAt         *time.Time `json:"recorded_at"`
}

// DriveAlertResponse represents drive alert information
type DriveAlertResponse struct {
	ID                   int64     `json:"id"`
	DriveID              string    `json:"drive_id"`
	AlertType            string    `json:"alert_type"`
	ThresholdValue       float64   `json:"threshold_value"`
	CurrentValue         float64   `json:"current_value"`
	AlertMessage         string    `json:"alert_message"`
	IsActive             *bool     `json:"is_active"`
	LastTriggered        *time.Time `json:"last_triggered"`
	AcknowledgementCount *int64    `json:"acknowledgement_count"`
	CreatedAt            *time.Time `json:"created_at"`
}

// SystemMetricResponse represents system performance metrics
type SystemMetricResponse struct {
	ID          int64     `json:"id"`
	MetricType  string    `json:"metric_type"`
	MetricName  string    `json:"metric_name"`
	MetricValue float64   `json:"metric_value"`
	Metadata    *string   `json:"metadata"`
	RecordedAt  *time.Time `json:"recorded_at"`
}

// PopularityTrendResponse represents content popularity trend data
type PopularityTrendResponse struct {
	ID                 int64     `json:"id"`
	TmdbID             int64     `json:"tmdb_id"`
	MediaType          string    `json:"media_type"`
	Title              string    `json:"title"`
	TrendSource        string    `json:"trend_source"`
	PopularityScore    float64   `json:"popularity_score"`
	TrendDirection     *string   `json:"trend_direction"`
	ForecastConfidence *float64  `json:"forecast_confidence"`
	Metadata           *string   `json:"metadata"`
	ValidUntil         *time.Time `json:"valid_until"`
	CreatedAt          *time.Time `json:"created_at"`
}

// Request Analytics Query Response Types

// RequestProcessingPerformanceResponse represents processing performance by media type and status
type RequestProcessingPerformanceResponse struct {
	MediaType string `json:"media_type"`
	Status    string `json:"status"`
	Count     int64  `json:"count"`
}

// RequestSuccessRatesResponse represents success rates by status
type RequestSuccessRatesResponse struct {
	Status        string  `json:"status"`
	TotalRequests int64   `json:"total_requests"`
	Percentage    float64 `json:"percentage"`
}

// RequestTrendsResponse represents request trends over time
type RequestTrendsResponse struct {
	RequestDate       interface{} `json:"request_date"`
	TotalRequests     int64       `json:"total_requests"`
	ApprovedRequests  int64       `json:"approved_requests"`
	CompletedRequests int64       `json:"completed_requests"`
	FailedRequests    int64       `json:"failed_requests"`
	MovieRequests     int64       `json:"movie_requests"`
	TvRequests        int64       `json:"tv_requests"`
}

// RequestVolumeByHourResponse represents request volume patterns by hour
type RequestVolumeByHourResponse struct {
	HourOfDay          interface{} `json:"hour_of_day"`
	TotalRequests      int64       `json:"total_requests"`
	SuccessfulRequests int64       `json:"successful_requests"`
	SuccessRate        float64     `json:"success_rate"`
}

// PopularRequestedContentResponse represents most requested content
type PopularRequestedContentResponse struct {
	TmdbID          *int64      `json:"tmdb_id"`
	Title           *string     `json:"title"`
	MediaType       string      `json:"media_type"`
	RequestCount    int64       `json:"request_count"`
	FulfilledCount  int64       `json:"fulfilled_count"`
	FailedCount     int64       `json:"failed_count"`
	FulfillmentRate float64     `json:"fulfillment_rate"`
	FirstRequested  interface{} `json:"first_requested"`
	LastRequested   interface{} `json:"last_requested"`
}

// RequestFulfillmentByUserResponse represents fulfillment metrics per user
type RequestFulfillmentByUserResponse struct {
	Username          string  `json:"username"`
	TotalRequests     int64   `json:"total_requests"`
	ApprovedRequests  int64   `json:"approved_requests"`
	CompletedRequests int64   `json:"completed_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	SuccessRate       float64 `json:"success_rate"`
}

// FailureAnalysisResponse represents analysis of failed requests
type FailureAnalysisResponse struct {
	MediaType          string `json:"media_type"`
	TotalFailures      int64  `json:"total_failures"`
	NotFoundFailures   int64  `json:"not_found_failures"`
	ConnectionFailures int64  `json:"connection_failures"`
	QualityFailures    int64  `json:"quality_failures"`
	StorageFailures    int64  `json:"storage_failures"`
}

// ContentAvailabilityVsRequestsResponse represents content popularity vs availability
type ContentAvailabilityVsRequestsResponse struct {
	TmdbID            *int64      `json:"tmdb_id"`
	Title             *string     `json:"title"`
	MediaType         string      `json:"media_type"`
	RequestCount      int64       `json:"request_count"`
	LastRequested     interface{} `json:"last_requested"`
	IsAvailable       interface{} `json:"is_available"`
	FulfilledRequests int64       `json:"fulfilled_requests"`
}

// Drive Analytics Response Types

// LatestDriveUsageResponse represents latest drive usage with drive info
type LatestDriveUsageResponse struct {
	ID                 int64     `json:"id"`
	DriveID            string    `json:"drive_id"`
	TotalSize          int64     `json:"total_size"`
	UsedSize           int64     `json:"used_size"`
	AvailableSize      int64     `json:"available_size"`
	UsagePercentage    float64   `json:"usage_percentage"`
	GrowthRateGbPerDay *float64  `json:"growth_rate_gb_per_day"`
	ProjectedFullDate  *time.Time `json:"projected_full_date"`
	RecordedAt         *time.Time `json:"recorded_at"`
	DriveName          string    `json:"drive_name"`
	DriveMountPath     string    `json:"drive_mount_path"`
}

// DriveGrowthPredictionResponse represents drive growth predictions
type DriveGrowthPredictionResponse struct {
	DriveID               string      `json:"drive_id"`
	AvgGrowthRateGbPerDay *float64    `json:"avg_growth_rate_gb_per_day"`
	CurrentAvgUsage       *float64    `json:"current_avg_usage"`
	EarliestFullDate      interface{} `json:"earliest_full_date"`
}

// ActiveDriveAlertsResponse represents active drive alerts with drive info
type ActiveDriveAlertsResponse struct {
	ID                   int64     `json:"id"`
	DriveID              string    `json:"drive_id"`
	AlertType            string    `json:"alert_type"`
	ThresholdValue       float64   `json:"threshold_value"`
	CurrentValue         float64   `json:"current_value"`
	AlertMessage         string    `json:"alert_message"`
	IsActive             *bool     `json:"is_active"`
	LastTriggered        *time.Time `json:"last_triggered"`
	AcknowledgementCount *int64    `json:"acknowledgement_count"`
	CreatedAt            *time.Time `json:"created_at"`
	DriveName            string    `json:"drive_name"`
	MountPath            string    `json:"mount_path"`
}

// Aggregate Analytics Response Types

// RequestSuccessRateAggregateResponse represents overall success rate metrics
type RequestSuccessRateAggregateResponse struct {
	TotalRequests     int64     `json:"total_requests"`
	FulfilledRequests *float64  `json:"fulfilled_requests"`
	SuccessRate       int64     `json:"success_rate"`
}

// MostRequestedContentResponse represents most requested content with metrics
type MostRequestedContentResponse struct {
	TmdbID          int64     `json:"tmdb_id"`
	MediaType       string    `json:"media_type"`
	Title           string    `json:"title"`
	RequestCount    *int64    `json:"request_count"`
	LastRequested   *time.Time `json:"last_requested"`
	PopularityScore *float64  `json:"popularity_score"`
}