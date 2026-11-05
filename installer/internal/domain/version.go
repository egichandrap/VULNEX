package domain

// Version holds release metadata for the VULNEX CLI and deployment stack.
type Version struct {
	Number    string
	Commit    string
	BuildDate string
}

// CurrentVersion is the authoritative version of the current build.
var CurrentVersion = Version{
	Number:    "1.1.1",
	Commit:    "HEAD",
	BuildDate: "release",
}

// String returns formatted version representation.
func (v Version) String() string {
	return "VULNEX CLI v" + v.Number + " (" + v.Commit + ", " + v.BuildDate + ")"
}
