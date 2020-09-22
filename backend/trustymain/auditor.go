package trustymain

type auditornoop struct{}

func (a auditornoop) Close() error {
	return nil
}

// Audit create an audit event
func (a auditornoop) Audit(
	source string,
	eventType string,
	identity string,
	contextID string,
	raftIndex uint64,
	message string) {
	// {contextID}:{identity}:{raftIndex}:{source}:{type}:{message}
	logger.Infof("audit:%s:%s:%s:%s:%d:%s\n",
		source, eventType, identity, contextID, raftIndex, message)
}
