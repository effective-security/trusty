package print

import (
	"fmt"
	"io"

	"github.com/go-phorce/trusty/api/v1/trustypb"
)

// ServerVersion will print .ServerStatus
func ServerVersion(w io.Writer, r *trustypb.ServerVersion) {
	fmt.Fprintf(w, "%s (%s)\n", r.Build, r.Runtime)
}
