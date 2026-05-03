// Package version trägt Build-Metadaten, die per -ldflags injiziert werden.
package version

// Build-Variablen, gesetzt via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info ist die JSON-serialisierbare Variante der Build-Metadaten.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

// Get liefert die aktuellen Build-Metadaten.
func Get() Info {
	return Info{Version: Version, Commit: Commit, BuildDate: BuildDate}
}
