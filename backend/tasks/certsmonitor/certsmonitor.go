package certsmonitor

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/certutil"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/tasks", "certsmonitor")

// TaskName is the name of this task
const TaskName = "certsmonitor"

const (
	typClient = "client"
	typServer = "server"
	typPeer   = "peer"
	typIssuer = "issuer"
)

// Task defines the healthcheck task
type Task struct {
	name     string
	schedule string
	certsMap map[string]string // map => location:type
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

	logger.ContextKV(t.ctx, xlog.TRACE,
		"task", TaskName,
		"certs_count", len(t.certsMap))

	for location, typ := range t.certsMap {
		chain, err := certutil.LoadChainFromPEM(location)
		if err != nil {
			logger.ContextKV(t.ctx, xlog.ERROR,
				"file", location,
				"err", err.Error())
		} else {
			for idx, cert := range chain {
				if idx > 0 {
					typ = typIssuer
				}
				logger.ContextKV(t.ctx, xlog.INFO,
					"type", typ,
					"cert", location,
					"cn", cert.Subject.CommonName,
					"expires", cert.NotAfter.Format(time.RFC3339),
				)
				if typ == typIssuer {
					PublishCACertExpirationInDays(cert, typ)
				} else {
					PublishShortLivedCertExpirationInDays(cert, typ)
				}
			}
		}
	}
}

func certsMapFromLocations(locations []string) map[string]string {
	certsMap := map[string]string{}
	for _, loc := range locations {
		tokens := strings.Split(loc, ":")
		if len(tokens) == 2 {
			certsMap[tokens[1]] = tokens[0]
		} else {
			certsMap[loc] = typClient
		}
	}
	return certsMap
}

func create(
	name string,
	schedule string,
	args ...string,
) (*Task, error) {
	task := &Task{
		name:     name,
		schedule: schedule,
		ctx:      correlation.WithID(context.Background()),
	}

	task.certsMap = certsMapFromLocations(args)
	for location, typ := range task.certsMap {
		logger.Infof("type=%q, location=%q,", typ, location)
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
	return func() error {
		task, err := create(name, schedule, args...)
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
