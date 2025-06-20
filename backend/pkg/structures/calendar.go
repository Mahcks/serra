package structures

import "time"

type CalendarItem struct {
	Title       string      `json:"title"`
	Source      ArrProvider `json:"source"` // "radarr" or "sonarr"
	ReleaseDate time.Time   `json:"releaseDate"`
}
