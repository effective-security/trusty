package config

// RegistrationAuthority contains configuration info for RA
type RegistrationAuthority struct {
	// PrivateRoots specifies the list of private Root Certs files.
	PrivateRoots []string `json:"private_roots,omitempty" yaml:"private_roots,omitempty"`

	// PublicRoots specifies the list of public Root Certs files.
	PublicRoots []string `json:"public_roots,omitempty" yaml:"public_roots,omitempty"`

	Publisher Publisher `json:"publisher,omitempty" yaml:"publisher,omitempty"`

	// GenCerts specifies a list of certs to generate on startup
	GenCerts GenCerts `json:"gen_certs" yaml:"gen_certs"`
}

// Publisher ontains configuration info for Publisher
type Publisher struct {
	BaseURL     string `json:"base_url" yaml:"base_url"`
	CertsBucket string `json:"cert_bucket" yaml:"cert_bucket"`
	CRLBucket   string `json:"crl_bucket" yaml:"crl_bucket"`
}

// GenCerts contains configuration info for the auto generated certificates
type GenCerts struct {
	// Schedule specifies a schedule for renewal task in format documented in /pkg/tasks. If it is empty, then the default value is used.
	Schedule string `json:"schedule" yaml:"schedule"`

	Profiles []GenCert `json:"profiles" yaml:"profiles"`
}

// GenCert contains configuration info for the auto generated certificate
type GenCert struct {

	// Disabled specifies if the certificate disabled to use
	Disabled bool `json:"disabled" yaml:"disabled"`

	// CertFile specifies location of the cert
	CertFile string `json:"cert_file" yaml:"cert_file"`

	// KeyFile specifies location of the key
	KeyFile string `json:"key_file" yaml:"key_file"`

	// Profile specifies the certificate profile
	Profile string `json:"profile" yaml:"profile"`

	// Renewal specifies value in 165h00m00s format for renewal before expiration date
	Renewal string `json:"renewal" yaml:"renewal"`

	// SAN specifies additional alt names to be included in the certificate
	SAN []string `json:"san" yaml:"san"`

	// CsrProfile specifies a profile for Certificate Request
	CsrProfile string `json:"csr_profile" yaml:"csr_profile"`
}
