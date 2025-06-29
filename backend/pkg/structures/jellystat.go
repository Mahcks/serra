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
