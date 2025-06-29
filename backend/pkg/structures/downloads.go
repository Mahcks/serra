package structures

type Download struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	TorrentTitle string  `json:"torrent_title"`
	Source       string  `json:"source"`
	TmdbID       *int64  `json:"tmdb_id,omitempty"`
	TvdbID       *int64  `json:"tvdb_id,omitempty"`
	Hash         *string `json:"hash,omitempty"`
	Progress     float64 `json:"progress"`
	TimeLeft     *string `json:"time_left,omitempty"`
	Status       *string `json:"status,omitempty"`
	UpdatedAt    *string `json:"update_at,omitempty"`
}
