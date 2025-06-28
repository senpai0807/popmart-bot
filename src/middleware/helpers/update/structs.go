package update

import _ "embed"

//go:embed version.json
var VersionData []byte

type VersionInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}
