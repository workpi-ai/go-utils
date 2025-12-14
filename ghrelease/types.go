package ghrelease

import (
	"time"

	"github.com/google/go-github/v68/github"
)

type UpdaterConfig struct {
	RepoOwner          string
	RepoName           string
	DestDir            string
	MetadataFilename   string
	ExtractFilter      ExtractFilter
	RequestTimeout     time.Duration
	DownloadTimeout    time.Duration
}

type ExtractFilter interface {
	ShouldExtract(filename string) bool
}

type Updater struct {
	config UpdaterConfig
	client *github.Client
}

type Metadata struct {
	Version     string `json:"version"`
	LastCheckAt string `json:"last_check_at"`
}
