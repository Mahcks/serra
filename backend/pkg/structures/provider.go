package structures

type Provider string

const (
	ProviderEmby     Provider = "emby"
	ProviderJellyfin Provider = "jellyfin"
)

type ArrProvider string

const (
	ProviderRadarr ArrProvider = "radarr"
	ProviderSonarr ArrProvider = "sonarr"
)

func (ap ArrProvider) IsValid() bool {
	switch ap {
	case ProviderRadarr, ProviderSonarr:
		return true
	default:
		return false
	}
}

func (ap ArrProvider) String() string {
	return string(ap)
}
