package tasks

import (
	"github.com/go-phorce/dolly/tasks"
)

// Status indicates whether the task failed or succeeded
type Status int

const (
	// FAILURE indicates that the task failed
	FAILURE Status = 0

	// SUCCESS indicates that the task succeeded
	SUCCESS Status = 1
)

// Factory creates tasks
type Factory func(
	scheduler tasks.Scheduler,
	name string,
	schedule string,
	args ...string,
) interface{}
