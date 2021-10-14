package fsync

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestLoadTargetDir(t *testing.T) {
	mt, err := LoadTargetDir("./testdir")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	if len(mt.files) != 2 {
		t.Fatalf("results are not equal, got: %v, want: 2", len(mt.files))
	}
}

// TestListenTarget covers a case of adding a new file to listen dir
func TestListenTargetOnAdd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mt, err := LoadTargetDir("./testdir")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	done := make(chan bool)
	errors := make(chan error, 1)
	go func() {
		if err := ListenTarget(ctx, mt); err != nil {
			errors <- err
		}
		done <- true
		errors <- nil
	}()

	// bring the change
	// create tmp file in testdir
	file, err := os.Create("./testdir/newfile.yml")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	// delete tmp file in testdir in teardown
	defer func() {
		_ = file.Close()
		_ = os.Remove("./testdir/newfile.yml")
	}()

	<-time.After(2 * time.Second)

	if len(mt.files) != 3 {
		// error must be called without Fatal which has os.Exit, so we trigger cancel flow
		t.Errorf("results are not equal, got: %d, want: 3", len(mt.files))
	}

	// stop ListTargetDir
	cancel()

	// wait for go-routine
	<-done
	if err := <-errors; err != nil {
		t.Errorf("want no error, got err: %s", err)
	}
}

// helper
func pp(list []*File) []File {
	newList := []File{}
	for _, file := range list {
		newList = append(newList, *file)
	}

	return newList
}

// TestListenTargetOnDelete
// plan is:
// ensure metadata is correct
// add tmp file to testdir
// get the new file deteced by updateTarget
// verify the change
// remove that file from testdir (simulate user)
// get deleted file detected by updateTarget
// verify the change
func TestUpdateTargetDir(t *testing.T) {
	// load metadata
	mt, err := LoadTargetDir("./testdir")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	// ensure metadata is correct
	f1 := File{Path: "./testdir", Root: true, Visited: true}
	f2 := File{Path: "testdir/testfile.yml", Visited: true}
	wantFiles := []*File{&f1, &f2}
	if !reflect.DeepEqual(mt.files, wantFiles) {
		t.Fatalf("results are not equal, got: %#v, want: %#v", pp(mt.files), pp(wantFiles))
	}

	// create tmp file in testdir (candidate for delete)
	file, err := os.Create("./testdir/tobedeleted.yml")
	if err != nil {
		t.Fatalf("failed to create tmp file in testdir, got: %s", err)
	}

	defer func() {
		_ = file.Close()
	}()

	// run UpdateTargetDir to detect a new file
	err = UpdateTargetDir(mt)
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	// verify new file is there
	f3 := File{Path: "testdir/tobedeleted.yml", Visited: true}
	wantFiles = []*File{&f1, &f2, &f3}
	if !reflect.DeepEqual(mt.files, wantFiles) {
		t.Fatalf("results are not equal, got: %v, want: %v", mt.files, wantFiles)
	}

	// delete (simulate user)
	err = os.Remove("./testdir/tobedeleted.yml")
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	// run UpdateTargetDir to detect the deletion
	err = UpdateTargetDir(mt)
	if err != nil {
		t.Fatalf("want no error, got err: %s", err)
	}

	// verify no tobedeted.yml in metedata
	wantFiles = []*File{&f1, &f2}
	if !reflect.DeepEqual(mt.files, wantFiles) {
		t.Fatalf("results are not equal, got: %#v, want: %#v", pp(mt.files), pp(wantFiles))
	}
}
