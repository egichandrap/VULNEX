package domain

// Version holds release metadata for the VULNEX CLI and deployment stack.
type Version struct {
	Number    string
	Commit    string
	BuildDate string
}

// CurrentVersion is the authoritative version of the current build.
var CurrentVersion = Version{
	Number:    "0.1.0-dev",
	Commit:    "HEAD",
	BuildDate: "unreleased",
}

// String returns formatted version representation.
func (v Version) String() string {
	return "VULNEX CLI v" + v.Number + " (" + v.Commit + ", " + v.BuildDate + ")"
}
