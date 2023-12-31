package healthcheck

import (
	"bytes"
	"context"
	"crypto"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/api/client"
	"github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ocsp"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/tasks", "healthcheck")

// TaskName is the name of this task
const TaskName = "health_check"

const userAgent = "trusty-healthcheck"

// Task defines the healthcheck task
type Task struct {
	conf       *config.Configuration
	name       string
	schedule   string
	crypto     *cryptoprov.Crypto
	factory    client.Factory
	caClient   pb.CAServer
	ocspClient *retriable.Client
	ocspCerts  []string
	ctx        context.Context
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

	var err error
	started := time.Now()
	if t.factory != nil {
		err = t.healthCheckIssuers(t.ctx)
		if err != nil {
			logger.ContextKV(t.ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "healthCheckIssuers",
				"elapsed", time.Since(started).String(),
				"err", err.Error())
		}
	}

	if t.crypto != nil {
		err = t.healthHsm(t.ctx)
		if err != nil {
			logger.ContextKV(t.ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "healthHsm",
				"elapsed", time.Since(started).String(),
				"err", err.Error())
		}
	}

	for _, cert := range t.ocspCerts {
		go func(cert string) {
			started := time.Now()
			_, err := t.healthCheckOCSP(t.ctx, cert)
			if err != nil {
				logger.ContextKV(t.ctx, xlog.ERROR,
					"task", TaskName,
					"reason", "ocsp",
					"cert", cert,
					"elapsed", time.Since(started).String(),
					"err", err.Error())
			}
		}(cert)
	}

	ver := version.Current()
	metricskey.HealthVersion.SetGauge(float64(ver.Commit))
	// emit an empty `log_errors` metric to have it available in Grafana
	metricskey.HealthLogErrors.IncrCounter(0, "ping")
}

func (t *Task) healthHsm(ctx context.Context) error {
	if keyProv, ok := t.crypto.Default().(cryptoprov.KeyManager); ok {
		ki, err := keyProv.EnumKeys(keyProv.CurrentSlotID(), "")
		if err != nil {
			metricskey.HealthKmsKeysStatusFailCount.IncrCounter(1)
			metricskey.HealthKmsKeysStatusSuccessCount.IncrCounter(0)
			return err
		}
		count := len(ki)
		metricskey.HealthKmsKeysStatusFailCount.IncrCounter(0)
		metricskey.HealthKmsKeysStatusSuccessCount.IncrCounter(1)

		metricskey.StatsKmsKeysTotal.SetGauge(float64(count))

		logger.ContextKV(ctx, xlog.DEBUG, "keys", count)
	}
	return nil
}

func (t *Task) healthCheckIssuers(ctx context.Context) error {
	if t.caClient == nil {
		cl, _, err := t.factory.CAClient("ca", client.WithAgent(userAgent))
		if err != nil {
			return errors.WithMessagef(err, "unable to create CA client")
		}
		t.caClient = cl
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	li, err := t.caClient.ListIssuers(ctx, &pb.ListIssuersRequest{
		//Limit:  1,
		After:  0,
		Bundle: false,
	})
	if err != nil {
		metricskey.HealthCAStatusFailCount.IncrCounter(1)
		metricskey.HealthCAStatusSuccessCount.IncrCounter(0)
		return err
	}
	metricskey.HealthCAStatusFailCount.IncrCounter(0)
	metricskey.HealthCAStatusSuccessCount.IncrCounter(1)

	count := len(li.Issuers)
	metricskey.StatsCAIssuersTotal.SetGauge(float64(count))
	logger.ContextKV(ctx, xlog.DEBUG, "issuers", count)

	return nil
}

const httpTimeout = 3 * time.Second

func (t *Task) healthCheckOCSP(ctx context.Context, cert string) (*int, error) {
	chain, err := certutil.LoadChainFromPEM(cert)
	if err != nil {
		return nil, err
	}

	if len(chain) < 2 {
		return nil, errors.Errorf("invalid chain of length %d: %s",
			len(chain), cert)
	}

	crt := chain[0]
	issuer := chain[1]

	if len(crt.OCSPServer) == 0 {
		logger.ContextKV(t.ctx, xlog.WARNING,
			"reason", "no_ocsp_url",
			"cert", cert)
		return nil, nil
	}

	req, err := certutil.CreateOCSPRequest(crt, issuer, crypto.SHA256)
	if err != nil {
		return nil, err
	}

	ur, err := url.Parse(crt.OCSPServer[0])
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid OCSP URL: %s", crt.OCSPServer[0])
	}

	defer metricskey.HealthOCSPCheckPerf.MeasureSince(time.Now(), ur.Host)

	w := bytes.NewBuffer([]byte{})
	host := fmt.Sprintf("%s://%s", ur.Scheme, ur.Host)
	ur.Hostname()
	_, _, err = t.ocspClient.Request(
		ctx,
		http.MethodPost,
		host,
		ur.Path,
		req,
		w)
	if err != nil {
		metricskey.HealthOCSPStatusFailCount.IncrCounter(1, ur.Host)
		metricskey.HealthOCSPStatusSuccessCount.IncrCounter(0, ur.Host)

		logger.ContextKV(ctx, xlog.ERROR,
			"host", ur.Host,
			"err", err.Error())
		return nil, errors.WithStack(err)
	}

	metricskey.HealthOCSPStatusFailCount.IncrCounter(0, ur.Host)
	metricskey.HealthOCSPStatusSuccessCount.IncrCounter(1, ur.Host)

	res, err := ocsp.ParseResponseForCert(w.Bytes(), crt, issuer)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logger.ContextKV(ctx, xlog.DEBUG,
		"host", ur.Host,
		"ocsp_status", res.Status,
		"with_cert", res.Certificate != nil,
	)

	return &res.Status, nil
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
	ocspPtr := flagSet.String("ocsp", "", "check status of OCSP")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to parse arguments: %v", args)
	}

	task := &Task{
		conf:     conf,
		name:     name,
		schedule: schedule,
		ctx:      correlation.WithID(context.Background()),
	}

	if caPtr != nil && *caPtr {
		task.factory = client.NewFactory(&conf.Client)
	}

	if hsmkeysPtr != nil && *hsmkeysPtr && conf != nil {
		task.crypto, err = cryptoprov.Load(conf.CryptoProv.Default, conf.CryptoProv.Providers)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to initialize crypto providers")
		}
	}

	if ocspPtr != nil && *ocspPtr != "" {
		task.ocspCerts = append(task.ocspCerts, *ocspPtr)
		task.ocspClient, err = retriable.New(
			retriable.ClientConfig{},
			retriable.WithName("ocsphealth"),
			retriable.WithTLS(nil),
			retriable.WithTimeout(httpTimeout),
			retriable.WithUserAgent(userAgent),
		)
		if err != nil {
			return nil, err
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
