package structures

import "time"

type CalendarItem struct {
	Title       string      `json:"title"`
	Source      ArrProvider `json:"source"` // "radarr" or "sonarr"
	ReleaseDate time.Time   `json:"releaseDate"`
	TmdbID      int64       `json:"tmdb_id"`
}
