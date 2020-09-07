package trustyserver

import "github.com/go-phorce/trusty/config"

// TrustyServer is the production implementation of the Server interface
type TrustyServer struct {
	cfg *config.Configuration
}
