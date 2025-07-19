package structures

type TMDBFullMediaResponse struct {
	TMDBPageResults
	Results []TMDBFullMediaItem `json:"results"`
}

type TMDBFullMediaItem struct {
	TMDBMediaItem
	InLibrary bool `json:"in_library"`
	Requested bool `json:"requested"`
}

// STRUCUTRES FOR TMDB API RESPONSES

type TMDBPageResults struct {
	Page         int64 `json:"page"`
	TotalPages   int64 `json:"total_pages"`
	TotalResults int64 `json:"total_results"`
}

type TMDBMediaResponse struct {
	TMDBPageResults
	Results []TMDBMediaItem `json:"results"`
}

type TMDBMediaItem struct {
	Adult              bool     `json:"adult,omitempty"`
	Gender             int      `json:"gender,omitempty"`
	BackdropPath       string   `json:"backdrop_path,omitempty"`
	GenreIDs           []int64  `json:"genre_ids,omitempty"`
	ID                 int64    `json:"id"`
	OriginalLanguage   string   `json:"original_language"`
	OriginalTitle      string   `json:"original_title,omitempty"`
	Overview           string   `json:"overview,omitempty"`
	PosterPath         string   `json:"poster_path,omitempty"`
	ReleaseDate        string   `json:"release_date,omitempty"`
	Title              string   `json:"title,omitempty"`
	Video              bool     `json:"video,omitempty"`
	VoteAverage        float32  `json:"vote_average,omitempty"`
	VoteCount          int64    `json:"vote_count,omitempty"`
	Popularity         float32  `json:"popularity,omitempty"`
	FirstAirDate       string   `json:"first_air_date,omitempty"`
	Name               string   `json:"name,omitempty"`
	OriginCountry      []string `json:"origin_country,omitempty"`
	OriginalName       string   `json:"original_name,omitempty"`
	KnownForDepartment string   `json:"known_for_department,omitempty"`
	ProfilePath        string   `json:"profile_path,omitempty"`
	MediaType          string   `json:"media_type,omitempty"`
	KnownFor           []struct {
		Adult            bool    `json:"adult"`
		BackdropPath     string  `json:"backdrop_path"`
		GenreIds         []int   `json:"genre_ids"`
		ID               int     `json:"id"`
		OriginalLanguage string  `json:"original_language"`
		OriginalTitle    string  `json:"original_title"`
		Overview         string  `json:"overview"`
		PosterPath       string  `json:"poster_path"`
		ReleaseDate      string  `json:"release_date"`
		Title            string  `json:"title"`
		Video            bool    `json:"video"`
		VoteAverage      float64 `json:"vote_average"`
		VoteCount        int     `json:"vote_count"`
		Popularity       float64 `json:"popularity"`
		MediaType        string  `json:"media_type"`
	} `json:"known_for,omitempty"`
}

type TVDetails struct {
	Adult               bool                `json:"adult"`
	BackdropPath        string              `json:"backdrop_path"`
	CreatedBy           []CreatedBy         `json:"created_by"`
	Credits             Credits             `json:"credits"`
	EpisodeRunTime      []int               `json:"episode_run_time"`
	FirstAirDate        string              `json:"first_air_date"`
	Genres              []Genre             `json:"genres"`
	Homepage            string              `json:"homepage"`
	ID                  int                 `json:"id"`
	InProduction        bool                `json:"in_production"`
	Languages           []string            `json:"languages"`
	LastAirDate         string              `json:"last_air_date"`
	LastEpisodeToAir    Episode             `json:"last_episode_to_air"`
	Name                string              `json:"name"`
	Networks            []Network           `json:"networks"`
	NextEpisodeToAir    interface{}         `json:"next_episode_to_air"`
	NumberOfEpisodes    int                 `json:"number_of_episodes"`
	NumberOfSeasons     int                 `json:"number_of_seasons"`
	OriginCountry       []string            `json:"origin_country"`
	OriginalLanguage    string              `json:"original_language"`
	OriginalName        string              `json:"original_name"`
	Overview            string              `json:"overview"`
	Popularity          float64             `json:"popularity"`
	PosterPath          string              `json:"poster_path"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	Seasons             []Season            `json:"seasons"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Type                string              `json:"type"`
	Videos              Videos              `json:"videos"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`
}

type CreatedBy struct {
	CreditID     string `json:"credit_id"`
	Gender       int    `json:"gender"`
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	ProfilePath  string `json:"profile_path"`
}

type Credits struct {
	Cast []CastMember `json:"cast"`
	Crew []CrewMember `json:"crew"`
}

type CastMember struct {
	Adult              bool    `json:"adult"`
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Gender             int     `json:"gender"`
	ID                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Order              int     `json:"order"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}

type CrewMember struct {
	Adult              bool    `json:"adult"`
	CreditID           string  `json:"credit_id"`
	Department         string  `json:"department"`
	Gender             int     `json:"gender"`
	ID                 int     `json:"id"`
	Job                string  `json:"job"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Network struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type ProductionCompany struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type ProductionCountry struct {
	ISO3166_1 string `json:"iso_3166_1"`
	Name      string `json:"name"`
}

type Season struct {
	AirDate      string  `json:"air_date"`
	EpisodeCount int     `json:"episode_count"`
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	SeasonNumber int     `json:"season_number"`
	VoteAverage  float64 `json:"vote_average"`
}

type SpokenLanguage struct {
	EnglishName string `json:"english_name"`
	ISO639_1    string `json:"iso_639_1"`
	Name        string `json:"name"`
}

type Episode struct {
	AirDate        string       `json:"air_date"`
	EpisodeNumber  int          `json:"episode_number"`
	EpisodeType    string       `json:"episode_type"`
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	Overview       string       `json:"overview"`
	ProductionCode string       `json:"production_code"`
	Runtime        int          `json:"runtime"`
	SeasonNumber   int          `json:"season_number"`
	ShowID         int          `json:"show_id"`
	StillPath      string       `json:"still_path"`
	VoteAverage    float64      `json:"vote_average"`
	VoteCount      int          `json:"vote_count"`
	Crew           []CrewMember `json:"crew"`
	GuestStars     []CastMember `json:"guest_stars"`
}
type Videos struct {
	Results []Video `json:"results"`
}

type Video struct {
	ID          string `json:"id"`
	ISO3166_1   string `json:"iso_3166_1"`
	ISO639_1    string `json:"iso_639_1"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Official    bool   `json:"official"`
	PublishedAt string `json:"published_at"`
	Site        string `json:"site"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
}

type MovieDetails struct {
	Adult               bool            `json:"adult"`
	BackdropPath        string          `json:"backdrop_path"`
	BelongsToCollection interface{}     `json:"belongs_to_collection"` // adjust if needed
	Budget              int64           `json:"budget"`
	Genres              []Genre         `json:"genres"`
	Homepage            string          `json:"homepage"`
	ID                  int64           `json:"id"`
	IMDBID              string          `json:"imdb_id"`
	OriginCountry       []string        `json:"origin_country"`
	OriginalLanguage    string          `json:"original_language"`
	OriginalTitle       string          `json:"original_title"`
	Overview            string          `json:"overview"`
	Popularity          float64         `json:"popularity"`
	PosterPath          string          `json:"poster_path"`
	ProductionCompanies []Company       `json:"production_companies"`
	ProductionCountries []Country       `json:"production_countries"`
	ReleaseDate         string          `json:"release_date"`
	Revenue             int64           `json:"revenue"`
	Runtime             int             `json:"runtime"`
	SpokenLanguages     []Language      `json:"spoken_languages"`
	Status              string          `json:"status"`
	Tagline             string          `json:"tagline"`
	Title               string          `json:"title"`
	Video               bool            `json:"video"`
	VoteAverage         float64         `json:"vote_average"`
	VoteCount           int64           `json:"vote_count"`
	Videos              VideosResponse  `json:"videos"`
	Credits             CreditsResponse `json:"credits"`
}

type Company struct {
	ID            int64   `json:"id"`
	LogoPath      *string `json:"logo_path"`
	Name          string  `json:"name"`
	OriginCountry string  `json:"origin_country"`
}

type Country struct {
	ISO3166_1 string `json:"iso_3166_1"`
	Name      string `json:"name"`
}

type Language struct {
	EnglishName string `json:"english_name"`
	ISO639_1    string `json:"iso_639_1"`
	Name        string `json:"name"`
}

type VideosResponse struct {
	Results []Video `json:"results"`
}

type CreditsResponse struct {
	Cast []CastMember `json:"cast"`
	Crew []CrewMember `json:"crew"`
}
type SeasonDetails struct {
	ID           string    `json:"_id"`
	AirDate      string    `json:"air_date"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	IDNum        int       `json:"id"` // note: duplicated as both _id (string) and id (int)
	PosterPath   string    `json:"poster_path"`
	SeasonNumber int       `json:"season_number"`
	VoteAverage  float64   `json:"vote_average"`
	Episodes     []Episode `json:"episodes"`
}

type TMDBWatchProvidersResponse struct {
	ID      int64                           `json:"id"`
	Results map[string]TMDBCountryProviders `json:"results"`
}

type TMDBCountryProviders struct {
	Link     string             `json:"link"`
	Rent     []TMDBProviderInfo `json:"rent,omitempty"`
	Buy      []TMDBProviderInfo `json:"buy,omitempty"`
	Flatrate []TMDBProviderInfo `json:"flatrate,omitempty"`
	// Add any other array fields TMDB might provide (e.g. 'ads', 'free')
}

type TMDBProviderInfo struct {
	LogoPath        string `json:"logo_path"`
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	DisplayPriority int    `json:"display_priority"`
}

type DiscoverMovieParams struct {
	Page int `url:"page,omitempty"`

	// Date range filters
	ReleaseDateGTE string `url:"release_date.gte,omitempty"` // Format: YYYY-MM-DD
	ReleaseDateLTE string `url:"release_date.lte,omitempty"` // Format: YYYY-MM-DD

	// Studio/company filter
	WithCompanies string `url:"with_companies,omitempty"` // TMDB company ID(s), comma/pipe separated

	// Genres
	WithGenres string `url:"with_genres,omitempty"` // TMDB genre ID(s), comma/pipe separated

	// Keywords
	WithKeywords string `url:"with_keywords,omitempty"` // TMDB keyword ID(s), comma/pipe separated

	// Language
	WithOriginalLanguage string `url:"with_original_language,omitempty"` // ISO 639-1 code (e.g. "en", "fr")

	// Runtime
	WithRuntimeGTE int `url:"with_runtime.gte,omitempty"`
	WithRuntimeLTE int `url:"with_runtime.lte,omitempty"`

	// TMDB user score (vote_average)
	VoteAverageGTE float64 `url:"vote_average.gte,omitempty"`
	VoteAverageLTE float64 `url:"vote_average.lte,omitempty"`

	// TMDB user vote count
	VoteCountGTE int `url:"vote_count.gte,omitempty"`
	VoteCountLTE int `url:"vote_count.lte,omitempty"`

	// Streaming services
	WithWatchProviders         string `url:"with_watch_providers,omitempty"`          // Comma/pipe separated TMDB provider IDs
	WithWatchMonetizationTypes string `url:"with_watch_monetization_types,omitempty"` // flatrate, free, ads, rent, buy
	WatchRegion                string `url:"watch_region,omitempty"`                  // e.g. "US", "GB"

	// Sorting (optional, but often useful)
	SortBy string `url:"sort_by,omitempty"`
}

type DiscoverTVParams struct {
	Page int `url:"page,omitempty"`

	// Date range filters for TV shows
	AirDateGTE           string `url:"air_date.gte,omitempty"`            // Format: YYYY-MM-DD
	AirDateLTE           string `url:"air_date.lte,omitempty"`            // Format: YYYY-MM-DD
	FirstAirDateYear     int    `url:"first_air_date_year,omitempty"`     // Year as integer
	FirstAirDateGTE      string `url:"first_air_date.gte,omitempty"`      // Format: YYYY-MM-DD
	FirstAirDateLTE      string `url:"first_air_date.lte,omitempty"`      // Format: YYYY-MM-DD

	// Content filters
	IncludeAdult               bool   `url:"include_adult,omitempty"`
	IncludeNullFirstAirDates   bool   `url:"include_null_first_air_dates,omitempty"`
	Language                   string `url:"language,omitempty"`           // Defaults to en-US
	ScreenedTheatrically       bool   `url:"screened_theatrically,omitempty"`
	Timezone                   string `url:"timezone,omitempty"`

	// Rating filters
	VoteAverageGTE float64 `url:"vote_average.gte,omitempty"`
	VoteAverageLTE float64 `url:"vote_average.lte,omitempty"`
	VoteCountGTE   float64 `url:"vote_count.gte,omitempty"`
	VoteCountLTE   float64 `url:"vote_count.lte,omitempty"`

	// Content categorization
	WithGenres               string `url:"with_genres,omitempty"`                 // Comma (AND) or pipe (OR) separated
	WithKeywords             string `url:"with_keywords,omitempty"`               // Comma (AND) or pipe (OR) separated
	WithCompanies            string `url:"with_companies,omitempty"`              // Comma (AND) or pipe (OR) separated
	WithNetworks             int    `url:"with_networks,omitempty"`               // TV network ID
	WithOriginCountry        string `url:"with_origin_country,omitempty"`
	WithOriginalLanguage     string `url:"with_original_language,omitempty"`      // ISO 639-1 code

	// Runtime filters
	WithRuntimeGTE int `url:"with_runtime.gte,omitempty"`
	WithRuntimeLTE int `url:"with_runtime.lte,omitempty"`

	// Status filters
	WithStatus string `url:"with_status,omitempty"` // 0,1,2,3,4,5 - comma (AND) or pipe (OR) separated

	// Watch providers
	WatchRegion                string `url:"watch_region,omitempty"`                  // Use with watch providers
	WithWatchMonetizationTypes string `url:"with_watch_monetization_types,omitempty"` // flatrate,free,ads,rent,buy
	WithWatchProviders         string `url:"with_watch_providers,omitempty"`          // Use with watch_region

	// Exclusion filters
	WithoutCompanies      string `url:"without_companies,omitempty"`
	WithoutGenres         string `url:"without_genres,omitempty"`
	WithoutKeywords       string `url:"without_keywords,omitempty"`
	WithoutWatchProviders string `url:"without_watch_providers,omitempty"`

	// Type filter
	WithType string `url:"with_type,omitempty"` // 0,1,2,3,4,5,6 - comma (AND) or pipe (OR) separated

	// Sorting
	SortBy string `url:"sort_by,omitempty"` // Defaults to popularity.desc
}

// Watch providers list response
type TMDBWatchProvidersListResponse struct {
	Results []TMDBWatchProvider `json:"results"`
}

type TMDBWatchProvider struct {
	DisplayPriorities map[string]int `json:"display_priorities"`
	DisplayPriority   int            `json:"display_priority"`
	LogoPath          string         `json:"logo_path"`
	ProviderName      string         `json:"provider_name"`
	ProviderID        int            `json:"provider_id"`
}

// Watch provider regions response
type TMDBWatchProviderRegionsResponse struct {
	Results []TMDBWatchProviderRegion `json:"results"`
}

type TMDBWatchProviderRegion struct {
	ISO3166_1   string `json:"iso_3166_1"`
	EnglishName string `json:"english_name"`
	NativeName  string `json:"native_name"`
}

// Company search response
type TMDBCompanySearchResponse struct {
	TMDBPageResults
	Results []TMDBCompany `json:"results"`
}

type TMDBCompany struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

// Release dates response
type TMDBReleaseDatesResponse struct {
	ID      int                     `json:"id"`
	Results []TMDBCountryReleaseDate `json:"results"`
}

type TMDBCountryReleaseDate struct {
	ISO3166_1    string             `json:"iso_3166_1"`
	ReleaseDates []TMDBReleaseDate `json:"release_dates"`
}

type TMDBReleaseDate struct {
	Certification string `json:"certification"`
	ReleaseDate   string `json:"release_date"`
	Type          int    `json:"type"`
	Note          string `json:"note"`
}

// Collection structures
type TMDBCollectionResponse struct {
	ID           int64             `json:"id"`
	Name         string            `json:"name"`
	Overview     string            `json:"overview"`
	PosterPath   string            `json:"poster_path"`
	BackdropPath string            `json:"backdrop_path"`
	Parts        []TMDBMediaItem   `json:"parts"`
}

type Collection struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// Person structures
type TMDBPersonResponse struct {
	Adult              bool                     `json:"adult"`
	AlsoKnownAs        []string                 `json:"also_known_as"`
	Biography          string                   `json:"biography"`
	Birthday           string                   `json:"birthday"`
	Deathday           string                   `json:"deathday"`
	Gender             int                      `json:"gender"`
	Homepage           string                   `json:"homepage"`
	ID                 int64                    `json:"id"`
	ImdbID             string                   `json:"imdb_id"`
	KnownForDepartment string                   `json:"known_for_department"`
	Name               string                   `json:"name"`
	PlaceOfBirth       string                   `json:"place_of_birth"`
	Popularity         float64                  `json:"popularity"`
	ProfilePath        string                   `json:"profile_path"`
	MovieCredits       TMDBPersonMovieCredits   `json:"movie_credits"`
	TvCredits          TMDBPersonTVCredits      `json:"tv_credits"`
}

type TMDBPersonMovieCredits struct {
	Cast []TMDBPersonMovieCast `json:"cast"`
	Crew []TMDBPersonMovieCrew `json:"crew"`
}

type TMDBPersonTVCredits struct {
	Cast []TMDBPersonTVCast `json:"cast"`
	Crew []TMDBPersonTVCrew `json:"crew"`
}

type TMDBPersonMovieCast struct {
	Adult         bool    `json:"adult"`
	BackdropPath  string  `json:"backdrop_path"`
	GenreIDs      []int   `json:"genre_ids"`
	ID            int64   `json:"id"`
	OriginalLanguage string `json:"original_language"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	Popularity    float64 `json:"popularity"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	Title         string  `json:"title"`
	Video         bool    `json:"video"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int64   `json:"vote_count"`
	Character     string  `json:"character"`
	CreditID      string  `json:"credit_id"`
	Order         int     `json:"order"`
}

type TMDBPersonMovieCrew struct {
	Adult         bool    `json:"adult"`
	BackdropPath  string  `json:"backdrop_path"`
	GenreIDs      []int   `json:"genre_ids"`
	ID            int64   `json:"id"`
	OriginalLanguage string `json:"original_language"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	Popularity    float64 `json:"popularity"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	Title         string  `json:"title"`
	Video         bool    `json:"video"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int64   `json:"vote_count"`
	CreditID      string  `json:"credit_id"`
	Department    string  `json:"department"`
	Job           string  `json:"job"`
}

type TMDBPersonTVCast struct {
	Adult         bool     `json:"adult"`
	BackdropPath  string   `json:"backdrop_path"`
	GenreIDs      []int    `json:"genre_ids"`
	ID            int64    `json:"id"`
	OriginCountry []string `json:"origin_country"`
	OriginalLanguage string `json:"original_language"`
	OriginalName  string   `json:"original_name"`
	Overview      string   `json:"overview"`
	Popularity    float64  `json:"popularity"`
	PosterPath    string   `json:"poster_path"`
	FirstAirDate  string   `json:"first_air_date"`
	Name          string   `json:"name"`
	VoteAverage   float64  `json:"vote_average"`
	VoteCount     int64    `json:"vote_count"`
	Character     string   `json:"character"`
	CreditID      string   `json:"credit_id"`
	EpisodeCount  int      `json:"episode_count"`
	FirstCreditAirDate string `json:"first_credit_air_date"`
}

type TMDBPersonTVCrew struct {
	Adult         bool     `json:"adult"`
	BackdropPath  string   `json:"backdrop_path"`
	GenreIDs      []int    `json:"genre_ids"`
	ID            int64    `json:"id"`
	OriginCountry []string `json:"origin_country"`
	OriginalLanguage string `json:"original_language"`
	OriginalName  string   `json:"original_name"`
	Overview      string   `json:"overview"`
	Popularity    float64  `json:"popularity"`
	PosterPath    string   `json:"poster_path"`
	FirstAirDate  string   `json:"first_air_date"`
	Name          string   `json:"name"`
	VoteAverage   float64  `json:"vote_average"`
	VoteCount     int64    `json:"vote_count"`
	CreditID      string   `json:"credit_id"`
	Department    string   `json:"department"`
	EpisodeCount  int      `json:"episode_count"`
	FirstCreditAirDate string `json:"first_credit_air_date"`
	Job           string   `json:"job"`
}
