package structures

type SonarrQualityProfile struct {
	ID                    int                        `json:"id"`
	Name                  string                     `json:"name"`
	UpgradeAllowed        bool                       `json:"upgrade_allowed"`
	Cutoff                int                        `json:"cutoff"`
	Items                 []SonarrQualityProfileItem `json:"items"`
	MinFormatScore        int                        `json:"min_format_score"`
	CutoffFormatScore     int                        `json:"cutoff_format_score"`
	MinUpgradeFormatScore int                        `json:"min_upgrade_format_score"`
	FormatItems           []SonarrFormatItem         `json:"format_items"`
	Language              SonarrLanguage             `json:"language"`
}

type SonarrQualityProfileItem struct {
	Quality SonarrQuality              `json:"quality"`
	Items   []SonarrQualityProfileItem `json:"items"`
	Allowed bool                       `json:"allowed"`
	Name    string                     `json:"name,omitempty"` // Only present for grouped items
	ID      int                        `json:"id,omitempty"`   // Only present for grouped items
}

type SonarrQuality struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

type SonarrFormatItem struct {
	Format int    `json:"format"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
}

type SonarrLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SonarrRootFolder struct {
	Path            string                 `json:"path"`
	Accessible      bool                   `json:"accessible"`
	FreeSpace       int64                  `json:"free_space"`
	UnmappedFolders []SonarrUnmappedFolder `json:"unmapped_folders,omitempty"`
}

type SonarrUnmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
}
