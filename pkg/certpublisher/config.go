package certpublisher

// Config provides configuration for Certification Authority
type Config struct {
	CertsBucket string `json:"cert_bucket" yaml:"cert_bucket"`
	CRLBucket   string `json:"crl_bucket" yaml:"crl_bucket"`
}
