package version

// GENERATED FILE DO NOT CHECK IN TO GIT!

import (
	"runtime"
)

var currentVersion = Info{
	Major:   0,
	Minor:   0,
	Commit:  0,
	Runtime: runtime.Version(),
	Build:   "v1.0.434-denislenovo",
}

func init() {
	currentVersion.PopulateFromBuild()
}

// Current returns the current version [set by the build]
func Current() Info {
	return currentVersion
}
