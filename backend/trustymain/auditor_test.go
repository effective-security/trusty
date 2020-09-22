package trustymain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditornoop(t *testing.T) {
	a := auditornoop{}

	a.Audit("source", "eventType", "identity", "contextID", 0, "message")
	err := a.Close()
	assert.NoError(t, err)
}
