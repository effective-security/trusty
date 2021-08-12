package config

// RegistrationAuthority contains configuration info for RA
type RegistrationAuthority struct {
	// PrivateRoots specifies the list of private Root Certs files.
	PrivateRoots []string `json:"private_roots,omitempty" yaml:"private_roots,omitempty"`

	// PublicRoots specifies the list of public Root Certs files.
	PublicRoots []string `json:"public_roots,omitempty" yaml:"public_roots,omitempty"`

	Publisher CertPublisher `json:"publisher,omitempty" yaml:"publisher,omitempty"`
}

// CertPublisher ontains configuration info for Publisher
type CertPublisher struct {
	CertsBucket string `json:"cert_bucket" yaml:"cert_bucket"`
	CRLBucket   string `json:"crl_bucket" yaml:"crl_bucket"`
}
