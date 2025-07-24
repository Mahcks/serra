package structures

type Setting string

type RequestSystem string

const (
	// RequestSystemBuiltIn uses Serra's built-in request system
	RequestSystemBuiltIn RequestSystem = "built_in"
	// RequestSystemExternal uses an external request system (like Jellyseerr) in an iframe
	RequestSystemExternal RequestSystem = "external"
)

type DownloadVisibility string

const (
	// DownloadVisibilityAll allows all users to see all downloads
	DownloadVisibilityAll DownloadVisibility = "all"
	// DownloadVisibilityOwn allows users to see only their own downloads
	DownloadVisibilityOwn DownloadVisibility = "own"
)

const (
	// SettingSetupComplete indicates that the initial setup has been completed.
	SettingSetupComplete Setting = "setup_complete"
	// SettingMediaServerType indicates the type of media server being used. Either "emby" or "jellyfin".
	SettingMediaServerType Setting = "media_server_type"
	// SettingMediaServerURL indicates the URL of the media server.
	SettingMediaServerURL Setting = "media_server_url"
	// SettingMediaServerAPIKey indicates the API key for the media server.
	SettingMediaServerAPIKey Setting = "media_server_api_key"
	// SettingRequestSystem indicates whether to use built-in request system or external system (like Jellyseerr)
	SettingRequestSystem Setting = "request_system"
	// SettingRequestSystemURL indicates the URL of the external request system (e.g., Jellyseerr)
	SettingRequestSystemURL Setting = "request_system_url"
	// SettingJellystatEnabled indicates whether Jellystat integration is enabled.
	SettingJellystatEnabled Setting = "jellystat_enabled"
	// SettingJellystatHost indicates the host/IP of the Jellystat service.
	SettingJellystatHost Setting = "jellystat_host"
	// SettingJellystatPort indicates the port of the Jellystat service.
	SettingJellystatPort Setting = "jellystat_port"
	// SettingJellystatUseSSL indicates whether to use HTTPS when connecting to Jellystat.
	SettingJellystatUseSSL Setting = "jellystat_use_ssl"
	// SettingJellystatURL indicates the URL of the Jellystat service.
	SettingJellystatURL Setting = "jellystat_url"
	// SettingJellystatAPIKey indicates the API key for the Jellystat service.
	SettingJellystatAPIKey Setting = "jellystat_api_key"
	// SettingDownloadVisibility controls whether users can see all downloads or only their own
	SettingDownloadVisibility Setting = "download_visibility"
	// SettingTMDBAPIKey indicates the API key for The Movie Database (TMDB) service
	SettingTMDBAPIKey Setting = "tmdb_api_key"
	// SettingEnableMediaServerAuth indicates whether users can authenticate using Emby/Jellyfin credentials
	SettingEnableMediaServerAuth Setting = "enable_media_server_auth"
	// SettingEnableLocalAuth indicates whether users can authenticate using local Serra accounts
	SettingEnableLocalAuth Setting = "enable_local_auth"
	// SettingEnableNewMediaServerAuth indicates whether new Emby/Jellyfin users can sign in without being imported first
	SettingEnableNewMediaServerAuth Setting = "enable_new_media_server_auth"
	// SettingGlobalMovieRequestLimit indicates the maximum number of movie requests per user (0 = unlimited)
	SettingGlobalMovieRequestLimit Setting = "global_movie_request_limit"
	// SettingGlobalSeriesRequestLimit indicates the maximum number of series requests per user (0 = unlimited)
	SettingGlobalSeriesRequestLimit Setting = "global_series_request_limit"
	// Default permission settings (individual booleans for each permission)
	// Owner permission
	SettingDefaultOwner Setting = "default_owner"
	// Admin permissions
	SettingDefaultAdminUsers    Setting = "default_admin_users"
	SettingDefaultAdminServices Setting = "default_admin_services"
	SettingDefaultAdminSystem   Setting = "default_admin_system"
	// Request permissions
	SettingDefaultRequestMovies              Setting = "default_request_movies"
	SettingDefaultRequestSeries              Setting = "default_request_series"
	SettingDefaultRequest4KMovies            Setting = "default_request_4k_movies"
	SettingDefaultRequest4KSeries            Setting = "default_request_4k_series"
	SettingDefaultRequestAutoApproveMovies   Setting = "default_request_auto_approve_movies"
	SettingDefaultRequestAutoApproveSeries   Setting = "default_request_auto_approve_series"
	SettingDefaultRequestAutoApprove4KMovies Setting = "default_request_auto_approve_4k_movies"
	SettingDefaultRequestAutoApprove4KSeries Setting = "default_request_auto_approve_4k_series"
	// Request management permissions
	SettingDefaultRequestsView    Setting = "default_requests_view"
	SettingDefaultRequestsApprove Setting = "default_requests_approve"
	SettingDefaultRequestsManage  Setting = "default_requests_manage"
)

func (s Setting) String() string {
	return string(s)
}

func (rs RequestSystem) String() string {
	return string(rs)
}

func (dv DownloadVisibility) String() string {
	return string(dv)
}
