package codelab

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"path"

	"github.com/ubuntu/tutorial-deployment/internaltools"
	"github.com/ubuntu/tutorial-deployment/paths"
)

const (
	gdocFileName = "gdoc.def"
	gdocPrefix   = "gdoc:"
)

// Discover existing codelabs in the import path
func Discover() (codelabs []string, err error) {
	p := paths.New()
	for _, fpath := range p.TutorialInputs {
		fi, err := os.Stat(fpath)
		if err != nil {
			return nil, fmt.Errorf("Couldn't stat: %s", err)
		}
		if fi.IsDir() {
			err = filepath.Walk(fpath, func(fpath string, i os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !i.IsDir() {
					newCodelabs, err := getCodelabReference(fpath)
					if err != nil {
						return err
					}
					codelabs = append(codelabs, newCodelabs...)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			newCodelabs, err := getCodelabReference(fpath)
			if err != nil {
				return nil, err
			}
			codelabs = append(codelabs, newCodelabs...)
		}
	}
	codelabs = internaltools.UniqueStrings(codelabs)

	return codelabs, nil
}

// return a list of file paths itself or gdoc:<ID> prefix. The file path will be a list of one element for markdown
// files and the number of gdoc elements in the gdoc file
// we ignore any file starting with _ and not ending up with .md
// nor being gdoc.def files (handling google doc definition files)
// those could be images or other assets.
func getCodelabReference(p string) (r []string, err error) {
	if strings.HasPrefix(path.Base(p), "_") {
		return nil, nil
	}

	if strings.HasSuffix(p, ".md") {
		return []string{p}, nil
	}
	if path.Base(p) != gdocFileName {
		return nil, nil
	}

	// open the gdoc definition file and pass references back
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open %s: %v", p, err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.Trim(s.Text(), " ")
		if strings.HasPrefix(l, "#") || len(l) == 0 {
			continue
		}
		r = append(r, fmt.Sprintf("%s%s", gdocPrefix, l))
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("Couldn't read %s: %v", p, err)
	}

	return r, nil
}
