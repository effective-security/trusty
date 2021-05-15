package mockpb

import (
	"context"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/gogo/protobuf/proto"
)

// MockCertInfoServer for testing
type MockCertInfoServer struct {
	pb.CertInfoServiceServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockCertInfoServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// Roots returns the root CAs
func (m *MockCertInfoServer) Roots(context.Context, *pb.EmptyRequest) (*pb.RootsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*pb.RootsResponse), nil
}
