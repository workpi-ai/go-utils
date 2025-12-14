package ghrelease

import (
	"strings"
)

type AllFilesFilter struct{}

func (f *AllFilesFilter) ShouldExtract(filename string) bool {
	return true
}

type SimpleFilter struct {
	FileExt string
}

func (f *SimpleFilter) ShouldExtract(filename string) bool {
	return strings.HasSuffix(filename, f.FileExt)
}
