package certinit_test

import (
	"context"
	"os"
	"testing"

	"github.com/ekspand/trusty/kubeca/certinit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	capi "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateFail(t *testing.T) {

	tcases := []struct {
		r   *certinit.Request
		err string
	}{
		{
			r:   &certinit.Request{},
			err: "missing required namespace parameter",
		},
		{
			r:   &certinit.Request{Namespace: "ns"},
			err: "missing required pod name parameter",
		},
		{
			r:   &certinit.Request{Namespace: "ns", PodName: "podname"},
			err: "missing required signer name parameter",
		},
	}

	for _, tc := range tcases {
		err := tc.r.Create(context.Background(), nil)
		require.Error(t, err)
		assert.Equal(t, tc.err, err.Error())
	}
}

func TestCreate(t *testing.T) {
	r := &certinit.Request{
		Namespace:          "test",
		PodName:            "pod.metadata.name",
		CertDir:            "/tmp/tests/certinit",
		ClusterDomain:      "test",
		Labels:             "key=value",
		QueryK8s:           true,
		SAN:                "email@test.com,test.com,127.0.0.1,spifee://test/service",
		ServiceNames:       "service1",
		IncludeUnqualified: true,
		SignerName:         "trusty.svc/peer",
	}
	os.RemoveAll(r.CertDir)

	p := mockedPods{}
	s := mockedServices{}
	c := mockedCertificates{}
	mocked := &certinit.CertClient{
		Pods:         &p,
		Services:     &s,
		Certificates: &c,
	}

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Hostname:  "pod1",
			Subdomain: "domain.com",
		},
	}
	p.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(pod, nil)

	list := v1.ServiceList{
		Items: []v1.Service{
			{Spec: v1.ServiceSpec{ClusterIP: "10.0.0.1"}},
		},
	}
	s.On("List", mock.Anything, mock.Anything).Return(&list, nil)

	csr1 := capi.CertificateSigningRequest{}
	c.On("Get", mock.Anything, mock.Anything, mock.Anything).Times(1).Return(&csr1, nil)
	csr2 := capi.CertificateSigningRequest{
		Status: capi.CertificateSigningRequestStatus{
			Certificate: []byte(`cert`),
		},
	}
	c.On("Get", mock.Anything, mock.Anything, mock.Anything).Times(1).Return(&csr2, nil)
	c.On("Create", mock.Anything, mock.Anything).Return(&csr1, nil)

	err := r.Create(context.Background(), mocked)
	require.Error(t, err)
	assert.Equal(t, "unable to save key: open /tmp/tests/certinit/tls.key: no such file or directory", err.Error())

	os.MkdirAll(r.CertDir, 0755)
	defer os.RemoveAll(r.CertDir)

	err = r.Create(context.Background(), mocked)
	require.NoError(t, err)

	r.SignerName = "trusty.svc/unsupported"
	err = r.Create(context.Background(), mocked)
	require.Error(t, err)
	assert.Equal(t, "unsupported profile: trusty.svc/unsupported", err.Error())
}

type mockedPods struct {
	mock.Mock
}

func (m *mockedPods) Get(ctx context.Context, name string, opts metaV1.GetOptions) (*v1.Pod, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

type mockedServices struct {
	mock.Mock
}

func (m *mockedServices) List(ctx context.Context, opts metaV1.ListOptions) (*v1.ServiceList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*v1.ServiceList), args.Error(1)
}

type mockedCertificates struct {
	mock.Mock
}

func (m *mockedCertificates) Create(ctx context.Context, certificateSigningRequest *capi.CertificateSigningRequest, opts metaV1.CreateOptions) (*capi.CertificateSigningRequest, error) {
	args := m.Called(ctx, certificateSigningRequest, opts)
	return args.Get(0).(*capi.CertificateSigningRequest), args.Error(1)
}

func (m *mockedCertificates) Get(ctx context.Context, name string, opts metaV1.GetOptions) (*capi.CertificateSigningRequest, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*capi.CertificateSigningRequest), args.Error(1)
}
