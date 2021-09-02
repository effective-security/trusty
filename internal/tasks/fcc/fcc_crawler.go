package fcc

import (
	"context"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/pkg/fcc"
	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal/tasks", "fcc")

// TaskName is the name of this task
const (
	TaskName                  = "fcc"
	fccDefaultTimoutInSeconds = 120
)

var (
	keyForFccFilersCount = []string{"fcc", "stats", "filers_count"}
)

// Task defines the fcc crawler task
type Task struct {
	name     string
	schedule string

	fccBaseURL string
	orgs       orgsdb.OrgsDb
}

func (t *Task) run() {
	logger.Infof("api=run, task=%s", TaskName)
	ctx := context.Background()
	fccClient := fcc.NewAPIClient(t.fccBaseURL, time.Duration(fccDefaultTimoutInSeconds)*time.Second)

	fQueryResults, err := fccClient.GetFiler499Results("")
	if err != nil {
		logger.Errorf(errors.ErrorStack(err))
		return
	}

	filersCount := 0
	for _, filer := range fQueryResults.Filers {
		id, err := db.ID(filer.Form499ID)
		if err != nil {
			logger.Errorf(errors.ErrorStack(err))
			continue
		}
		filerAsArray := martini.ToFilersDto(&fcc.Filer499Results{
			XMLName: fQueryResults.XMLName,
			Filers: []fcc.Filer{
				filer,
			},
		})
		res := &v1.FccFrnResponse{
			Filers: filerAsArray,
		}

		js, _ := marshal.EncodeBytes(marshal.DontPrettyPrint, res)
		_, err = t.orgs.UpdateFRNResponse(ctx, id, string(js))
		if err != nil {
			logger.Errorf("filerID=%d, err=%s", id, errors.Details(err))
			continue
		}

		filersCount++
	}

	metrics.IncrCounter(keyForFccFilersCount, float32(filersCount))
	logger.Infof("filers_count=%d", filersCount)
}

func create(
	name string,
	schedule string,
	fccBaseURL string,
	orgs orgsdb.OrgsDb,
) (*Task, error) {
	task := &Task{
		name:       name,
		orgs:       orgs,
		fccBaseURL: fccBaseURL,
		schedule:   schedule,
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
	return func(orgs orgsdb.OrgsDb) error {
		fccBaseURL := ""
		if len(args) > 0 {
			fccBaseURL = args[0]
		}

		task, err := create(name, schedule, fccBaseURL, orgs)
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
