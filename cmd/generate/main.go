package main

import (
	"flag"
	"log"
	"path"

	"os"

	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/consts"
	"github.com/ubuntu/tutorial-deployment/internaltools"
	"github.com/ubuntu/tutorial-deployment/paths"

	// allow parsers to register themselves
	_ "github.com/didrocks/codelab-ubuntu-tools/claat/parser/gdoc"
	_ "github.com/didrocks/codelab-ubuntu-tools/claat/parser/md"
)

func main() {
	flag.Parse()
	args := internaltools.UniqueStrings(flag.Args())

	p := paths.New()
	if err := p.DetectPaths(); err != nil {
		log.Fatalf("Couldn't detect required paths: %s", err)
	}
	if err := p.ImportTutorialPaths(args); err != nil {
		log.Fatalf("Couldn't load tutorial paths: %s", err)
	}

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
	for _, src := range codelabRefs {
		go func(ref string) {
			c, err := codelab.New(ref, p.Export, template, false)
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
	}
	if hasError {
		os.Exit(1)
	}

}
