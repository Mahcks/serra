package jobs

import (
	"context"

	"github.com/mahcks/serra/pkg/structures"
)

// Job is the interface all jobs must implement.
type Job interface {
	Start()
	Stop(ctx context.Context) error
	Name() structures.Job
	Trigger() error
}
