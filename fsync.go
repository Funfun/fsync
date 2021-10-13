package fsync

import (
	"context"
	"io/fs"
	"path/filepath"
	"time"
)

/*
	Fsync does:
	1. parse target directory
	2. save loaded structure into metadata config
	3. listen a target directory for any changes
	4. update metadata config if new changes have appeared

	Entities:
	- Metadata config is in-memory hash-map, keys are paths, values are dir structures
	- Target is a path (input param)
	- Listener is a blocking process
*/

const listenInterval = 1000 * time.Millisecond // TODO: extract to config

// Metadata is keys list of files.
type Metadata struct {
	files  []string
	target string
}

// Add appends path to files list.
func (m *Metadata) Add(path string) {
	m.files = append(m.files, path)
}

// Includes scans for presence of path in files list.
func (m *Metadata) Included(path string) bool {
	for _, file := range m.files {
		if file == path {
			return true
		}
	}

	return false
}

// LoadTargetDir parses Dir and returns list of all paths inside of Dir.
func LoadTargetDir(target string) (*Metadata, error) {
	metadata := Metadata{target: target}
	fn := func(path string, d fs.DirEntry, err error) error {
		metadata.Add(path)

		return nil
	}
	if err := filepath.WalkDir(metadata.target, fn); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// ListenTarget checks every N seconds the target directory
func ListenTarget(ctx context.Context, metadata *Metadata) error {
	checker := time.NewTicker(listenInterval)

	// checkFn looks for path in metadata
	// if path is not present, then add to Metadata
	checkFn := func(path string, d fs.DirEntry, err error) error {
		if !metadata.Included(path) {
			metadata.Add(path)
		}

		return nil
	}

	for {
		select {
		case <-checker.C:
			// it should handle errors & propogate it to upstream
			if err := filepath.WalkDir(metadata.target, checkFn); err != nil {
				return err
			}

			// TODO: after walking is done, time to delete all old directries and files
		case <-ctx.Done():
			return nil
		}
	}
}
