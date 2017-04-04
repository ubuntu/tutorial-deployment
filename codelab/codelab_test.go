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
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update generated files")

func TestGenerateCodelabs(t *testing.T) {

	var template = "testdata/template.html"
	var generatedpath = "testdata/codelabgenerated"

	if *update {
		if err := os.RemoveAll(generatedpath); err != nil {
			t.Fatalf("err: %v", generatedpath)
		}
	}

	testCases := []struct {
		src   string
		watch bool

		wantFilesWatched []string
		wantErr          bool
	}{
		{"testdata/codelabsrc/markdown-no-image.md", false, nil, false},
		{"testdata/codelabsrc/markdown-no-image.md", true, []string{"testdata/codelabsrc/markdown-no-image.md"}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("generate %s", tc.src), func(t *testing.T) {
			out, teardown := tempDir(t)
			defer teardown()

			destcompare := path.Join(generatedpath, path.Base(tc.src))

			// On update, override destcompare
			if *update {
				// Skip the ones where we want an error to happen
				if tc.wantErr {
					return
				}
				out = destcompare
			}

			c, err := New(tc.src, out, template, tc.watch)

			if err != nil && *update {
				t.Fatalf("Couldn't update %s: An error occurend: %v", tc.src, err)
			}

			if (err != nil) != tc.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			compareAll(t, destcompare, out)

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
func compareAll(t *testing.T, original, generated string) {
	if err := filepath.Walk(original, func(f string, fi os.FileInfo, err error) error {
		relp := strings.TrimPrefix(f, original)
		// root path
		if relp == "" {
			return nil
		}
		p := path.Join(generated, relp[1:])

		if fi.IsDir() {
			// we just test existing and go to next
			fi, err := os.Stat(p)
			if err != nil {
				t.Fatalf("%s is a directory and %s doesn't exist", f, p)
			}
			if !fi.IsDir() {
				t.Fatalf("%s is a directory and %s isn't", f, p)
			}
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
		if !bytes.Equal(actual, wanted) {
			t.Errorf("%s and %s content differs", p, f)
		}
		return nil
	}); err != nil {
		t.Fatalf("err: %s", err)
	}
}
