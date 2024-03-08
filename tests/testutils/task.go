package testutils

import (
	"github.com/effective-security/porto/pkg/tasks"
)

// MockScheduler provides dummy task
type MockScheduler struct {
	Tasks []tasks.Task
}

// Get gets a task from the pool of scheduled tasks
func (m *MockScheduler) Get(id string) tasks.Task {
	for _, t := range m.Tasks {
		if t.ID() == id {
			return t
		}
	}
	return nil
}

// Add adds a task to a pool of scheduled tasks
func (m *MockScheduler) Add(t tasks.Task) tasks.Scheduler {
	m.Tasks = append(m.Tasks, t)
	return m
}

// Clear will delete all scheduled tasks
func (m *MockScheduler) Clear() {
	m.Tasks = nil
}

func (m *MockScheduler) SetPublisher(p tasks.Publisher) tasks.Scheduler {
	return m
}

// List returns the tasks
func (m *MockScheduler) List() []tasks.Task {
	return m.Tasks[:]
}

// Count returns the number of registered tasks
func (m *MockScheduler) Count() int {
	return len(m.Tasks)
}

// IsRunning return the status
func (m *MockScheduler) IsRunning() bool {
	return false
}

// Start all the pending tasks
func (m *MockScheduler) Start() error {
	return nil
}

// Stop the scheduler
func (m *MockScheduler) Stop() error {
	return nil
}

// RunPending executes all pending tasks
func (m *MockScheduler) RunPending() {
	for _, t := range m.Tasks {
		if t.ShouldRun() {
			t.Run()
		}
	}
}

// Publish the tasks to Publisher
func (m *MockScheduler) Publish() {
}
