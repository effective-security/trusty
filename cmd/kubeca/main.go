package main

import (
	"flag"
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	"github.com/effective-security/xlog"
	"github.com/effective-security/xlog/stackdriver"
	"github.com/effective-security/xpki/crypto11"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/effective-security/xpki/cryptoprov/awskmscrypto"
	"github.com/effective-security/xpki/cryptoprov/gcpkmscrypto"
	"github.com/martinisecurity/trusty/kubeca/controller"
)

func main() {
	cryptoprov.Register("SoftHSM", crypto11.LoadProvider)
	cryptoprov.Register("PKCS11", crypto11.LoadProvider)
	cryptoprov.Register(awskmscrypto.ProviderName, awskmscrypto.KmsLoader)
	cryptoprov.Register(awskmscrypto.ProviderName+"-1", awskmscrypto.KmsLoader)
	cryptoprov.Register(awskmscrypto.ProviderName+"-2", awskmscrypto.KmsLoader)

	cryptoprov.Register(gcpkmscrypto.ProviderName, gcpkmscrypto.KmsLoader)
	cryptoprov.Register(gcpkmscrypto.ProviderName+"-1", gcpkmscrypto.KmsLoader)
	cryptoprov.Register(gcpkmscrypto.ProviderName+"-2", gcpkmscrypto.KmsLoader)

	f := controller.CertificateSigningRequestControllerFlags{}
	var debugLogging bool
	var withStackdriver bool
	flag.BoolVar(&debugLogging, "debug", false, "Enable debug logging.")
	flag.StringVar(&f.MetricsAddr, "metrics-addr", ":9090", "The address the metric endpoint binds to.")
	flag.IntVar(&f.Port, "port", 9443, "The port for the controller.")
	flag.BoolVar(&f.EnableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&f.LeaderElectionID, "leader-election-id", "kube-ca-leader-election",
		"The name of the configmap used to coordinate leader election between controller-managers.")

	flag.StringVar(&f.CaCfgPath, "ca-cfg", "/trusty/etc/ca-config.yaml", "Location of CA configuration file.")
	flag.StringVar(&f.HsmCfgPath, "hsm-cfg", "/trusty/etc/aws-kms-us-west-2.json", "Location of HSM configuration file.")
	flag.BoolVar(&withStackdriver, "stackdriver", false, "Enable stackdriver logs formatting.")

	flag.Parse()

	if withStackdriver {
		formatter := stackdriver.NewFormatter(os.Stderr, "kubeca")
		xlog.SetFormatter(formatter)
	}
	xlog.GetFormatter().WithCaller(true)

	ctrl.SetLogger(zap.New(zap.UseDevMode(debugLogging)))

	err := controller.StartCertificateSigningRequestController(&f)
	if err != nil {
		fmt.Printf("failed to start: %s\n", err.Error())
		os.Exit(1)
	}
}
