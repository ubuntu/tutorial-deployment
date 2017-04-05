package apis

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/paths"
)

const apiFileName = "codelab.json"

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
	return ioutil.WriteFile(path.Join(p.API, apiFileName), dat, 0644)
}

/*

// Codelab is the stringify representation of a codelab needed by the site
type Codelab struct {
	Source     string             `json:"source"`
	Title      string             `json:"title"`
	Summary    string             `json:"summary"`
	Category   []string           `json:"category"`
	Difficulty int                `json:"difficulty"`
	Duration   int                `json:"duration"`
	Tags       []string           `json:"tags"`
	Updated    *types.ContextTime `json:"updated"`
	URL        string             `json:"url"`
}

*/
