package testutils

import (
	"os"
	"time"

	"github.com/go-phorce/trusty/internal/db"
	"github.com/sony/sonyflake"
)

var idGenerator = sonyflake.NewSonyflake(sonyflake.Settings{
	StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	MachineID: func() (uint16, error) {
		return uint16(os.Getpid()), nil
	},
})

// IDGenerator returns static ID generator for the process
func IDGenerator() db.IDGenerator {
	return idGenerator
}
