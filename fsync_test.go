package fsync

import (
	"reflect"
	"testing"
)

func TestLoadTargetDir(t *testing.T) {
	files, err := LoadTargetDir("./testdir")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	wantFiles := []string{"./testdir", "testdir/testfile.yml"}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Fatalf("results are not equal, got: %v, want: %v", files, wantFiles)
	}
}
