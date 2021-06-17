package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/ekspand/trusty/kubeca/certinit"
)

func main() {
	var kubeConfig string
	r := &certinit.Request{}
	flag.StringVar(&kubeConfig, "kubeconfig", "", "(optional) path to kubeconfig file")
	flag.StringVar(&r.Namespace, "namespace", "", "namespace as defined by pod.metadata.namespace")
	flag.StringVar(&r.PodName, "pod-name", "", "name as defined by pod.metadata.name")
	flag.StringVar(&r.CertDir, "cert-dir", "/etc/tls", "directory where the TLS certs should be written")
	flag.StringVar(&r.ClusterDomain, "cluster-domain", "cluster.local", "kubernetes cluster domain")
	flag.StringVar(&r.Labels, "labels", "", "labels to include in CertificateSigningRequest object; comma separated list of key=value")
	flag.BoolVar(&r.QueryK8s, "query-k8s", false, "query kubernetes for names appropriate to this Pod")
	flag.StringVar(&r.SAN, "san", "", "additional SAN; comma separated")
	flag.StringVar(&r.ServiceNames, "service-names", "", "additional service names that resolve to this Pod; comma separated")
	flag.BoolVar(&r.IncludeUnqualified, "include-unqualified", false, "include unqualified .svc domains in names from --query-k8s")
	flag.StringVar(&r.SignerName, "signer", "", "signer name")
	flag.Parse()

	// Create a Kubernetes client.
	client, err := certinit.NewClient(kubeConfig, r.Namespace)
	if err != nil {
		log.Printf("unable to create Kubernetes client: %v\n", err)
		os.Exit(2)
	}

	err = r.Create(context.Background(), client)
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(2)
	}

	os.Exit(0)
}
