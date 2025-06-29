package structures

type Setting string

type RequestSystem string

const (
	// RequestSystemBuiltIn uses Serra's built-in request system
	RequestSystemBuiltIn RequestSystem = "built_in"
	// RequestSystemExternal uses an external request system (like Jellyseerr) in an iframe
	RequestSystemExternal RequestSystem = "external"
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
	// SettingJellystatURL indicates the URL of the Jellystat service.
	SettingJellystatURL Setting = "jellystat_url"
	// SettingJellystatAPIKey indicates the API key for the Jellystat service.
	SettingJellystatAPIKey Setting = "jellystat_api_key"
)

func (s Setting) String() string {
	return string(s)
}

func (rs RequestSystem) String() string {
	return string(rs)
}
