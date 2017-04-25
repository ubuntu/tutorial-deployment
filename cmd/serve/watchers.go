package main

import (
	"log"
	"path"

	"sync"

	"fmt"

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
		log.Printf("DEBUG: codelab: %+v\n", c)
	}
	watchedDirs = internaltools.UniqueStrings(watchedDirs)
	return watchdirs()
}

func listenForChanges(wg *sync.WaitGroup, stop <-chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer watcher.Close()
		p := paths.New()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create {
					cs := impactedCodelabs(event.Name)
					if err := refreshCodelabs(cs, *p); err != nil {
						log.Print(err)
						// do not send refresh signal
						continue
					}
					for _, c := range cs {
						hub.Send([]byte(c.URL))
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

func refreshCodelabs(cs []*codelab.Codelab, p paths.Path) error {
	defer func() {
		if err := registerAllWatchers(); err != nil {
			log.Printf("Couldn't watch dirs: %v", err)
		}
	}()
	if err := unwatchdirs(); err != nil {
		return fmt.Errorf("Couldn't unwatch all dirs: %v", err)
	}
	for _, c := range cs {
		if err := c.Refresh(); err != nil {
			return fmt.Errorf("Couldn't refresh successfully %s", c.RefURI)
		}
	}
	if err := refreshAPIs(codelabs, p.API); err != nil {
		return fmt.Errorf("Couldn't refresh: %s", err)
	}

	return nil
}

func impactedCodelabs(file string) []*codelab.Codelab {
	w, ok := watchedTriggers[file]
	if !ok {
		return nil
	}
	return []*codelab.Codelab(w)
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
