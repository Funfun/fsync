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

// LoadTargetDir parses Dir and returns list of all paths inside of Dir
func LoadTargetDir(target string) ([]string, error) {
	files := []string{}
	fn := func(path string, d fs.DirEntry, err error) error {
		files = append(files, path)

		return nil
	}
	if err := filepath.WalkDir(target, fn); err != nil {
		return files, err
	}

	return files, nil
}
