package apis

import (
	"encoding/json"

	"github.com/ubuntu/tutorial-deployment/codelab"
)

// site API main info
type site struct {
	Categories Categories        `json:"categories"`
	Codelabs   []codelab.Codelab `json:"codelabs"`
	Events     Events            `json:"events"`
}

// GenerateAPIcontent for website api, preparing and saving event images already
func GenerateAPIcontent(c []codelab.Codelab) ([]byte, error) {
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
