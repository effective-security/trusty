package stats

import (
	"context"

	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/internal/db/cadb"
	"github.com/martinisecurity/trusty/internal/db/orgsdb"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/internal/tasks", "stats")

// TaskName is the name of this task
const TaskName = "stats"

var (
	keyForDbUsersCount   = []string{"db", "stats", "users_count"}
	keyForDbOrgsCount    = []string{"db", "stats", "orgs_count"}
	keyForDbCertsCount   = []string{"db", "stats", "certs_count"}
	keyForDbRevokedCount = []string{"db", "stats", "revoked_count"}
)

// Task defines the healthcheck task
type Task struct {
	name     string
	schedule string
	ca       cadb.CaReadonlyDb
	orgs     orgsdb.OrgsReadOnlyDb
}

func (t *Task) run() {
	logger.Infof("api=run, task=%s", TaskName)
	ctx := context.Background()

	c, err := t.orgs.GetUsersCount(ctx)
	if err != nil {
		logger.Errorf(errors.ErrorStack(err))
	} else {
		metrics.IncrCounter(keyForDbUsersCount, float32(c))
		logger.Infof("users_count=%d", c)
	}

	c, err = t.orgs.GetOrgsCount(ctx)
	if err != nil {
		logger.Errorf(errors.ErrorStack(err))
	} else {
		metrics.IncrCounter(keyForDbOrgsCount, float32(c))
		logger.Infof("orgs_count=%d", c)
	}

	c, err = t.ca.GetCertsCount(ctx)
	if err != nil {
		logger.Errorf(errors.ErrorStack(err))
	} else {
		metrics.IncrCounter(keyForDbCertsCount, float32(c))
		logger.Infof("certs_count=%d", c)
	}

	c, err = t.ca.GetRevokedCount(ctx)
	if err != nil {
		logger.Errorf(errors.ErrorStack(err))
	} else {
		metrics.IncrCounter(keyForDbRevokedCount, float32(c))
		logger.Infof("revoked_count=%d", c)
	}
}

func create(
	name string,
	ca cadb.CaReadonlyDb,
	orgs orgsdb.OrgsReadOnlyDb,
	schedule string,
) (*Task, error) {
	task := &Task{
		ca:       ca,
		orgs:     orgs,
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
	return func(ca cadb.CaReadonlyDb, orgs orgsdb.OrgsReadOnlyDb) error {
		task, err := create(name, ca, orgs, schedule)
		if err != nil {
			return errors.Trace(err)
		}

		job, err := tasks.NewTask(task.schedule)
		if err != nil {
			return errors.Annotatef(err, "unable to schedule a job on schedule: %q", task.schedule)
		}

		t := job.Do(task.name, task.run)
		s.Add(t)
		// execute immideately
		go t.Run()
		return nil
	}
}
