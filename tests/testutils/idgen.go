package testutils

import (
	"github.com/martinisecurity/trusty/backend/appcontainer"
	"github.com/martinisecurity/trusty/backend/db"
)

// IDGenerator returns static ID generator for the app
func IDGenerator() db.IDGenerator {
	return appcontainer.IDGenerator
}
