package tasks_test

import (
	"testing"

	"github.com/effective-security/trusty/backend/tasks"
	"github.com/effective-security/trusty/backend/tasks/certsmonitor"
	"github.com/stretchr/testify/require"
)

var factories = map[string]tasks.Factory{
	certsmonitor.TaskName: certsmonitor.Factory,
}

func Test_invalidArgs(t *testing.T) {
	for _, f := range factories {
		fact := f(nil, "", "")
		require.NotNil(t, fact)
	}
}
