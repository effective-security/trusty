package audit

import (
	"io"
)

// Source declares the general area that the audit event was raised from
type Source interface {
	ID() int
	String() string
}

// EventType defines a specific event from the Source
type EventType interface {
	ID() int
	String() string
}

// Auditor defines an interface that can receive information about audit events
type Auditor interface {
	// Call at shutdown to cleanly close the audit destination
	io.Closer

	// Audit event
	// source indicates the area that the event was triggered by
	// eventType indicates the specific event that occured
	// identity specifies the identity of the user that triggered this event, typically this is <role>/<cn>
	// contextID specifies the request ContextID that the event was triggered in [this can be used for cross service correlation of logs]
	// raftIndex indicates the index# of the raft log in RAFT that the event occured in [if applicable]
	// message contains any additional information about this event that is eventType specific
	Audit(
		source string,
		eventType string,
		identity string,
		contextID string,
		raftIndex uint64,
		message string)
}
