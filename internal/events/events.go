package events

const (
	// EvtSourceCluster specifies source for Cluster
	EvtSourceCluster = "cluster"

	// EvtSourceCatalog specifies source for Catalog
	EvtSourceCatalog = "catalog"

	// EvtSourceCA specifies source for CA
	EvtSourceCA = "CA"

	// EvtSourceStatus specifies source for service Status
	EvtSourceStatus = "status"
)

const (
	// EvtServiceStarted specifies Service Started event
	EvtServiceStarted = "service started"

	// EvtServiceStopped specifies Service Stopped event
	EvtServiceStopped = "service stopped"

	// EvtDataSnapshotCreated specifies audit event
	EvtDataSnapshotCreated = "snapshot_created"
	// EvtDataSnapshotDeleted specifies audit event
	EvtDataSnapshotDeleted = "snapshot_deleted"

	// EvtCertificateIssued specifies audit event
	EvtCertificateIssued = "certificate_issued"
	// EvtCertificateRevoked specifies audit event
	EvtCertificateRevoked = "certificate_revoked"
	// EvtCRLPublished specifies audit event
	EvtCRLPublished = "crl_published"

	// EvtIssuerUnregistered specifies audit event
	EvtIssuerUnregistered = "issuer_unregistered"
)
