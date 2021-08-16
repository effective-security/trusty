package testutils

import "github.com/go-phorce/dolly/tasks"

// MockTask provides dummy task
type MockTask struct {
	Tasks []tasks.Task
}

// Add adds a task to a pool of scheduled tasks
func (m *MockTask) Add(t tasks.Task) tasks.Scheduler {
	m.Tasks = append(m.Tasks, t)
	return m
}

// Clear will delete all scheduled tasks
func (m *MockTask) Clear() {
	m.Tasks = nil
}

// Count returns the number of registered tasks
func (m *MockTask) Count() int {
	return len(m.Tasks)
}

// IsRunning return the status
func (m *MockTask) IsRunning() bool {
	return false
}

// Start all the pending tasks
func (m *MockTask) Start() error {
	return nil
}

// Stop the scheduler
func (m *MockTask) Stop() error {
	return nil
}
