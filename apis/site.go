package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/paths"
)

const apiFileName = "codelabs.json"

// site API main info
type site struct {
	Categories Categories        `json:"categories"`
	Codelabs   []codelab.Codelab `json:"codelabs"`
	Events     Events            `json:"events"`
}

// GenerateContent for website api, preparing and saving event images already
func GenerateContent(c []codelab.Codelab) ([]byte, error) {
	e, err := NewEvents()
	if err != nil {
		return nil, err
	}
	if err := e.SaveImages(); err != nil {
		return nil, err
	}
	cat, err := NewCategories()
	if err != nil {
		return nil, err
	}

	s := site{
		Categories: *cat,
		Codelabs:   c,
		Events:     *e,
	}
	return json.MarshalIndent(s, "", "  ")
}

// Save bytes on disk in API file
func Save(dat []byte) error {
	p := paths.New()
	// ensure that the API dir exists
	if err := os.MkdirAll(p.API, 0775); err != nil {
		return fmt.Errorf("couldn't create %s: %v", p.API, err)
	}
	return ioutil.WriteFile(path.Join(p.API, apiFileName), dat, 0644)
}
