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

var tables = []string{
	cadb.TableNameForCertificates,
	cadb.TableNameForRevoked,
	cadb.TableNameForCrls,
	cadb.TableNameForIssuers,
	cadb.TableNameForRoots,
	cadb.TableNameForCertProfiles,
	cadb.TableNameForNonces,
}

// Task defines the healthcheck task
type Task struct {
	name     string
	schedule string
	cadb     cadb.CaReadonlyDb
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

	for _, table := range tables {
		count, err := t.cadb.GetTableRowsCount(ctx, table)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR, "table", table, "err", err)
		} else {
			metrics.SetGauge(metricskey.StatsDbTableTotalPrefix, float32(count),
				metrics.Tag{Name: "table", Value: table})

			logger.ContextKV(ctx, xlog.TRACE, "table", table, "rows", count)
		}
	}
}

func create(
	name string,
	ca cadb.CaReadonlyDb,
	schedule string,
) (*Task, error) {
	task := &Task{
		cadb:     ca,
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
