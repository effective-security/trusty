package stats

import (
	"context"
	"runtime/debug"

	"github.com/effective-security/metrics"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/tasks", "stats")

// TaskName is the name of this task
const TaskName = "stats"

// Task defines the healthcheck task
type Task struct {
	name     string
	schedule string
	ca       cadb.CaReadonlyDb
}

func (t *Task) run() {
	ctx := correlation.WithID(context.Background())
	defer func() {
		if r := recover(); r != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "recover",
				"err", r,
				"stack", debug.Stack())
		}
	}()

	logger.ContextKV(ctx, xlog.TRACE,
		"task", TaskName,
	)

	c, err := t.ca.GetCertsCount(ctx)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR, "err", err)
	} else {
		metrics.SetGauge(metricskey.StatsDbCertsTotal, float32(c))
		logger.ContextKV(ctx, xlog.INFO, "certs_count", c)
	}

	c, err = t.ca.GetRevokedCount(ctx)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR, "err", err)
	} else {
		metrics.SetGauge(metricskey.StatsDbRevokedTotal, float32(c))
		logger.ContextKV(ctx, xlog.INFO, "revoked_count", c)
	}
}

func create(
	name string,
	ca cadb.CaReadonlyDb,
	schedule string,
) (*Task, error) {
	task := &Task{
		ca:       ca,
		name:     name,
		schedule: schedule,
	}

	return task, nil
}

// Factory returns an task factory
func Factory(
	s tasks.Scheduler,
	name string,
	schedule string,
	args ...string,
) interface{} {
	return func(ca cadb.CaReadonlyDb) error {
		task, err := create(name, ca, schedule)
		if err != nil {
			return errors.WithStack(err)
		}

		job, err := tasks.NewTask(task.schedule)
		if err != nil {
			return errors.WithMessagef(err, "unable to schedule a job on schedule: %q", task.schedule)
		}

		t := job.Do(task.name, task.run)
		s.Add(t)
		// execute immideately
		go t.Run()
		return nil
	}
}
