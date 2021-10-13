package fsync

import (
	"io/fs"
	"path/filepath"
)

/*
	Fsync does:
	1. parse target directory
	2. save loaded structure into metadata config
	3. listen a target directory for any changes
	4. update metadata config if new changes appeared

	Entities:
	- Metadata config is in-memory hash-map, keys are paths, values are dir structures
	- Target is a path (input param)
	- Listener is a blocking process
*/

// Metadata is keys list of files
type Metadata struct {
	files []string
}

func (m *Metadata) Add(path string) {
	m.files = append(m.files, path)
}

// LoadTargetDir parses Dir and returns list of all paths inside of Dir
func LoadTargetDir(target string) (*Metadata, error) {
	metadata := Metadata{}
	fn := func(path string, d fs.DirEntry, err error) error {
		metadata.Add(path)

		return nil
	}
	if err := filepath.WalkDir(target, fn); err != nil {
		return nil, err
	}

	return &metadata, nil
}
