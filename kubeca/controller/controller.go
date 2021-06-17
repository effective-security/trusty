package controller

import (
	capi "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	// +kubebuilder:scaffold:imports
	"github.com/ekspand/trusty/authority"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
)

var (
	scheme = runtime.NewScheme()
	log    = ctrl.Log.WithName("controller")
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
		log.Error(err, "unable to start manager")
		return err
	}

	crypto, err := cryptoprov.Load(f.HsmCfgPath, nil)
	if err != nil {
		log.Error(err, "unable to load HSM config", "config", f.HsmCfgPath)
		return err
	}

	caCfg, err := authority.LoadConfig(f.CaCfgPath)
	if err != nil {
		log.Error(err, "unable to load CA config", "config", f.CaCfgPath)
		return err
	}

	ca, err := authority.NewAuthority(caCfg, crypto)
	if err != nil {
		log.Error(err, "unable to create CA")
		return err
	}
	if err := (&CertificateSigningRequestSigningReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName(controllerName),
		Scheme:        mgr.GetScheme(),
		Authority:     ca,
		EventRecorder: mgr.GetEventRecorderFor(controllerName),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create Controller", "controller", controllerName)
		return err

	}
	// +kubebuilder:scaffold:builder

	log.Info("starting controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to start controller", "controller", controllerName)
		return err
	}
	return nil
}
