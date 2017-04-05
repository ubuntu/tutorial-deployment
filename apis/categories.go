package apis

import (
	"fmt"
	"io/ioutil"
	"path"

	yaml "gopkg.in/yaml.v2"

	"github.com/ubuntu/tutorial-deployment/paths"
)

const (
	categoriesFilename = "categories.yaml"
)

// Categories are all supported category for codelabs
type Categories map[string]category

type category struct {
	Lightcolor     string `json:"lightcolor"`
	Maincolor      string `json:"maincolor"`
	Secondarycolor string `json:"secondarycolor"`
}

// NewCategories return all categories for main site
func NewCategories() (*Categories, error) {
	c := Categories{}
	p := paths.New()

	f := path.Join(p.MetaData, categoriesFilename)
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't read from %s: %v", f, err)
	}
	if err := yaml.Unmarshal(dat, &c); err != nil {
		return nil, fmt.Errorf("couldn't decode %s: %v", f, err)
	}

	return &c, nil
}
