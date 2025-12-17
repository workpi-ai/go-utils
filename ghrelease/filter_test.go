package ghrelease

import "testing"

func TestKeepAllTransformer_Transform(t *testing.T) {
	transformer := &KeepAllTransformer{}

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "regular file",
			filename: "file.txt",
			want:     "file.txt",
		},
		{
			name:     "nested path",
			filename: "path/to/file.md",
			want:     "path/to/file.md",
		},
		{
			name:     "empty string",
			filename: "",
			want:     "",
		},
		{
			name:     "directory",
			filename: "path/to/dir/",
			want:     "path/to/dir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformer.Transform(tt.filename)
			if got != tt.want {
				t.Errorf("Transform(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExtTransformer_Transform(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		filename string
		want     string
	}{
		{
			name:     "matching extension",
			ext:      ".md",
			filename: "file.md",
			want:     "file.md",
		},
		{
			name:     "matching extension with path",
			ext:      ".md",
			filename: "path/to/file.md",
			want:     "path/to/file.md",
		},
		{
			name:     "non-matching extension",
			ext:      ".md",
			filename: "file.txt",
			want:     "",
		},
		{
			name:     "no extension",
			ext:      ".md",
			filename: "file",
			want:     "",
		},
		{
			name:     "empty filename",
			ext:      ".md",
			filename: "",
			want:     "",
		},
		{
			name:     "extension in middle",
			ext:      ".md",
			filename: "file.md.txt",
			want:     "",
		},
		{
			name:     "case sensitive",
			ext:      ".md",
			filename: "file.MD",
			want:     "",
		},
		{
			name:     "go extension",
			ext:      ".go",
			filename: "main.go",
			want:     "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := &ExtTransformer{Ext: tt.ext}
			got := transformer.Transform(tt.filename)
			if got != tt.want {
				t.Errorf("Transform(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestSubDirTransformer_Transform(t *testing.T) {
	tests := []struct {
		name     string
		subDir   string
		ext      string
		filename string
		want     string
	}{
		{
			name:     "matching subdir and ext",
			subDir:   "agents",
			ext:      ".md",
			filename: "agents/foo.md",
			want:     "foo.md",
		},
		{
			name:     "matching subdir nested path",
			subDir:   "commands",
			ext:      ".md",
			filename: "commands/code/review.md",
			want:     "code/review.md",
		},
		{
			name:     "matching subdir no ext filter",
			subDir:   "agents",
			ext:      "",
			filename: "agents/foo.txt",
			want:     "foo.txt",
		},
		{
			name:     "non-matching subdir",
			subDir:   "agents",
			ext:      ".md",
			filename: "commands/foo.md",
			want:     "",
		},
		{
			name:     "non-matching ext",
			subDir:   "agents",
			ext:      ".md",
			filename: "agents/foo.txt",
			want:     "",
		},
		{
			name:     "subdir only no trailing slash",
			subDir:   "agents",
			ext:      ".md",
			filename: "agents",
			want:     "",
		},
		{
			name:     "empty filename",
			subDir:   "agents",
			ext:      ".md",
			filename: "",
			want:     "",
		},
		{
			name:     "partial subdir match",
			subDir:   "agents",
			ext:      ".md",
			filename: "agents-new/foo.md",
			want:     "",
		},
		{
			name:     "subdir as prefix of filename",
			subDir:   "agent",
			ext:      ".md",
			filename: "agents/foo.md",
			want:     "",
		},
		{
			name:     "deeply nested",
			subDir:   "commands",
			ext:      ".md",
			filename: "commands/a/b/c/d.md",
			want:     "a/b/c/d.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := &SubDirTransformer{SubDir: tt.subDir, Ext: tt.ext}
			got := transformer.Transform(tt.filename)
			if got != tt.want {
				t.Errorf("Transform(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
