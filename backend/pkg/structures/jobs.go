package structures

type Job string

const (
	JobDownloadPoller Job = "download_poller"
	JobDriveMonitor   Job = "drive_monitor"
)

func (j Job) String() string {
	return string(j)
}
