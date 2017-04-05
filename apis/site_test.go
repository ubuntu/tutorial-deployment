package apis

import (
	"flag"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"bytes"

	"github.com/didrocks/codelab-ubuntu-tools/claat/types"
	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/paths"
)

var update = flag.Bool("update", false, "update generated files")

func TestGenerateAPIcontent(t *testing.T) {
	published := types.LegacyStatus([]string{"Published"})
	exCodelabs := []codelab.Codelab{
		codelab.Codelab{RefURI: "REFPATH1", Codelab: types.Codelab{Meta: types.Meta{ID: "123", Title: "A title", Status: &published, Summary: "Awesome tutorial", URL: "https://tutorial1.com", Difficulty: 3, Categories: []string{"category1", "category2"}, Tags: []string{"foo", "bar"}, Duration: 60, Feedback: "http://feedback.com"}}, Updated: stringToContextTime(t, "1983-09-13"), FilesWatched: []string{"onefile", "twofiles"}},
		codelab.Codelab{},
		codelab.Codelab{RefURI: "REFPATH2", Codelab: types.Codelab{Meta: types.Meta{ID: "456"}}, Updated: stringToContextTime(t, "2017-04-05")},
	}
	testCases := []struct {
		metaDir  string
		codelabs []codelab.Codelab

		wantReturnPath string
		wantErr        bool
	}{
		{"testdata/sites/valid", exCodelabs, "testdata/sites/valid/valid-api-output.json", false},
		{"testdata/sites/categories-missing", exCodelabs, "", true},
		{"testdata/sites/events-missing", exCodelabs, "", true},
		{"testdata/sites/valid", []codelab.Codelab{}, "testdata/sites/valid/valid-without-codelab.json", false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("generate api with codelab: %+v, metadata: %s", tc.codelabs, tc.metaDir), func(t *testing.T) {
			// Setup/Teardown
			p, teardown := paths.MockPath()
			defer teardown()
			p.MetaData = tc.metaDir

			// Test
			dat, err := GenerateAPIcontent(tc.codelabs)

			if (err != nil) != tc.wantErr {
				t.Errorf("GenerateAPIcontent() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if *update {
				if err := ioutil.WriteFile(tc.wantReturnPath, dat, 0644); err != nil {
					t.Fatalf("failed updating %s: %v", tc.wantReturnPath, err)
				}
			}

			wanted, err := ioutil.ReadFile(tc.wantReturnPath)
			if err != nil {
				t.Fatalf("couldn't read %s: %v", tc.wantReturnPath, err)
			}

			if !bytes.Equal(dat, wanted) {
				t.Errorf("generate api: got %s; want %s", dat, wanted)
			}
		})
	}
}

func stringToContextTime(t *testing.T, date string) types.ContextTime {
	tt, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatalf("couldn't convert time from %s", date)
	}
	return types.ContextTime(tt)
}
