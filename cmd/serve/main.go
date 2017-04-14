package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ubuntu/tutorial-deployment/apis"
	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/consts"
	"github.com/ubuntu/tutorial-deployment/internaltools"
	"github.com/ubuntu/tutorial-deployment/paths"
)

var codelabs []codelab.Codelab

func main() {
	flag.Usage = usage
	flag.Parse()
	args := internaltools.UniqueStrings(flag.Args())

	p := paths.New()
	if err := p.DetectPaths(); err != nil {
		log.Fatalf("Couldn't detect required paths: %s", err)
	}
	if err := p.ImportTutorialPaths(args); err != nil {
		log.Fatalf("Couldn't load tutorial paths: %s", err)
	}

	if err := p.CreateTempOutPath(); err != nil {
		log.Fatalf("Couldn't create temporary export paths: %s", err)
	}
	defer func() {
		if err := p.CleanTempPath(); err != nil {
			log.Printf("Couldn't clean temporary export directory: %v\n", err)
		}
	}()

	var err error
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	template := path.Join(p.MetaData, consts.TemplateFileName)

	type result struct {
		c   codelab.Codelab
		err error
	}
	ch := make(chan result)
	// export codelabs
	codelabRefs, err := codelab.Discover()
	if err != nil {
		log.Fatalf("Couldn't detect codelabs: %s", err)
	}
	if err := os.RemoveAll(p.Export); err != nil {
		log.Fatalf("Couldn't remove codelab export path %s: %v", p.Export, err)
	}
	for _, src := range codelabRefs {
		go func(ref string) {
			c, err := codelab.New(ref, p.Export, template, true)
			if err != nil {
				c = &codelab.Codelab{RefURI: ref}
			}
			ch <- result{*c, err}
		}(src)
	}

	hasError := false
	for _ = range codelabRefs {
		res := <-ch
		if res.err != nil {
			log.Printf("ERROR in %s: %v", res.c.RefURI, res.err)
			hasError = true
			continue
		}
		codelabs = append(codelabs, res.c)
	}
	if hasError {
		os.Exit(1)
	}

	if err := refreshAPIs(codelabs, p.API); err != nil {
		log.Fatalf("Couldn't refresh: %s", err)
	}

	// Install listeners and trigger refreshes
	if err := registerAllWatchers(); err != nil {
		log.Fatalf("Couldn't register watchers: %v", err)
	}
	wg := sync.WaitGroup{}
	stop := make(chan bool)
	listenForChanges(&wg, stop)

	wg.Wait()

}

func refreshAPIs(codelabs []codelab.Codelab, apiDir string) error {
	if err := os.RemoveAll(apiDir); err != nil {
		return fmt.Errorf("Couldn't remove API export path %s: %v", apiDir, err)
	}
	dat, err := apis.GenerateContent(codelabs)
	if err != nil {
		return fmt.Errorf("Couldn't generate API: %s", err)
	}
	if err != apis.Save(dat) {
		return fmt.Errorf("Couldn't save API: %s", err)
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: %s [options] [CodelabsDirOrFilesToWatch…]\n", os.Args[0], os.Args[0])
	fmt.Fprintf(os.Stderr, `Generate tutorials in html, using Polymerjs, to be dynamically served.

Those codelabs are generated on the fly in a temporary directory and in a non
destructive way. Any save on local source files (codelab markdown file or any
referenced local images) will retrigger the corresponding codelab build and
API generation, served by this local http webserver (default port is **8080**)

If the currently written codelabs are out of tree, they can be specified (files or
directories) directly on the command line.

Every default directories will be detected by the tool if present in the tutorial
directories. Arguments and options can tweak this behavior.

`)
	flag.PrintDefaults()
}
