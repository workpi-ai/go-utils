package ghrelease

import (
	"time"

	"github.com/google/go-github/v68/github"
)

type ExtractTarget struct {
	PathTransformer PathTransformer
	DestDir         string
}

type UpdaterConfig struct {
	RepoOwner       string
	RepoName        string
	MetadataFile    string
	Targets         []ExtractTarget
	RequestTimeout  time.Duration
	DownloadTimeout time.Duration
}

type PathTransformer interface {
	Transform(filename string) string
}

type Updater struct {
	config UpdaterConfig
	client *github.Client
}

type Metadata struct {
	Version     string `json:"version"`
	LastCheckAt string `json:"last_check_at"`
}
