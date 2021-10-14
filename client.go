package fsync

import (
	"context"
	"io/fs"
	"log"
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
	- TODO: Metadata config is in-memory hash-map, keys are paths, values are dir structures
	- Metadata is a simple fix array for time being
	- Target is a path (input param)
	- Listener is a blocking process
*/

const listenInterval = 1000 * time.Millisecond // TODO: extract to config

type File struct {
	Root    bool
	Path    string
	Visited bool
}

// Metadata is keys list of files.
type Metadata struct {
	files  []*File
	target string
}

// Add appends path to files list.
func (m *Metadata) Add(path string) {
	m.files = append(m.files, &File{Path: path, Root: m.target == path, Visited: true})
}

// Includes scans for presence of path in files list.
func (m *Metadata) Includes(path string) bool {
	for _, file := range m.files {
		if file.Path == path {
			return true
		}
	}

	return false
}

func (m *Metadata) Visited(path string) {
	// find and visit with path
	for _, file := range m.files {
		if file.Path == path {
			file.Visited = true
		}
	}
}

func (m *Metadata) RemoveUntouched() error {
	for i, file := range m.files {
		if file.Root {
			continue
		}

		if !file.Visited {
			// delete from metadata
			m.files = append(m.files[:i], m.files[i+1:]...)
		}
	}

	return nil
}

func (m *Metadata) Reset() {
	for _, file := range m.files {
		file.Visited = false
	}
}

func (m *Metadata) Inspect() {
	log.Println("DEBUG: Files being watch at moment:")
	for _, file := range m.files {
		log.Printf("DEBUG: path: %s\n, root: %v", file.Path, file.Root)
	}
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

	for {
		select {
		case <-checker.C:
			// go and update metadata whenever there is a change in target dir
			// TODO: after walking is done, time to delete all old directries and files
			if err := UpdateTargetDir(metadata); err != nil {
				return err
			}

			// print current state (debugging)
			metadata.Inspect()
		case <-ctx.Done():
			return nil
		}
	}
}

func UpdateTargetDir(metadata *Metadata) error {
	// reset all visit states
	metadata.Reset()

	// checkFn looks for path in metadata
	// if path is not present, then add to Metadata
	checkFn := func(path string, d fs.DirEntry, err error) error {
		if metadata.Includes(path) {
			metadata.Visited(path)
		} else {
			metadata.Add(path)
		}

		return nil
	}

	// it should handle errors & propogate it to upstream
	err := filepath.WalkDir(metadata.target, checkFn)
	if err != nil {
		return err
	}

	// after walk, check metadata for unvisited files and delete them
	return metadata.RemoveUntouched()
}
