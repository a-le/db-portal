package jsminifier

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dchest/jsmin"
)

type JSMinifierStatus struct {
	ManifestPath        string
	MinifiedPath        string
	SourceFiles         []string
	LatestSourceModTime time.Time
	MinifiedModTime     time.Time
	Expired             bool
}

func NewJSMinifyStatus(manifestPath string, jsMinPath string) (status JSMinifierStatus, err error) {
	status.ManifestPath = manifestPath
	status.MinifiedPath = jsMinPath

	// Read manifest.json
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return
	}

	// Parse JSON array of file paths
	if err = json.Unmarshal(manifestBytes, &status.SourceFiles); err != nil {
		return
	}

	// Find the most recent mod time among files
	if len(status.SourceFiles) == 0 {
		err = fmt.Errorf("%s should not be empty", status.ManifestPath)
		return
	}
	for _, file := range status.SourceFiles {
		file = "." + file // path is relative in server context
		var info os.FileInfo
		info, err = os.Stat(file)
		if err != nil {
			return
		}
		if info.ModTime().After(status.LatestSourceModTime) {
			status.LatestSourceModTime = info.ModTime()
		}
	}

	// Get mod time of jsMinPath
	info, err := os.Stat(jsMinPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If min file doesn't exist, treat as expired
			status.Expired = true
		}
		return
	}
	status.MinifiedModTime = info.ModTime()

	status.Expired = status.LatestSourceModTime.After(status.MinifiedModTime)

	return
}

func (status JSMinifierStatus) Combinify() (newStatus JSMinifierStatus, err error) {
	newStatus = status
	var combined []byte
	for _, file := range status.SourceFiles {
		file = "." + file // path is relative in server context
		var b []byte
		b, err = os.ReadFile(file)
		if err != nil {
			return
		}
		if len(combined) > 0 {
			combined = append(combined, "\n"...)
		}
		combined = append(combined, b...)
	}

	var minified []byte
	minified, err = jsmin.Minify(combined)
	if err != nil {
		return
	}

	if err = os.WriteFile(status.MinifiedPath, minified, 0644); err != nil {
		return
	}

	info, _ := os.Stat(status.MinifiedPath)
	newStatus.MinifiedModTime = info.ModTime()
	newStatus.Expired = false

	return
}
