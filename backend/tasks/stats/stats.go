package stats

import (
	"context"

	"github.com/effective-security/metrics"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/xlog"
	"github.com/martinisecurity/trusty/backend/db/cadb"
	"github.com/martinisecurity/trusty/pkg/metricskey"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/backend/tasks", "stats")

// TaskName is the name of this task
const TaskName = "stats"

// Task defines the healthcheck task
type Task struct {
	name     string
	schedule string
	ca       cadb.CaReadonlyDb
}

func (t *Task) run() {
	logger.Infof("api=run, task=%s", TaskName)
	ctx := context.Background()

	c, err := t.ca.GetCertsCount(ctx)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
	} else {
		metrics.SetGauge(metricskey.StatsDbCertsTotal, float32(c))
		logger.Infof("certs_count=%d", c)
	}

	c, err = t.ca.GetRevokedCount(ctx)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
	} else {
		metrics.SetGauge(metricskey.StatsDbRevokedTotal, float32(c))
		logger.Infof("revoked_count=%d", c)
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
