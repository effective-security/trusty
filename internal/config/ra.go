package config

// RegistrationAuthority contains configuration info for RA
type RegistrationAuthority struct {
	// PrivateRoots specifies the list of private Root Certs files.
	PrivateRoots []string `json:"private_roots,omitempty" yaml:"private_roots,omitempty"`

	// PublicRoots specifies the list of public Root Certs files.
	PublicRoots []string `json:"public_roots,omitempty" yaml:"public_roots,omitempty"`
}
