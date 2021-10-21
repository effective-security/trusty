package certsmonitor

import (
	"strings"
	"time"

	"github.com/go-phorce/dolly/fileutil"
	metricsutil "github.com/go-phorce/dolly/metrics/util"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/backend/tasks", "certsmonitor")

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
	conf     *config.Configuration
	name     string
	schedule string
	certsMap map[string]string // map => location:type
}

func (t *Task) run() {
	logger.Infof("api=run, task=%s, count=%d", TaskName, len(t.certsMap))

	for location, typ := range t.certsMap {
		cert, err := certutil.LoadFromPEM(location)
		if err != nil {
			logger.Errorf("api=certsmonitor, file=%q, err=[%+v]", location, err.Error())
		} else {
			logger.Infof("cert=%q, subject=%q, expires=%q", location, cert.Subject.CommonName, cert.NotAfter.Format(time.RFC3339))
			if typ == typIssuer {
				metricsutil.PublishCertExpirationInDays(cert, typ)
			} else {
				metricsutil.PublishShortLivedCertExpirationInDays(cert, typ)
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
	conf *config.Configuration,
	schedule string,
	args ...string,
) (*Task, error) {
	task := &Task{
		conf:     conf,
		name:     name,
		schedule: schedule,
	}

	if conf != nil {
		task.certsMap = certsMapFromLocations(conf.CertsMonitor.Locations)
		if !conf.TrustyClient.ClientTLS.Empty() &&
			fileutil.FileExists(conf.TrustyClient.ClientTLS.CertFile) == nil {
			task.certsMap[conf.TrustyClient.ClientTLS.CertFile] = typClient
		}

		for _, srv := range conf.HTTPServers {
			if !srv.Disabled && !srv.ServerTLS.Empty() &&
				fileutil.FileExists(srv.ServerTLS.CertFile) == nil {
				task.certsMap[srv.ServerTLS.CertFile] = typServer
			}
		}

		for location, typ := range task.certsMap {
			logger.Infof("api=create, type=%q, location=%q,", typ, location)
		}
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
	return func(cfg *config.Configuration) error {

		task, err := create(name, cfg, schedule, args...)
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
