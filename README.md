# Go Utils

Common utilities for Go projects.

## Packages

### ghrelease

GitHub Release updater for downloading and extracting release assets.

#### Features

- Download latest release from GitHub
- Extract files with custom filtering
- Version management with local caching
- Configurable timeouts

#### Usage

```go
package main

import (
    "github.com/workpi-ai/go-utils/ghrelease"
)

func main() {
    updater := ghrelease.NewUpdater(ghrelease.UpdaterConfig{
        RepoOwner:        "owner",
        RepoName:         "repo",
        DestDir:          "/path/to/dest",
        MetadataFilename: "metadata.json",
        ExtractFilter:    &ghrelease.SimpleFilter{
            FileExt: ".md",
        },
    })
    
    if err := updater.Update(); err != nil {
        panic(err)
    }
}
```

#### Built-in Filters

**AllFilesFilter** - Extract all files:
```go
ExtractFilter: &ghrelease.AllFilesFilter{}
```

**SimpleFilter** - Extract files by extension:
```go
ExtractFilter: &ghrelease.SimpleFilter{
    FileExt: ".json",
}
```

#### Custom Filter

Implement `ExtractFilter` interface for custom extraction logic:

```go
type ExtractFilter interface {
    ShouldExtract(filename string) bool
}
```

## Installation

```bash
go get github.com/workpi-ai/go-utils
```

## License

MIT
