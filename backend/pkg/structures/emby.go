package structures

type EmbyMediaItem struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	OriginalTitle           string `json:"original_title,omitempty"`
	Type                    string `json:"type"`
	ParentID                string `json:"parent_id,omitempty"`
	SeriesID                string `json:"series_id,omitempty"`
	SeasonNumber            int    `json:"season_number,omitempty"`
	EpisodeNumber           int    `json:"episode_number,omitempty"`
	Year                    int    `json:"year,omitempty"`
	PremiereDate            string `json:"premiere_date,omitempty"`
	EndDate                 string `json:"end_date,omitempty"`
	CommunityRating         float64 `json:"community_rating,omitempty"`
	CriticRating            float64 `json:"critic_rating,omitempty"`
	OfficialRating          string `json:"official_rating,omitempty"`
	Overview                string `json:"overview,omitempty"`
	Tagline                 string `json:"tagline,omitempty"`
	Genres                  []string `json:"genres,omitempty"`
	Studios                 []string `json:"studios,omitempty"`
	People                  []EmbyPerson `json:"people,omitempty"`
	TmdbID                  string `json:"tmdb_id,omitempty"`
	ImdbID                  string `json:"imdb_id,omitempty"`
	TvdbID                  string `json:"tvdb_id,omitempty"`
	MusicBrainzID           string `json:"musicbrainz_id,omitempty"`
	Path                    string `json:"path,omitempty"`
	Container               string `json:"container,omitempty"`
	SizeBytes               int64  `json:"size_bytes,omitempty"`
	Bitrate                 int    `json:"bitrate,omitempty"`
	Width                   int    `json:"width,omitempty"`
	Height                  int    `json:"height,omitempty"`
	AspectRatio             string `json:"aspect_ratio,omitempty"`
	VideoCodec              string `json:"video_codec,omitempty"`
	AudioCodec              string `json:"audio_codec,omitempty"`
	SubtitleTracks          []EmbyMediaTrack `json:"subtitle_tracks,omitempty"`
	AudioTracks             []EmbyMediaTrack `json:"audio_tracks,omitempty"`
	RuntimeTicks            int64  `json:"runtime_ticks,omitempty"`
	RuntimeMinutes          int    `json:"runtime_minutes,omitempty"`
	IsFolder                bool   `json:"is_folder,omitempty"`
	IsResumable             bool   `json:"is_resumable,omitempty"`
	PlayCount               int    `json:"play_count,omitempty"`
	DateCreated             string `json:"date_created,omitempty"`
	DateModified            string `json:"date_modified,omitempty"`
	LastPlayedDate          string `json:"last_played_date,omitempty"`
	UserData                map[string]interface{} `json:"user_data,omitempty"`
	ChapterImagesExtracted  bool   `json:"chapter_images_extracted,omitempty"`
	PrimaryImageTag         string `json:"primary_image_tag,omitempty"`
	BackdropImageTags       []string `json:"backdrop_image_tags,omitempty"`
	LogoImageTag            string `json:"logo_image_tag,omitempty"`
	ArtImageTag             string `json:"art_image_tag,omitempty"`
	ThumbImageTag           string `json:"thumb_image_tag,omitempty"`
	IsHD                    bool   `json:"is_hd,omitempty"`
	Is4K                    bool   `json:"is_4k,omitempty"`
	Is3D                    bool   `json:"is_3d,omitempty"`
	Locked                  bool   `json:"locked,omitempty"`
	ProviderIds             map[string]string `json:"provider_ids,omitempty"`
	ExternalUrls            map[string]string `json:"external_urls,omitempty"`
	Tags                    []string `json:"tags,omitempty"`
	SortName                string `json:"sort_name,omitempty"`
	ForcedSortName          string `json:"forced_sort_name,omitempty"`
	
	// Legacy fields for backwards compatibility
	Poster                  string `json:"poster,omitempty"`
}

type EmbyPerson struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Type string `json:"type"` // Actor, Director, Producer, etc.
}

type EmbyMediaTrack struct {
	Index    int    `json:"index"`
	Language string `json:"language,omitempty"`
	Codec    string `json:"codec,omitempty"`
	Title    string `json:"title,omitempty"`
	IsDefault bool  `json:"is_default,omitempty"`
	IsForced bool   `json:"is_forced,omitempty"`
}
