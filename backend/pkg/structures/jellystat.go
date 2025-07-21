package structures

type JellystatLibrary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CollectionType string `json:"collection_type"`
	LibraryCount   int    `json:"library_count"`
	SeasonCount    int    `json:"season_count"`
	EpisodeCount   int    `json:"episode_count"`
}

type JellystatUserActivity struct {
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	TotalPlays     int    `json:"total_plays"`
	TotalWatchTime int    `json:"total_watch_time"`
}

type JellystatPopularContent struct {
	ItemID        string  `json:"item_id"`
	ItemName      string  `json:"item_name"`
	ItemType      string  `json:"item_type"`
	TotalPlays    int     `json:"total_plays"`
	TotalRuntime  int     `json:"total_runtime"`
	AverageRating float64 `json:"average_rating"`
	LibraryName   string  `json:"library_name"`
}

type JellystatRecentlyWatched struct {
	ItemID       string `json:"item_id"`
	ItemName     string `json:"item_name"`
	ItemType     string `json:"item_type"`
	UserName     string `json:"user_name"`
	PlayDuration int    `json:"play_duration"`
	WatchedAt    string `json:"watched_at"`
	LibraryName  string `json:"library_name"`
}

type JellystatWatchHistory struct {
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	ItemID       string `json:"item_id"`
	ItemName     string `json:"item_name"`
	ItemType     string `json:"item_type"`
	PlayDuration int    `json:"play_duration"`
	WatchedAt    string `json:"watched_at"`
	IsCompleted  bool   `json:"is_completed"`
}

type JellystatActiveUser struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Plays    int    `json:"plays"`
}

type JellystatPlaybackMethod struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
