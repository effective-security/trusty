package certinit

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-phorce/dolly/xlog"
	"github.com/pkg/errors"
	capi "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty", "certinit")

// Request parameters
type Request struct {
	// Namespace as defined by pod.metadata.namespace
	Namespace string
	// PodName name as defined by pod.metadata.name
	PodName string
	// CertDir is directory where the TLS certs should be written
	CertDir string
	// ClusterDomain specifies kubernetes cluster domain
	ClusterDomain string
	// Labels to include in CertificateSigningRequest object; comma separated list of key=value
	Labels string
	// QueryK8s specifies to query kubernetes for names appropriate to this Pod
	QueryK8s bool
	// SAN is additional comma separated DNS, IP, URI or Emails to include in SAN
	SAN string
	// ServiceNames specifies additional service names that resolve to this Pod; comma separated
	ServiceNames string
	// IncludeUnqualified specifies to include unqualified .svc domains in names from --query-k8s
	IncludeUnqualified bool
	// SignerName specifies the signer name
	SignerName string

	san       []string
	labelsMap map[string]string
}

// CertClient provides minimum interfaces to create a CSR
type CertClient struct {
	Pods         MinPods
	Services     MinServices
	Certificates MinCertificates
}

// MinPods is minimum Pods interface
type MinPods interface {
	Get(ctx context.Context, name string, opts metaV1.GetOptions) (*v1.Pod, error)
}

// MinServices is minimum Services interface
type MinServices interface {
	List(ctx context.Context, opts metaV1.ListOptions) (*v1.ServiceList, error)
}

// MinCertificates is minimum Certificates interface
type MinCertificates interface {
	Create(ctx context.Context, certificateSigningRequest *capi.CertificateSigningRequest, opts metaV1.CreateOptions) (*capi.CertificateSigningRequest, error)
	Get(ctx context.Context, name string, opts metaV1.GetOptions) (*capi.CertificateSigningRequest, error)
}

// Create certificate request and wait for issuance
func (r *Request) Create(ctx context.Context, client *CertClient) error {
	if r.Namespace == "" {
		return errors.New("missing required namespace parameter")
	}
	if r.PodName == "" {
		return errors.New("missing required pod name parameter")
	}
	if r.SignerName == "" {
		return errors.New("missing required signer name parameter")
	}

	logger.Infof("Request, ns=%s, pod=%s, signer=%s, san=%q",
		r.Namespace, r.PodName, r.SignerName, r.SAN)

	// Gather the list of labels that will be added to the CreateCertificateSigningRequest object
	r.labelsMap = make(map[string]string)
	for _, n := range strings.Split(r.Labels, ",") {
		if n == "" {
			continue
		}
		s := strings.Split(n, "=")
		if len(s) == 2 {
			label, key := s[0], s[1]
			if label == "" {
				continue
			}
			r.labelsMap[label] = key
		}
	}

	if r.QueryK8s {
		pod, err := client.Pods.Get(context.Background(), r.PodName, metaV1.GetOptions{})
		if err != nil {
			return errors.WithMessagef(err, "failed to query pod %q in namespace %q", r.PodName, r.Namespace)
		}

		r.san, err = getNamesForPod(ctx, client.Services, *pod, r.ClusterDomain, r.IncludeUnqualified)
		if err != nil {
			return errors.WithMessagef(err, "failed to query names for pod %q in namespace %q", r.PodName, r.Namespace)
		}
	}

	for _, s := range strings.Split(r.SAN, ",") {
		if s == "" {
			continue
		}
		r.san = append(r.san, s)
	}

	err := r.requestCertificate(ctx, client.Certificates)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// NewClient returns new client
func NewClient(kubeconfig, namespace string) (*CertClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CertClient{
		Pods:         c.CoreV1().Pods(namespace),
		Services:     c.CoreV1().Services(namespace),
		Certificates: c.CertificatesV1().CertificateSigningRequests(),
	}, nil
}

// getNamesForPod returns the DNS names and IPs that a given POD is permitted to have,
// either in its own right or by dint of matching services.
// Does not currently pay attention to static Endpoints.
func getNamesForPod(ctx context.Context, client MinServices, pod v1.Pod, clusterDomain string, allowUnqualified bool) (san []string, err error) {
	san = []string{fmt.Sprintf("%s.%s.pod.%s", ipToName(pod.Status.PodIP), pod.Namespace, clusterDomain)}
	if pod.Spec.Hostname != "" && pod.Spec.Subdomain != "" {
		san = append(san, fmt.Sprintf("%s.%s.%s.svc.%s", pod.Spec.Hostname, pod.Spec.Subdomain, pod.Namespace, clusterDomain))
		if allowUnqualified {
			san = append(san, fmt.Sprintf("%s.%s.%s.svc", pod.Spec.Hostname, pod.Spec.Subdomain, pod.Namespace))
		}
	}

	podLabels := labels.Set(pod.Labels)

	serviceList, err := client.List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, service := range serviceList.Items {
		if service.Spec.Selector == nil {
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(podLabels) {
			san = append(san, fmt.Sprintf("%s.%s.svc.%s", service.Name, service.Namespace, clusterDomain))
			if allowUnqualified {
				san = append(san, fmt.Sprintf("%s.%s.svc", service.Name, service.Namespace))
			}

			if service.Spec.Type == v1.ServiceTypeExternalName {
				if service.Spec.ExternalName != "" {
					san = append(san, service.Spec.ExternalName)
				}
			} else {
				san = append(san, service.Spec.ClusterIP)
			}

			if len(service.Spec.ExternalIPs) > 0 {
				san = append(san, service.Spec.ExternalIPs...)
			}
		}
	}

	return
}

func ipToName(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}
