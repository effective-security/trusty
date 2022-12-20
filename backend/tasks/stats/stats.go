package stats

import (
	"context"
	"runtime/debug"

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
	ctx      context.Context
}

func (t *Task) run() {
	defer func() {
		if r := recover(); r != nil {
			logger.ContextKV(t.ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "recover",
				"err", r,
				"stack", debug.Stack())
		}
	}()

	logger.ContextKV(t.ctx, xlog.TRACE, "task", TaskName)

	for _, table := range tables {
		count, err := t.cadb.GetTableRowsCount(t.ctx, table)
		if err != nil {
			logger.ContextKV(t.ctx, xlog.ERROR, "table", table, "err", err)
		} else {
			metricskey.StatsDbTableRowsTotal.SetGauge(float64(count), table)

			logger.ContextKV(t.ctx, xlog.TRACE, "table", table, "rows", count)
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
		ctx:      correlation.WithID(context.Background()),
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
