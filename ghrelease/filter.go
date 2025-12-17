package ghrelease

import (
	"strings"
)

type KeepAllTransformer struct{}

func (t *KeepAllTransformer) Transform(filename string) string {
	return filename
}

type ExtTransformer struct {
	Ext string
}

func (t *ExtTransformer) Transform(filename string) string {
	if strings.HasSuffix(filename, t.Ext) {
		return filename
	}
	return ""
}

type SubDirTransformer struct {
	SubDir string
	Ext    string
}

func (t *SubDirTransformer) Transform(filename string) string {
	prefix := t.SubDir + "/"
	if !strings.HasPrefix(filename, prefix) {
		return ""
	}
	path := strings.TrimPrefix(filename, prefix)
	if t.Ext != "" && !strings.HasSuffix(path, t.Ext) {
		return ""
	}
	return path
}
