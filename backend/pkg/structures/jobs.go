package structures

type Job string

const (
	JobDownloadPoller        Job = "download_poller"
	JobDriveMonitor          Job = "drive_monitor"
	JobRequestProcessor      Job = "request_processor"
	JobLibrarySyncFull       Job = "library_sync_full"
	JobLibrarySyncIncremental Job = "library_sync_incremental"
)

func (j Job) String() string {
	return string(j)
}
