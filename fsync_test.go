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

	wantFiles := []string{"./testdir", "testdir/testfile.yml"}
	if !reflect.DeepEqual(mt.files, wantFiles) {
		t.Fatalf("results are not equal, got: %v, want: %v", mt.files, wantFiles)
	}
}

func TestListenTarget(t *testing.T) {
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

	// verify the change
	wantFiles := []string{"./testdir", "testdir/testfile.yml", "testdir/newfile.yml"}
	if !reflect.DeepEqual(mt.files, wantFiles) {
		// error must be called without Fatal which has os.Exit, so we trigger cancel flow
		t.Errorf("results are not equal, got: %v, want: %v", mt.files, wantFiles)
	}

	// stop ListTargetDir
	cancel()

	// wait for go-routine
	<-done
	if err := <-errors; err != nil {
		t.Errorf("want no error, got err: %s", err)
	}
}
