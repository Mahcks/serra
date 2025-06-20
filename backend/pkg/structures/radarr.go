package structures

type RadarrQualityProfile struct {
	ID                    int                        `json:"id"`
	Name                  string                     `json:"name"`
	UpgradeAllowed        bool                       `json:"upgrade_allowed"`
	Cutoff                int                        `json:"cutoff"`
	Items                 []RadarrQualityProfileItem `json:"items"`
	MinFormatScore        int                        `json:"min_format_score"`
	CutoffFormatScore     int                        `json:"cutoff_format_score"`
	MinUpgradeFormatScore int                        `json:"min_upgrade_format_score"`
	FormatItems           []RadarrFormatItem         `json:"format_items"`
	Language              RadarrLanguage             `json:"language"`
}

type RadarrQualityProfileItem struct {
	Quality RadarrQuality              `json:"quality"`
	Items   []RadarrQualityProfileItem `json:"items"`
	Allowed bool                       `json:"allowed"`
	Name    string                     `json:"name,omitempty"` // Only present for grouped items
	ID      int                        `json:"id,omitempty"`   // Only present for grouped items
}

type RadarrQuality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

type RadarrFormatItem struct {
	Format int    `json:"format"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
}

type RadarrLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RadarrRootFolder struct {
	Path            string                 `json:"path"`
	Accessible      bool                   `json:"accessible"`
	FreeSpace       int64                  `json:"free_space"`
	UnmappedFolders []RadarrUnmappedFolder `json:"unmapped_folders,omitempty"`
}

type RadarrUnmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
}
