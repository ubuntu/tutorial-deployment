package apis

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/didrocks/codelab-ubuntu-tools/claat/types"
	"github.com/ubuntu/tutorial-deployment/codelab"
	"github.com/ubuntu/tutorial-deployment/paths"
	"github.com/ubuntu/tutorial-deployment/testtools"
)

var update = flag.Bool("update", false, "update generated files")

func TestGenerateContent(t *testing.T) {
	published := types.LegacyStatus([]string{"Published"})
	exCodelabs := []codelab.Codelab{
		codelab.Codelab{RefURI: "REFPATH1", Codelab: types.Codelab{Meta: types.Meta{ID: "123", Title: "A title", Status: &published, Published: stringToContextTime(t, "1983-09-13"), Summary: "Awesome tutorial", URL: "https://tutorial1.com", Difficulty: 3, Categories: []string{"category1", "category2"}, Tags: []string{"foo", "bar"}, Duration: 60, Feedback: "http://feedback.com", Image: "image.png"}}, FilesWatched: []string{"onefile", "twofiles"}},
		codelab.Codelab{},
		codelab.Codelab{RefURI: "REFPATH2", Codelab: types.Codelab{Meta: types.Meta{ID: "456", Published: stringToContextTime(t, "1984-04-22")}}},
	}
	testCases := []struct {
		metaDir  string
		codelabs []codelab.Codelab

		wantReturnPath string
		wantAssets     []string
		wantErr        bool
	}{
		{"testdata/sites/valid", exCodelabs, "testdata/sites/valid/valid-api-output.json", []string{"event1.jpg", "event2.jpg"}, false},
		{"testdata/sites/categories-missing", exCodelabs, "", nil, true},
		{"testdata/sites/events-missing", exCodelabs, "", nil, true},
		{"testdata/sites/valid", []codelab.Codelab{}, "testdata/sites/valid/valid-without-codelab.json", []string{"event1.jpg", "event2.jpg"}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("generate api with codelab: %+v, metadata: %s", tc.codelabs, tc.metaDir), func(t *testing.T) {
			// Setup/Teardown
			p, teardown := paths.MockPath()
			defer teardown()
			apidir, teardown := testtools.TempDir(t)
			defer teardown()
			imagesdir, teardown := testtools.TempDir(t)
			defer teardown()
			p.MetaData = tc.metaDir
			p.API = apidir
			p.Images = imagesdir

			// Test
			dat, err := GenerateContent(tc.codelabs)

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

			files, err := ioutil.ReadDir(imagesdir)
			if err != nil {
				t.Fatalf("couldn't list %s: %v", imagesdir, err)
			}
			var assets []string
			for _, file := range files {
				assets = append(assets, file.Name())
			}
			sort.Strings(assets)
			sort.Strings(tc.wantAssets)
			if !reflect.DeepEqual(assets, tc.wantAssets) {
				t.Errorf("assets not matching. Got %+v; want %+v", assets, tc.wantAssets)
			}
		})
	}
}

func TestSaveAPI(t *testing.T) {
	// Setup/Teardown
	p, teardown := paths.MockPath()
	defer teardown()
	apidir, teardown := testtools.TempDir(t)
	defer teardown()
	p.API = apidir

	content := []byte("something")
	if err := Save(content); err != nil {
		t.Fatalf("Couldn't save API: %v", err)
	}
	f := path.Join(apidir, apiFileName)
	contentF, err := ioutil.ReadFile(f)
	if err != nil {
		t.Errorf("%s wasn't saved on disk: %v", f, err)
	}
	if !reflect.DeepEqual(contentF, content) {
		t.Errorf("Got %+v; want %+v", contentF, content)
	}
}

func TestSaveAPINoPath(t *testing.T) {
	// Setup/Teardown
	p, teardown := paths.MockPath()
	defer teardown()
	apidir, teardown := testtools.TempDir(t)
	defer teardown()
	p.API = apidir

	if err := os.Remove(apidir); err != nil {
		t.Fatalf("Couldn't remove api dir for test: %v", err)
	}

	if err := Save([]byte("something")); err != nil {
		t.Fatalf("Couldn't save API on non existing directory: %v", err)
	}
	f := path.Join(apidir, apiFileName)
	if _, err := os.Stat(f); err != nil {
		t.Errorf("%s was expected to exist when it doesn't: %v", f, err)
	}
}

func stringToContextTime(t *testing.T, date string) types.ContextTime {
	tt, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatalf("couldn't convert time from %s", date)
	}
	return types.ContextTime(tt)
}
