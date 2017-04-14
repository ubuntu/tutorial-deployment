package main

import (
	"fmt"
	"log"
	"path"

	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/internaltools"
	"github.com/ubuntu/tutorial-deployment/paths"
)

// watchTrigger is a list of codelab to triggers when an event happen on a file
type watchTrigger []*codelab.Codelab

var (
	watcher *fsnotify.Watcher

	watchedTriggers map[string]watchTrigger
	watchedDirs     []string
)

// reisterAllWatchers needs to have all watchers cleaned before up
func registerAllWatchers() error {
	if watchedDirs != nil {
		log.Fatalf("Programer critical error. watchedDirs should be empty before calling registerAllWatchers. Got: %v", watchedDirs)
	}

	watchedTriggers = make(map[string]watchTrigger)
	for k := range codelabs {
		// get the reference, and not the copy for codelab
		c := &codelabs[k]
		for _, f := range c.FilesWatched {
			m := watchedTriggers[f]
			m = append(m, c)
			watchedTriggers[f] = m
			watchedDirs = append(watchedDirs, path.Dir(f))
		}
		fmt.Printf("DEBUG: codelab: %+v\n", c)
	}
	watchedDirs = internaltools.UniqueStrings(watchedDirs)
	return watchdirs()
}

func listenForChanges(wg *sync.WaitGroup, stop chan bool) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		p := paths.New()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create {
					if w, ok := watchedTriggers[event.Name]; ok {
						if err := unwatchdirs(); err != nil {
							log.Fatalf("Couldn't unwatch all dirs: %v", err)
						}
						for _, c := range w {
							if err := c.Refresh(); err != nil {
								log.Printf("Couldn't refresh successfully %s", c.RefURI)
							}
						}
						if err := refreshAPIs(codelabs, p.API); err != nil {
							log.Fatalf("Couldn't refresh: %s", err)
						}
						if err := registerAllWatchers(); err != nil {
							log.Fatalf("Couldn't watch dirs: %v", err)
						}
					}
				}

			case err := <-watcher.Errors:
				log.Println("Watch error:", err)
			case <-stop:
				return
			}
		}
	}()
}

func watchdirs() error {
	for _, dir := range watchedDirs {
		if err := watcher.Add(dir); err != nil {
			return err
		}
	}
	return nil
}

func unwatchdirs() error {
	for _, dir := range watchedDirs {
		if err := watcher.Remove(dir); err != nil {
			return err
		}
	}
	watchedDirs = nil
	return nil
}
