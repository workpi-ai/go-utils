package ghrelease

import (
	"testing"
	"time"
)

func TestNewUpdater(t *testing.T) {
	tests := []struct {
		name   string
		config UpdaterConfig
		want   UpdaterConfig
	}{
		{
			name: "all fields provided",
			config: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   5 * time.Second,
				DownloadTimeout:  60 * time.Second,
			},
			want: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   5 * time.Second,
				DownloadTimeout:  60 * time.Second,
			},
		},
		{
			name: "use default timeouts",
			config: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				ExtractFilter:    &AllFilesFilter{},
			},
			want: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   defaultRequestTimeout,
				DownloadTimeout:  defaultDownloadTimeout,
			},
		},
		{
			name: "use default metadata filename",
			config: UpdaterConfig{
				RepoOwner:       "owner",
				RepoName:        "repo",
				DestDir:         "/tmp/dest",
				ExtractFilter:   &AllFilesFilter{},
				RequestTimeout:  5 * time.Second,
				DownloadTimeout: 60 * time.Second,
			},
			want: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: defaultMetadataFile,
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   5 * time.Second,
				DownloadTimeout:  60 * time.Second,
			},
		},
		{
			name: "use default filter",
			config: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				RequestTimeout:   5 * time.Second,
				DownloadTimeout:  60 * time.Second,
			},
			want: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: "custom.json",
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   5 * time.Second,
				DownloadTimeout:  60 * time.Second,
			},
		},
		{
			name: "use all defaults",
			config: UpdaterConfig{
				RepoOwner: "owner",
				RepoName:  "repo",
				DestDir:   "/tmp/dest",
			},
			want: UpdaterConfig{
				RepoOwner:        "owner",
				RepoName:         "repo",
				DestDir:          "/tmp/dest",
				MetadataFilename: defaultMetadataFile,
				ExtractFilter:    &AllFilesFilter{},
				RequestTimeout:   defaultRequestTimeout,
				DownloadTimeout:  defaultDownloadTimeout,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := NewUpdater(tt.config)

			if updater == nil {
				t.Fatal("NewUpdater returned nil")
			}

			if updater.config.RepoOwner != tt.want.RepoOwner {
				t.Errorf("RepoOwner = %v, want %v", updater.config.RepoOwner, tt.want.RepoOwner)
			}
			if updater.config.RepoName != tt.want.RepoName {
				t.Errorf("RepoName = %v, want %v", updater.config.RepoName, tt.want.RepoName)
			}
			if updater.config.DestDir != tt.want.DestDir {
				t.Errorf("DestDir = %v, want %v", updater.config.DestDir, tt.want.DestDir)
			}
			if updater.config.MetadataFilename != tt.want.MetadataFilename {
				t.Errorf("MetadataFilename = %v, want %v", updater.config.MetadataFilename, tt.want.MetadataFilename)
			}
			if updater.config.RequestTimeout != tt.want.RequestTimeout {
				t.Errorf("RequestTimeout = %v, want %v", updater.config.RequestTimeout, tt.want.RequestTimeout)
			}
			if updater.config.DownloadTimeout != tt.want.DownloadTimeout {
				t.Errorf("DownloadTimeout = %v, want %v", updater.config.DownloadTimeout, tt.want.DownloadTimeout)
			}
			if updater.config.ExtractFilter == nil {
				t.Error("ExtractFilter is nil")
			}

			if updater.client == nil {
				t.Error("client is nil")
			}
		})
	}
}
