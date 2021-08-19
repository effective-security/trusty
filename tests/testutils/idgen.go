package testutils

import (
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/db"
)

// IDGenerator returns static ID generator for the app
func IDGenerator() db.IDGenerator {
	return appcontainer.IDGenerator
}
