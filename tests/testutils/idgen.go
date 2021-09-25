package testutils

import (
	"github.com/martinisecurity/trusty/internal/appcontainer"
	"github.com/martinisecurity/trusty/internal/db"
)

// IDGenerator returns static ID generator for the app
func IDGenerator() db.IDGenerator {
	return appcontainer.IDGenerator
}
