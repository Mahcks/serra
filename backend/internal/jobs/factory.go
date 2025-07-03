package jobs

import (
	"fmt"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// NewJob creates a job by name
func NewJob(name structures.Job, gctx global.Context) (Job, error) {
	switch name {
	case structures.JobDownloadPoller:
		return NewDownloadPoller(gctx)
	case structures.JobDriveMonitor:
		return NewDriveMonitor(gctx)
	default:
		return nil, fmt.Errorf("unknown job: %s", name)
	}
}

// AllJobNames returns all available job names
func AllJobNames() []structures.Job {
	return []structures.Job{structures.JobDownloadPoller, structures.JobDriveMonitor}
}
