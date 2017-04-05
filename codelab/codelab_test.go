package codelab

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ubuntu/tutorial-deployment/consts"
)

var update = flag.Bool("update", false, "update generated files")

func TestGenerateCodelabs(t *testing.T) {
	var template = "testdata/template.html"
	var generatedpath = "testdata/codelabgenerated"

	testCases := []struct {
		src   string
		watch bool

		wantFilesWatched []string
		wantDiffFiles    []string
		wantErr          bool
	}{
		{"/doesnt/exist", false, nil, nil, true},
		{"testdata/codelabsrc/markdown-no-image.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-no-image.md", true, []string{"testdata/codelabsrc/markdown-no-image.md"}, nil, false},
		{"testdata/codelabsrc/markdown-invalid-generated-html.md", false, nil, []string{"example-snap-tutorial/index.inc"}, false},
		{"testdata/codelabsrc/markdown-with-images-simple.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-with-images-simple.md", true, []string{"testdata/codelabsrc/markdown-with-images-simple.md", "testdata/codelabsrc/foo.png"}, nil, false},
		{"testdata/codelabsrc/markdown-with-images-online.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-with-images-online.md", true, []string{"testdata/codelabsrc/markdown-with-images-online.md"}, nil, false}, // online images aren't tracked
		{"testdata/codelabsrc/markdown-with-images-relative-upper-path.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-with-images-duplicate-images.md", false, nil, nil, false}, // duplicated images have only one image
		{"testdata/codelabsrc/markdown-with-images-extension-preserved.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-with-images-online-jpg.md", false, nil, nil, false}, // it downloads the remote file in png
		{"testdata/codelabsrc/markdown-with-images.md", false, nil, nil, false},
		{"testdata/codelabsrc/markdown-with-images.md", true, []string{"testdata/codelabsrc/markdown-with-images.md", "testdata/codelabsrc/baz.jpg", "testdata/codelabsrc/foo.png", "testdata/bar.png"}, nil, false}, // watch local images only
		{"testdata/codelabsrc/markdown-missing-image.md", false, nil, nil, true},
		{"testdata/codelabsrc/markdown-modified-image.md", false, nil, nil, true},
		{fmt.Sprintf("%s1XUIwNcJj0IIFtza-py5BGDUlWoNeXyO2V0XgNOQvyDQ", consts.GdocPrefix), false, nil, nil, false},
		{fmt.Sprintf("%s17GGTeNbjAnnU3jNuKs9SrmQ_DhSqWJPmxxRSbWIjTiY", consts.GdocPrefix), false, nil, nil, false}, // with images
		{fmt.Sprintf("%s1XUIwNcJj0IIFtza-py5BGDUlWoNeXyO2V0XgNOQvyDQ", consts.GdocPrefix), true, nil, nil, false},  // no files to track
		{fmt.Sprintf("%s17GGTeNbjAnnU3jNuKs9SrmQ_DhSqWJPmxxRSbWIjTiY", consts.GdocPrefix), true, nil, nil, false},  // no files and images to track
		{fmt.Sprintf("%sinvalid", consts.GdocPrefix), false, nil, nil, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("generate %s, watch: %v", tc.src, tc.watch), func(t *testing.T) {
			out, teardown := tempDir(t)
			defer teardown()

			destcompare := path.Join(generatedpath, path.Base(tc.src))

			// On update, override destcompare
			if *update {
				// Skip the ones where we want an error to happen or where content isn't identical
				if tc.wantErr || tc.wantDiffFiles != nil {
					return
				}
				out = destcompare
				if err := os.RemoveAll(out); err != nil {
					t.Fatalf("err: %v", err)
				}
			}

			c, err := New(tc.src, out, template, tc.watch)

			if err != nil && *update {
				t.Fatalf("Couldn't update %s: An error occured: %v", tc.src, err)
			}

			if (err != nil) != tc.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			// we can't continue once we get an error: c isn't valid
			if err != nil {
				return
			}

			compareAll(t, destcompare, out, tc.wantDiffFiles)

			sort.Strings(c.FilesWatched)
			sort.Strings(tc.wantFilesWatched)
			if !reflect.DeepEqual(c.FilesWatched, tc.wantFilesWatched) {
				t.Errorf("got %+v; want %+v", c.FilesWatched, tc.wantFilesWatched)
			}
		})
	}
}

func tempDir(t *testing.T) (string, func()) {
	path, err := ioutil.TempDir("", "tutorial-test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return path, func() {
		if err := os.RemoveAll(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

// compare recursively all original and generated file content
func compareAll(t *testing.T, original, generated string, ignoresf []string) {
	if err := filepath.Walk(original, func(f string, fi os.FileInfo, err error) error {
		relp := strings.TrimPrefix(f, original)
		// root path
		if relp == "" {
			return nil
		}
		relp = relp[1:]
		p := path.Join(generated, relp)

		fo, err := os.Stat(p)
		if err != nil {
			t.Fatalf("%s doesn't exist while %s does", p, f)
		}

		if fi.IsDir() {
			if !fo.IsDir() {
				t.Fatalf("%s is a directory and %s isn't", f, p)
			}
			// else, it's a directory as well and we are done.
			return nil
		}

		wanted, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatalf("Couldn't read %s: %v", f, err)
		}
		actual, err := ioutil.ReadFile(p)
		if err != nil {
			t.Fatalf("Couldn't read %s: %v", p, err)
		}
		if !bytes.Equal(actual, wanted) && !contains(ignoresf, relp) {
			t.Errorf("%s and %s content differs:\nACTUAL:\n%s\n\nWANTED:\n%s", p, f, actual, wanted)
		}
		if bytes.Equal(actual, wanted) && contains(ignoresf, relp) {
			t.Errorf("We wanted %s and %s to differ and they don't", p, f)
		}
		return nil
	}); err != nil {
		t.Fatalf("err: %s", err)
	}

	// on the other side, check that all generated items are in origin
	if err := filepath.Walk(generated, func(f string, _ os.FileInfo, err error) error {
		relp := strings.TrimPrefix(f, generated)
		// root path
		if relp == "" {
			return nil
		}
		relp = relp[1:]
		p := path.Join(original, relp)

		if _, err := os.Stat(p); err != nil {
			t.Fatalf("%s doesn't exist while %s does", p, f)
		}
		return nil
	}); err != nil {
		t.Fatalf("err: %s", err)
	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
