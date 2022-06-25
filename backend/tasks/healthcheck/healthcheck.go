package healthcheck

import (
	"context"
	"flag"
	"time"

	"github.com/effective-security/metrics"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/metricskey"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/backend/tasks", "healthcheck")

// TaskName is the name of this task
const TaskName = "health_check"

// Task defines the healthcheck task
type Task struct {
	conf     *config.Configuration
	name     string
	schedule string
	crypto   *cryptoprov.Crypto
	factory  client.Factory
	client   *client.Client
}

func (t *Task) run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("task=%s, reason=recover, err=[%+v]", TaskName, r)
		}
	}()

	logger.Infof("task=%s", TaskName)

	ctx := correlation.WithID(context.Background())

	var err error
	started := time.Now()
	if t.factory != nil {
		if t.client == nil {
			t.client, err = t.factory.NewClient("ca")
			if err != nil {
				logger.KV(xlog.ERROR,
					"task", TaskName,
					"reason", "ca",
					"err", err.Error())
			}
		}
		err = t.healthCheckIssuers(ctx)
		if err != nil {
			logger.KV(xlog.ERROR,
				"task", TaskName,
				"reason", "healthCheckIssuers",
				"elapsed", time.Since(started).String(),
				"ctx", correlation.ID(ctx),
				"err", err.Error())
		}
	}

	if t.crypto != nil {
		err = t.healthHsm()
		if err != nil {
			logger.KV(xlog.ERROR,
				"task", TaskName,
				"reason", "healthHsm",
				"elapsed", time.Since(started).String(),
				"ctx", correlation.ID(ctx),
				"err", err.Error())
		}
	}
}

func (t *Task) healthHsm() error {
	if keyProv, ok := t.crypto.Default().(cryptoprov.KeyManager); ok {
		ki, err := keyProv.EnumKeys(keyProv.CurrentSlotID(), "")
		if err != nil {
			metrics.IncrCounter(metricskey.HealthKmsKeysStatusFailedCount, 1)
			return err
		}
		count := len(ki)
		metrics.IncrCounter(metricskey.HealthKmsKeysStatusSuccessfulCount, 1)
		metrics.SetGauge(metricskey.StatsKmsKeysTotal, float32(count))
		logger.Infof("keys=%d", count)
	}
	return nil
}

func (t *Task) healthCheckIssuers(ctx context.Context) error {
	if t.client == nil {
		return errors.Errorf("CA client not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	li, err := t.client.CAClient().ListIssuers(ctx, &pb.ListIssuersRequest{
		//Limit:  1,
		After:  0,
		Bundle: false,
	})
	if err != nil {
		metrics.IncrCounter(metricskey.HealthCAStatusFailedCount, 1)
		return errors.WithStack(err)
	}
	metrics.IncrCounter(metricskey.HealthCAStatusSuccessfulCount, 1)
	logger.Infof("issuers=%d", len(li.Issuers))
	return nil
}

func create(
	name string,
	conf *config.Configuration,
	schedule string,
	args []string,
) (*Task, error) {
	flagSet := flag.NewFlagSet("flags", flag.ContinueOnError)
	caPtr := flagSet.Bool("ca", false, "check status of CA")
	hsmkeysPtr := flagSet.Bool("hsmkeys", false, "list keys in HSM or KMS")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to parse arguments: %v", args)
	}

	task := &Task{
		conf:     conf,
		name:     name,
		schedule: schedule,
	}

	if caPtr != nil && *caPtr {
		task.factory = client.NewFactory(&conf.TrustyClient)
	}

	if hsmkeysPtr != nil && *hsmkeysPtr && conf != nil {
		task.crypto, err = cryptoprov.Load(conf.CryptoProv.Default, conf.CryptoProv.Providers)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to initialize crypto providers")
		}
	}
	return task, nil
}

// Factory returns a task factory
func Factory(
	s tasks.Scheduler,
	name string,
	schedule string,
	args ...string,
) interface{} {
	return func(cfg *config.Configuration) error {
		task, err := create(name, cfg, schedule, args)
		if err != nil {
			return errors.WithStack(err)
		}

		job, err := tasks.NewTask(task.schedule)
		if err != nil {
			return errors.WithMessagef(err, "unable to schedule a job on schedule: %q", task.schedule)
		}

		t := job.Do(task.name, task.run)
		s.Add(t)
		// Do not execute immideately
		// go t.Run()
		return nil
	}
}
