package apis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"

	"github.com/ubuntu/tutorial-deployment/consts"
	"github.com/ubuntu/tutorial-deployment/paths"
)

const (
	eventFilename = "events.yaml"
	assetsDir     = "assets"
)

// Events are all events planned and grouping some codelabs
type Events map[string]event

type event struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
}

// NewEvents return all events for main site
func NewEvents() (*Events, error) {
	e := Events{}
	p := paths.New()

	f := path.Join(p.MetaData, eventFilename)
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't read from %s: %v", f, err)
	}
	if err := yaml.Unmarshal(dat, &e); err != nil {
		return nil, fmt.Errorf("couldn't decode %s: %v", f, err)
	}

	return &e, nil
}

// SaveImages redirect and moves them to api directory
func (evs *Events) SaveImages() error {
	p := paths.New()
	destdir := path.Join(p.API, assetsDir)
	if err := os.MkdirAll(destdir, 0775); err != nil {
		return fmt.Errorf("couldn't create %s: %v", destdir, err)
	}
	for k, e := range *evs {
		// path is relative to metadata directory (where the events file is located)
		src := path.Join(p.MetaData, e.Logo)
		dest := path.Join(assetsDir, path.Base(e.Logo))
		e.Logo = path.Join(consts.APIURL, dest)

		data, err := ioutil.ReadFile(src)
		if err != nil {
			return fmt.Errorf("%s doesn't exist: %v", src, err)
		}

		if err := ioutil.WriteFile(path.Join(p.API, dest), data, 0644); err != nil {
			return fmt.Errorf("couldn't create %s: %v", path.Join(p.API, dest), err)
		}

		(*evs)[k] = e
	}
	return nil
}
