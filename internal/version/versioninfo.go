package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Info describes a version of an executable
type Info struct {
	Major   uint   `json:"major"`
	Minor   uint   `json:"minor"`
	Commit  uint   `json:"commit"`
	Build   string `json:"build"`
	Runtime string `json:"runtime"`
	flt     float32
}

// PopulateFromBuild will parse the major/minor values from the build string
// the build string is expected to be in the format
// major.minor-commit
// and can be populated from git using
//
//	GIT_VERSION := $(shell git describe --dirty --always --tags --long)
//
// and then using gofmt to substitute it into a template
func (v *Info) PopulateFromBuild() {
	bld := strings.Split(v.Build, "-")[0]

	fmt.Sscanf(bld, "v%d.%d.%d", &v.Major, &v.Minor, &v.Commit)
	fmt.Sscanf(bld, "v%f.", &v.flt)

	var fCommit float32
	switch {
	case v.Commit > 1000000:
		fCommit = 0.0000001 * float32(v.Commit)
	case v.Commit > 100000:
		fCommit = 0.000001 * float32(v.Commit)
	case v.Commit > 10000:
		fCommit = 0.00001 * float32(v.Commit)
	case v.Commit > 1000:
		fCommit = 0.0001 * float32(v.Commit)
	case v.Commit > 100:
		fCommit = 0.001 * float32(v.Commit)
	case v.Commit > 10:
		fCommit = 0.01 * float32(v.Commit)
	case v.Commit > 1:
		fCommit = 0.1 * float32(v.Commit)
	}

	v.flt = v.flt*100 + fCommit
	v.Runtime = runtime.Version()
}

func (v Info) String() string {
	return v.Build
}

// GreaterOrEqual returns true if the version 'v' is the same or new that the supplied parameter 'other'
// This only examines the Major & Minor field (as the SHA in Build provides no ordering indication)
func (v Info) GreaterOrEqual(than Info) bool {
	if v.Major > than.Major {
		return true
	}
	if v.Major < than.Major {
		return false
	}
	return v.Minor >= than.Minor
}

// Float returns the version Major/Minor as a float Major.Minor
// e.g. given Major:3 Minor:52001, it'll return 3.52001
// this is only valid if PopulateFromBuild has been called.
func (v Info) Float() float32 {
	return v.flt
}
