package controller

import (
	capi "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	// +kubebuilder:scaffold:imports
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/authority"
)

var (
	scheme = runtime.NewScheme()
	logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/kubeca", "controller")
)

const controllerName = "CSRSigningReconciler"

func init() {
	_ = capi.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// CertificateSigningRequestControllerFlags provides controller flags
type CertificateSigningRequestControllerFlags struct {
	MetricsAddr          string
	Port                 int
	EnableLeaderElection bool
	LeaderElectionID     string
	CaCfgPath            string
	HsmCfgPath           string
}

// StartCertificateSigningRequestController starts controller loop
func StartCertificateSigningRequestController(f *CertificateSigningRequestControllerFlags) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: f.MetricsAddr,
		Port:               f.Port,
		LeaderElection:     f.EnableLeaderElection,
		LeaderElectionID:   f.LeaderElectionID,
	})
	if err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to start manager",
			"err", err)
		return err
	}

	crypto, err := cryptoprov.Load(f.HsmCfgPath, nil)
	if err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to load HSM config",
			"config", f.HsmCfgPath,
			"err", err)
		return err
	}

	caCfg, err := authority.LoadConfig(f.CaCfgPath)
	if err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to load CA config",
			"config", f.CaCfgPath,
			"err", err)
		return err
	}

	ca, err := authority.NewAuthority(caCfg, crypto)
	if err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to create CA",
			"err", err)
		return err
	}
	if err := (&CertificateSigningRequestSigningReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName(controllerName),
		Scheme:        mgr.GetScheme(),
		Authority:     ca,
		EventRecorder: mgr.GetEventRecorderFor(controllerName),
	}).SetupWithManager(mgr); err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to create Controller",
			"controller", controllerName,
			"err", err)
		return err

	}
	// +kubebuilder:scaffold:builder

	logger.Info("starting controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.KV(xlog.ERROR,
			"reason", "unable to start controller",
			"controller", controllerName,
			"err", err)
		return err
	}
	return nil
}
