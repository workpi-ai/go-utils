package ghrelease

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

const (
	defaultRequestTimeout  = 3 * time.Second
	defaultDownloadTimeout = 30 * time.Second
	defaultMetadataFile    = "metadata.json"
	defaultDirPerm         = 0755
	defaultFilePerm        = 0644
)

func NewUpdater(config UpdaterConfig) *Updater {
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	if config.DownloadTimeout == 0 {
		config.DownloadTimeout = defaultDownloadTimeout
	}
	if config.MetadataFilename == "" {
		config.MetadataFilename = defaultMetadataFile
	}
	if config.ExtractFilter == nil {
		config.ExtractFilter = &AllFilesFilter{}
	}

	return &Updater{
		config: config,
		client: github.NewClient(nil),
	}
}

func (u *Updater) Update() error {
	latestVersion, err := u.getLatestVersion()
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}

	localVersion := u.getLocalVersion()

	if latestVersion == localVersion {
		u.saveLocalVersion(localVersion)
		return nil
	}

	if err := u.downloadRelease(latestVersion); err != nil {
		return fmt.Errorf("download release: %w", err)
	}

	u.saveLocalVersion(latestVersion)

	return nil
}

func (u *Updater) getLatestVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.config.RequestTimeout)
	defer cancel()

	release, _, err := u.client.Repositories.GetLatestRelease(ctx, u.config.RepoOwner, u.config.RepoName)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	if release.TagName == nil {
		return "", fmt.Errorf("release tag name is nil")
	}

	return *release.TagName, nil
}

func (u *Updater) getLocalVersion() string {
	path := filepath.Join(u.config.DestDir, u.config.MetadataFilename)

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var v Metadata
	if err := json.Unmarshal(data, &v); err != nil {
		return ""
	}

	return v.Version
}

func (u *Updater) downloadRelease(version string) error {
	ctx, cancel := context.WithTimeout(context.Background(), u.config.DownloadTimeout)
	defer cancel()

	release, _, err := u.client.Repositories.GetReleaseByTag(ctx, u.config.RepoOwner, u.config.RepoName, version)
	if err != nil {
		return fmt.Errorf("failed to get release info: %w", err)
	}

	if release.ZipballURL == nil {
		return fmt.Errorf("release zipball_url is nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *release.ZipballURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download zipball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, *release.ZipballURL)
	}

	return u.extractZip(resp.Body)
}

func (u *Updater) extractZip(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		if !u.config.ExtractFilter.ShouldExtract(file.Name) {
			continue
		}

		relPath := u.stripRootDir(file.Name)
		if relPath == "" {
			continue
		}

		destPath := filepath.Join(u.config.DestDir, relPath)

		if err = os.MkdirAll(filepath.Dir(destPath), defaultDirPerm); err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		f, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(f, rc); err != nil {
			return err
		}
	}

	return nil
}

func (u *Updater) stripRootDir(filename string) string {
	idx := strings.Index(filename, "/")
	if idx == -1 {
		return ""
	}
	return filename[idx+1:]
}

func (u *Updater) saveLocalVersion(version string) error {
	v := Metadata{
		Version:     version,
		LastCheckAt: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(u.config.DestDir, u.config.MetadataFilename)
	return os.WriteFile(path, data, defaultFilePerm)
}
