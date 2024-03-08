package tasks

import (
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/trusty/backend/tasks/certsmonitor"
	"github.com/effective-security/trusty/backend/tasks/healthcheck"
	"github.com/effective-security/trusty/backend/tasks/stats"
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
) any

// Factories provides map of Factory
var Factories = map[string]Factory{
	certsmonitor.TaskName: certsmonitor.Factory,
	stats.TaskName:        stats.Factory,
	healthcheck.TaskName:  healthcheck.Factory,
}
