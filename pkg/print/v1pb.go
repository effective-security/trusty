package print

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/olekukonko/tablewriter"
)

// ServerVersion prints ServerVersion
func ServerVersion(w io.Writer, r *trustypb.ServerVersion) {
	fmt.Fprintf(w, "%s (%s)\n", r.Build, r.Runtime)
}

// ServerStatusResponse prints trustypb.ServerStatusResponse
func ServerStatusResponse(w io.Writer, r *trustypb.ServerStatusResponse) {

	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Append([]string{"Name", r.Status.Name})
	table.Append([]string{"Node", r.Status.Nodename})
	table.Append([]string{"Host", r.Status.Hostname})
	table.Append([]string{"Listen URLs", strings.Join(r.Status.ListenUrls, ",")})
	table.Append([]string{"Version", r.Version.Build})
	table.Append([]string{"Runtime", r.Version.Runtime})

	startedAt := time.Unix(r.Status.StartedAt, 0)
	uptime := time.Now().Sub(startedAt) / time.Second * time.Second
	table.Append([]string{"Started", startedAt.Format(time.RFC3339)})
	table.Append([]string{"Uptime", uptime.String()})

	table.Render()
	fmt.Fprintln(w)
}

// CallerStatusResponse prints trustypb.CallerStatusResponse
func CallerStatusResponse(w io.Writer, r *trustypb.CallerStatusResponse) {
	table := tablewriter.NewWriter(w)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Append([]string{"Name", r.Name})
	table.Append([]string{"ID", r.Id})
	table.Append([]string{"Role", r.Role})
	table.Render()
	fmt.Fprintln(w)
}
