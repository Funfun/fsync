package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Funfun/fsync"
)

var directoryToWatch = flag.String("targetDirectory", "./", "Specify directiry to watch for changes")

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()

	log.Printf("loading %s\n", *directoryToWatch)
	mt, err := fsync.LoadTargetDir(*directoryToWatch)
	if err != nil {
		log.Fatalf("failed to load %s directory, got: %s", *directoryToWatch, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// prepare work forces and start watch dir
	go func() {
		log.Println("watching for changes")
		if err := fsync.ListenTarget(ctx, mt); err != nil {
			log.Println(err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	s := <-signalCh
	fmt.Println("Got signal:", s)
}
