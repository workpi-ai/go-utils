package ghrelease

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var defaultTarget = ExtractTarget{PathTransformer: &KeepAllTransformer{}, DestDir: "/tmp"}

func mustNewUpdater(t *testing.T, config UpdaterConfig) *Updater {
	t.Helper()
	if config.RepoOwner == "" {
		config.RepoOwner = "owner"
	}
	if config.RepoName == "" {
		config.RepoName = "repo"
	}
	if len(config.Targets) == 0 {
		config.Targets = []ExtractTarget{defaultTarget}
	}
	updater, err := NewUpdater(config)
	if err != nil {
		t.Fatalf("NewUpdater() error = %v", err)
	}
	return updater
}

func TestNewUpdater(t *testing.T) {
	tests := []struct {
		name    string
		config  UpdaterConfig
		want    UpdaterConfig
		wantErr bool
	}{
		{
			name: "all fields provided",
			config: UpdaterConfig{
				RepoOwner:       "owner",
				RepoName:        "repo",
				MetadataFile:    "/tmp/metadata.json",
				RequestTimeout:  5 * time.Second,
				DownloadTimeout: 60 * time.Second,
				Targets: []ExtractTarget{
					{PathTransformer: &KeepAllTransformer{}, DestDir: "/tmp"},
				},
			},
			want: UpdaterConfig{
				RepoOwner:       "owner",
				RepoName:        "repo",
				MetadataFile:    "/tmp/metadata.json",
				RequestTimeout:  5 * time.Second,
				DownloadTimeout: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "use default timeouts",
			config: UpdaterConfig{
				RepoOwner:    "owner",
				RepoName:     "repo",
				MetadataFile: "/tmp/metadata.json",
				Targets: []ExtractTarget{
					{PathTransformer: &KeepAllTransformer{}, DestDir: "/tmp"},
				},
			},
			want: UpdaterConfig{
				RepoOwner:       "owner",
				RepoName:        "repo",
				MetadataFile:    "/tmp/metadata.json",
				RequestTimeout:  defaultRequestTimeout,
				DownloadTimeout: defaultDownloadTimeout,
			},
			wantErr: false,
		},
		{
			name: "empty targets",
			config: UpdaterConfig{
				RepoOwner:    "owner",
				RepoName:     "repo",
				MetadataFile: "/tmp/metadata.json",
				Targets:      []ExtractTarget{},
			},
			wantErr: true,
		},
		{
			name: "nil targets",
			config: UpdaterConfig{
				RepoOwner:    "owner",
				RepoName:     "repo",
				MetadataFile: "/tmp/metadata.json",
			},
			wantErr: true,
		},
		{
			name: "empty repo owner",
			config: UpdaterConfig{
				RepoOwner: "",
				RepoName:  "repo",
				Targets:   []ExtractTarget{{PathTransformer: &KeepAllTransformer{}, DestDir: "/tmp"}},
			},
			wantErr: true,
		},
		{
			name: "empty repo name",
			config: UpdaterConfig{
				RepoOwner: "owner",
				RepoName:  "",
				Targets:   []ExtractTarget{{PathTransformer: &KeepAllTransformer{}, DestDir: "/tmp"}},
			},
			wantErr: true,
		},
		{
			name: "nil path transformer",
			config: UpdaterConfig{
				RepoOwner: "owner",
				RepoName:  "repo",
				Targets:   []ExtractTarget{{PathTransformer: nil, DestDir: "/tmp"}},
			},
			wantErr: true,
		},
		{
			name: "empty dest dir",
			config: UpdaterConfig{
				RepoOwner: "owner",
				RepoName:  "repo",
				Targets:   []ExtractTarget{{PathTransformer: &KeepAllTransformer{}, DestDir: ""}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater, err := NewUpdater(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewUpdater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if updater != nil {
					t.Error("NewUpdater() should return nil updater on error")
				}
				return
			}

			if updater == nil {
				t.Fatal("NewUpdater returned nil")
			}
			if updater.config.RepoOwner != tt.want.RepoOwner {
				t.Errorf("RepoOwner = %v, want %v", updater.config.RepoOwner, tt.want.RepoOwner)
			}
			if updater.config.RepoName != tt.want.RepoName {
				t.Errorf("RepoName = %v, want %v", updater.config.RepoName, tt.want.RepoName)
			}
			if updater.config.MetadataFile != tt.want.MetadataFile {
				t.Errorf("MetadataFile = %v, want %v", updater.config.MetadataFile, tt.want.MetadataFile)
			}
			if updater.config.RequestTimeout != tt.want.RequestTimeout {
				t.Errorf("RequestTimeout = %v, want %v", updater.config.RequestTimeout, tt.want.RequestTimeout)
			}
			if updater.config.DownloadTimeout != tt.want.DownloadTimeout {
				t.Errorf("DownloadTimeout = %v, want %v", updater.config.DownloadTimeout, tt.want.DownloadTimeout)
			}
			if updater.client == nil {
				t.Error("client is nil")
			}
		})
	}
}

func TestUpdater_getLocalVersion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		metadataFile string
		setup        func()
		want         string
	}{
		{
			name:         "file not exists",
			metadataFile: filepath.Join(tmpDir, "not_exists.json"),
			setup:        func() {},
			want:         "",
		},
		{
			name:         "valid metadata",
			metadataFile: filepath.Join(tmpDir, "valid.json"),
			setup: func() {
				data, _ := json.Marshal(Metadata{Version: "v1.0.0", LastCheckAt: "2024-01-01T00:00:00Z"})
				os.WriteFile(filepath.Join(tmpDir, "valid.json"), data, 0644)
			},
			want: "v1.0.0",
		},
		{
			name:         "invalid json",
			metadataFile: filepath.Join(tmpDir, "invalid.json"),
			setup: func() {
				os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("not json"), 0644)
			},
			want: "",
		},
		{
			name:         "empty version",
			metadataFile: filepath.Join(tmpDir, "empty.json"),
			setup: func() {
				data, _ := json.Marshal(Metadata{Version: "", LastCheckAt: "2024-01-01T00:00:00Z"})
				os.WriteFile(filepath.Join(tmpDir, "empty.json"), data, 0644)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			updater := mustNewUpdater(t, UpdaterConfig{MetadataFile: tt.metadataFile})
			got := updater.getLocalVersion()
			if got != tt.want {
				t.Errorf("getLocalVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdater_saveLocalVersion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		metadataFile string
		version      string
		wantErr      bool
	}{
		{
			name:         "save to new file",
			metadataFile: filepath.Join(tmpDir, "new.json"),
			version:      "v1.0.0",
			wantErr:      false,
		},
		{
			name:         "save to nested dir",
			metadataFile: filepath.Join(tmpDir, "nested", "dir", "metadata.json"),
			version:      "v2.0.0",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := mustNewUpdater(t, UpdaterConfig{MetadataFile: tt.metadataFile})
			err := updater.saveLocalVersion(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("saveLocalVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				data, err := os.ReadFile(tt.metadataFile)
				if err != nil {
					t.Fatalf("failed to read metadata file: %v", err)
				}

				var m Metadata
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("failed to unmarshal metadata: %v", err)
				}

				if m.Version != tt.version {
					t.Errorf("saved version = %v, want %v", m.Version, tt.version)
				}

				if m.LastCheckAt == "" {
					t.Error("LastCheckAt is empty")
				}
			}
		})
	}
}

func TestUpdater_stripRootDir(t *testing.T) {
	updater := mustNewUpdater(t, UpdaterConfig{})

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "with root dir",
			filename: "repo-v1.0.0/agents/foo.md",
			want:     "agents/foo.md",
		},
		{
			name:     "nested path",
			filename: "repo-v1.0.0/commands/code/review.md",
			want:     "commands/code/review.md",
		},
		{
			name:     "no slash",
			filename: "file.txt",
			want:     "",
		},
		{
			name:     "only root dir",
			filename: "repo-v1.0.0/",
			want:     "",
		},
		{
			name:     "empty string",
			filename: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updater.stripRootDir(tt.filename)
			if got != tt.want {
				t.Errorf("stripRootDir(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestUpdater_extractZip(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	commandsDir := filepath.Join(tmpDir, "commands")

	zipData := createTestZip(t, map[string]string{
		"repo-v1.0.0/agents/foo.md":          "agent foo content",
		"repo-v1.0.0/agents/bar.md":          "agent bar content",
		"repo-v1.0.0/commands/code/review.md": "review content",
		"repo-v1.0.0/README.md":              "readme content",
		"repo-v1.0.0/other/file.txt":         "other content",
	})

	updater := mustNewUpdater(t, UpdaterConfig{
		Targets: []ExtractTarget{
			{
				PathTransformer: &SubDirTransformer{SubDir: "agents", Ext: ".md"},
				DestDir:         agentsDir,
			},
			{
				PathTransformer: &SubDirTransformer{SubDir: "commands", Ext: ".md"},
				DestDir:         commandsDir,
			},
		},
	})

	err := updater.extractZip(bytes.NewReader(zipData))
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	expectedFiles := map[string]string{
		filepath.Join(agentsDir, "foo.md"):            "agent foo content",
		filepath.Join(agentsDir, "bar.md"):            "agent bar content",
		filepath.Join(commandsDir, "code/review.md"): "review content",
	}

	for path, expectedContent := range expectedFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read %s: %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("content of %s = %q, want %q", path, string(content), expectedContent)
		}
	}

	notExpectedFiles := []string{
		filepath.Join(tmpDir, "README.md"),
		filepath.Join(tmpDir, "other/file.txt"),
	}

	for _, path := range notExpectedFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should not exist", path)
		}
	}
}

func TestUpdater_extractZip_emptyTargets(t *testing.T) {
	_, err := NewUpdater(UpdaterConfig{
		Targets: []ExtractTarget{},
	})

	if err == nil {
		t.Fatal("NewUpdater() should return error for empty targets")
	}
}

func TestUpdater_extractZip_invalidZip(t *testing.T) {
	updater := mustNewUpdater(t, UpdaterConfig{})

	err := updater.extractZip(bytes.NewReader([]byte("not a zip file")))
	if err == nil {
		t.Error("extractZip() should return error for invalid zip")
	}
}

func TestUpdater_extractZip_withKeepAllTransformer(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	zipData := createTestZip(t, map[string]string{
		"repo-v1.0.0/file1.txt": "content1",
		"repo-v1.0.0/file2.md":  "content2",
		"repo-v1.0.0/dir/file3.go": "content3",
	})

	updater := mustNewUpdater(t, UpdaterConfig{
		Targets: []ExtractTarget{
			{
				PathTransformer: &KeepAllTransformer{},
				DestDir:         destDir,
			},
		},
	})

	err := updater.extractZip(bytes.NewReader(zipData))
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	expectedFiles := []string{
		filepath.Join(destDir, "file1.txt"),
		filepath.Join(destDir, "file2.md"),
		filepath.Join(destDir, "dir/file3.go"),
	}

	for _, path := range expectedFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}
}

func TestUpdater_extractZip_withExtTransformer(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	zipData := createTestZip(t, map[string]string{
		"repo-v1.0.0/file1.txt": "content1",
		"repo-v1.0.0/file2.md":  "content2",
		"repo-v1.0.0/file3.md":  "content3",
	})

	updater := mustNewUpdater(t, UpdaterConfig{
		Targets: []ExtractTarget{
			{
				PathTransformer: &ExtTransformer{Ext: ".md"},
				DestDir:         destDir,
			},
		},
	})

	err := updater.extractZip(bytes.NewReader(zipData))
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	expectedFiles := []string{
		filepath.Join(destDir, "file2.md"),
		filepath.Join(destDir, "file3.md"),
	}

	for _, path := range expectedFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}

	notExpectedFiles := []string{
		filepath.Join(destDir, "file1.txt"),
	}

	for _, path := range notExpectedFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should not exist", path)
		}
	}
}

func TestUpdater_extractZip_multipleTargetsSameFile(t *testing.T) {
	tmpDir := t.TempDir()
	destDir1 := filepath.Join(tmpDir, "dest1")
	destDir2 := filepath.Join(tmpDir, "dest2")

	zipData := createTestZip(t, map[string]string{
		"repo-v1.0.0/shared/file.md": "shared content",
	})

	updater := mustNewUpdater(t, UpdaterConfig{
		Targets: []ExtractTarget{
			{
				PathTransformer: &SubDirTransformer{SubDir: "shared", Ext: ".md"},
				DestDir:         destDir1,
			},
			{
				PathTransformer: &SubDirTransformer{SubDir: "shared", Ext: ".md"},
				DestDir:         destDir2,
			},
		},
	})

	err := updater.extractZip(bytes.NewReader(zipData))
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	for _, dir := range []string{destDir1, destDir2} {
		path := filepath.Join(dir, "file.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read %s: %v", path, err)
			continue
		}
		if string(content) != "shared content" {
			t.Errorf("content of %s = %q, want %q", path, string(content), "shared content")
		}
	}
}

func TestUpdater_extractZip_directoryEntries(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	_, err := w.Create("repo-v1.0.0/")
	if err != nil {
		t.Fatalf("failed to create dir entry: %v", err)
	}

	_, err = w.Create("repo-v1.0.0/subdir/")
	if err != nil {
		t.Fatalf("failed to create subdir entry: %v", err)
	}

	f, err := w.Create("repo-v1.0.0/file.txt")
	if err != nil {
		t.Fatalf("failed to create file entry: %v", err)
	}
	f.Write([]byte("content"))

	w.Close()

	updater := mustNewUpdater(t, UpdaterConfig{
		Targets: []ExtractTarget{
			{
				PathTransformer: &KeepAllTransformer{},
				DestDir:         destDir,
			},
		},
	})

	err = updater.extractZip(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	filePath := filepath.Join(destDir, "file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", filePath)
	}
}

func TestUpdater_extractFile(t *testing.T) {
	tmpDir := t.TempDir()

	zipData := createTestZip(t, map[string]string{
		"test.txt": "test content",
	})

	zipReader, _ := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	file := zipReader.File[0]

	updater := mustNewUpdater(t, UpdaterConfig{})

	destPath := filepath.Join(tmpDir, "nested", "dir", "output.txt")
	err := updater.extractFile(file, destPath)
	if err != nil {
		t.Fatalf("extractFile() error = %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("content = %q, want %q", string(content), "test content")
	}
}

func TestUpdater_extractFile_overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	destPath := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(destPath, []byte("old content"), 0644)

	zipData := createTestZip(t, map[string]string{
		"test.txt": "new content",
	})

	zipReader, _ := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	file := zipReader.File[0]

	updater := mustNewUpdater(t, UpdaterConfig{})

	err := updater.extractFile(file, destPath)
	if err != nil {
		t.Fatalf("extractFile() error = %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(content) != "new content" {
		t.Errorf("content = %q, want %q", string(content), "new content")
	}
}

func TestMetadata(t *testing.T) {
	m := Metadata{
		Version:     "v1.0.0",
		LastCheckAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m2 Metadata
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if m2.Version != m.Version {
		t.Errorf("Version = %v, want %v", m2.Version, m.Version)
	}
	if m2.LastCheckAt != m.LastCheckAt {
		t.Errorf("LastCheckAt = %v, want %v", m2.LastCheckAt, m.LastCheckAt)
	}
}

func TestUpdaterConfig(t *testing.T) {
	config := UpdaterConfig{
		RepoOwner:       "owner",
		RepoName:        "repo",
		MetadataFile:    "/path/to/metadata.json",
		RequestTimeout:  5 * time.Second,
		DownloadTimeout: 60 * time.Second,
		Targets: []ExtractTarget{
			{
				PathTransformer: &KeepAllTransformer{},
				DestDir:         "/dest",
			},
		},
	}

	if config.RepoOwner != "owner" {
		t.Errorf("RepoOwner = %v, want %v", config.RepoOwner, "owner")
	}
	if config.RepoName != "repo" {
		t.Errorf("RepoName = %v, want %v", config.RepoName, "repo")
	}
	if len(config.Targets) != 1 {
		t.Errorf("Targets length = %v, want %v", len(config.Targets), 1)
	}
}

func TestExtractTarget(t *testing.T) {
	target := ExtractTarget{
		PathTransformer: &SubDirTransformer{SubDir: "agents", Ext: ".md"},
		DestDir:         "/dest/agents",
	}

	if target.DestDir != "/dest/agents" {
		t.Errorf("DestDir = %v, want %v", target.DestDir, "/dest/agents")
	}

	result := target.PathTransformer.Transform("agents/foo.md")
	if result != "foo.md" {
		t.Errorf("Transform result = %v, want %v", result, "foo.md")
	}
}

func createTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write zip entry %s: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}
