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
	defaultDirPerm         = 0755
	defaultFilePerm        = 0644
)

func NewUpdater(config UpdaterConfig) (*Updater, error) {
	if config.RepoOwner == "" {
		return nil, fmt.Errorf("repo owner cannot be empty")
	}
	if config.RepoName == "" {
		return nil, fmt.Errorf("repo name cannot be empty")
	}
	if len(config.Targets) == 0 {
		return nil, fmt.Errorf("targets cannot be empty")
	}
	for i, target := range config.Targets {
		if target.PathTransformer == nil {
			return nil, fmt.Errorf("target[%d].PathTransformer cannot be nil", i)
		}
		if target.DestDir == "" {
			return nil, fmt.Errorf("target[%d].DestDir cannot be empty", i)
		}
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}
	if config.DownloadTimeout == 0 {
		config.DownloadTimeout = defaultDownloadTimeout
	}

	return &Updater{
		config: config,
		client: github.NewClient(nil),
	}, nil
}

func (u *Updater) Update() error {
	latestVersion, err := u.getLatestVersion()
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}

	localVersion := u.getLocalVersion()

	needsDownload := latestVersion != localVersion || u.needsRedownload()
	if !needsDownload {
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
	data, err := os.ReadFile(u.config.MetadataFile)
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

		relPath := u.stripRootDir(file.Name)
		if relPath == "" {
			continue
		}

		for _, target := range u.config.Targets {
			destPath := target.PathTransformer.Transform(relPath)
			if destPath == "" {
				continue
			}

			fullPath := filepath.Join(target.DestDir, destPath)

			if err := u.extractFile(file, fullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *Updater) extractFile(file *zip.File, destPath string) error {
	destPath = filepath.Clean(destPath)
	if !filepath.IsAbs(destPath) {
		return fmt.Errorf("destination path must be absolute: %s", destPath)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), defaultDirPerm); err != nil {
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

	_, err = io.Copy(f, rc)
	return err
}

func (u *Updater) needsRedownload() bool {
	for _, target := range u.config.Targets {
		stat, err := os.Stat(target.DestDir)
		if os.IsNotExist(err) {
			return true
		}
		if err != nil {
			return true
		}
		if !stat.IsDir() {
			return true
		}

		entries, err := os.ReadDir(target.DestDir)
		if err != nil {
			return true
		}
		if len(entries) == 0 {
			return true
		}
	}
	return false
}

func (u *Updater) stripRootDir(filename string) string {
	idx := strings.Index(filename, "/")
	if idx == -1 {
		return ""
	}
	return filename[idx+1:]
}

func (u *Updater) saveLocalVersion(version string) error {
	if err := os.MkdirAll(filepath.Dir(u.config.MetadataFile), defaultDirPerm); err != nil {
		return err
	}

	v := Metadata{
		Version:     version,
		LastCheckAt: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(u.config.MetadataFile, data, defaultFilePerm)
}
