package healthcheck

import (
	"bytes"
	"context"
	"crypto"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/effective-security/metrics"
	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/client"
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

// Task defines the healthcheck task
type Task struct {
	conf       *config.Configuration
	name       string
	schedule   string
	crypto     *cryptoprov.Crypto
	factory    client.Factory
	caClient   client.CAClient
	ocspClient *retriable.Client
	ocspCerts  []string
}

func (t *Task) run() {
	ctx := correlation.WithID(context.Background())
	defer func() {
		if r := recover(); r != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "recover",
				"err", r)
		}
	}()

	var err error
	started := time.Now()
	if t.factory != nil {
		err = t.healthCheckIssuers(ctx)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "healthCheckIssuers",
				"elapsed", time.Since(started).String(),
				"err", err.Error())
		}
	}

	if t.crypto != nil {
		err = t.healthHsm(ctx)
		if err != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"task", TaskName,
				"reason", "healthHsm",
				"elapsed", time.Since(started).String(),
				"err", err.Error())
		}
	}

	for _, cert := range t.ocspCerts {
		go func(cert string) {
			started := time.Now()
			_, err := t.healthCheckOCSP(ctx, cert)
			if err != nil {
				logger.ContextKV(ctx, xlog.ERROR,
					"task", TaskName,
					"reason", "ocsp",
					"cert", cert,
					"elapsed", time.Since(started).String(),
					"err", err.Error())
			}
		}(cert)
	}

	metrics.SetGauge([]string{"version"}, version.Current().Float())
}

func (t *Task) healthHsm(ctx context.Context) error {
	if keyProv, ok := t.crypto.Default().(cryptoprov.KeyManager); ok {
		ki, err := keyProv.EnumKeys(keyProv.CurrentSlotID(), "")
		if err != nil {
			metrics.IncrCounter(metricskey.HealthKmsKeysStatusFailedCount, 1)
			return err
		}
		count := len(ki)
		metrics.IncrCounter(metricskey.HealthKmsKeysStatusSuccessfulCount, 1)
		metrics.SetGauge(metricskey.StatsKmsKeysTotal, float32(count))
		logger.ContextKV(ctx, xlog.DEBUG, "keys", count)
	}
	return nil
}

func (t *Task) healthCheckIssuers(ctx context.Context) error {
	if t.caClient == nil {
		client, err := t.factory.NewClient("ca")
		if err != nil {
			return errors.WithMessagef(err, "unable to create client")
		}
		t.caClient = client.CAClient()
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	li, err := t.caClient.ListIssuers(ctx, &pb.ListIssuersRequest{
		//Limit:  1,
		After:  0,
		Bundle: false,
	})
	if err != nil {
		metrics.IncrCounter(metricskey.HealthCAStatusFailedCount, 1)
		return errors.WithStack(err)
	}
	metrics.IncrCounter(metricskey.HealthCAStatusSuccessfulCount, 1)
	logger.ContextKV(ctx, xlog.DEBUG, "issuers", len(li.Issuers))
	return nil
}

const httpTimeout = 3 * time.Second

func (t *Task) healthCheckOCSP(ctx context.Context, cert string) (*int, error) {
	chain, err := certutil.LoadChainFromPEM(cert)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(chain) < 2 {
		return nil, errors.Errorf("invalid chain of length %d: %s",
			len(chain), cert)
	}

	crt := chain[0]
	issuer := chain[1]

	if len(crt.OCSPServer) == 0 {
		return nil, errors.Errorf("certificate does not have OCSP URL: %s",
			cert)
	}

	req, err := certutil.CreateOCSPRequest(crt, issuer, crypto.SHA256)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ur, err := url.Parse(crt.OCSPServer[0])
	if err != nil {
		return nil, errors.WithMessagef(err, "invalid OCSP URL: %s", crt.OCSPServer[0])
	}
	tag := metrics.Tag{Name: "host", Value: ur.Host}

	defer metrics.MeasureSince(metricskey.HealthOCSPCheckPerf, time.Now())

	metrics.IncrCounter(metricskey.HealthOCSPStatusTotalCount, 1, tag)

	w := bytes.NewBuffer([]byte{})
	host := fmt.Sprintf("%s://%s", ur.Scheme, ur.Host)
	ur.Hostname()
	_, _, err = t.ocspClient.Request(
		ctx,
		http.MethodPost,
		[]string{host},
		ur.Path,
		req,
		w)
	if err != nil {
		metrics.IncrCounter(metricskey.HealthOCSPStatusFailedCount, 1, tag)
		logger.ContextKV(ctx, xlog.ERROR,
			"host", ur.Host,
			"err", err.Error())
		return nil, errors.WithStack(err)
	}

	metrics.IncrCounter(metricskey.HealthOCSPStatusSuccessfulCount, 1, tag)

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

	if ocspPtr != nil && *ocspPtr != "" {
		task.ocspCerts = append(task.ocspCerts, *ocspPtr)
		task.ocspClient = retriable.New(
			retriable.WithName("ocsphealth"),
			retriable.WithTLS(nil),
			retriable.WithTimeout(httpTimeout),
		)
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
