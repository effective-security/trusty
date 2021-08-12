package main

import (
	"flag"
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	"github.com/ekspand/trusty/kubeca/controller"
	"github.com/ekspand/trusty/pkg/awskmscrypto"
	"github.com/ekspand/trusty/pkg/gcpkmscrypto"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
)

func main() {
	xlog.GetFormatter().WithCaller(true)

	cryptoprov.Register("SoftHSM", cryptoprov.Crypto11Loader)
	cryptoprov.Register("PKCS11", cryptoprov.Crypto11Loader)
	cryptoprov.Register("AWSKMS", awskmscrypto.KmsLoader)
	cryptoprov.Register("GCPKMS", gcpkmscrypto.KmsLoader)
	cryptoprov.Register("GCPKMS-roots", gcpkmscrypto.KmsLoader)

	f := controller.CertificateSigningRequestControllerFlags{}
	var debugLogging bool
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

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(debugLogging)))

	err := controller.StartCertificateSigningRequestController(&f)
	if err != nil {
		fmt.Printf("failed to start: %s\n", err.Error())
		os.Exit(1)
	}
}
