package version

// GENERATED FILE DO NOT CHECK IN TO GIT!

var currentVersion = Info{
	Major:  0,
	Minor:  0,
	Commit: 0,
	Build:  "0.1.1",
}

func init() {
	currentVersion.PopulateFromBuild()
}

// Current returns the current version [set by the build]
func Current() Info {
	return currentVersion
}
