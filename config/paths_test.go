package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDetectPaths(t *testing.T) {
	testCases := []struct {
		websitePath  string
		exportPath   string
		metadataPath string
		apiPath      string
		cwd          string

		// the wanted paths are relative to cwd defined above
		wantWebsitePath  string
		wantExportPath   string
		wantMetaDataPath string
		wantAPIPath      string

		errExpected bool
	}{
		{"/defined/website", "/other/export", "/metadata", "/api", "", "/defined/website", "/other/export", "/metadata", "/api", false},
		{"/defined/website", "export/path", "alt/metadata", "alt/api", "", "/defined/website", "export/path", "alt/metadata", "alt/api", false},
		{"", "export/path", "alt/metadata", "alt/api", "testdata/nosite", "", "", "", "", true},                                 // Error due to no site detected
		{"", "export/path", "alt/metadata", "alt/api", "testdata/partialwebsite", "", "", "", "", true},                         // Error due to no site detected
		{"", "export/path", "alt/metadata", "alt/api", "testdata/website", "", "export/path", "alt/metadata", "alt/api", false}, // Defined path are always relative to cwd, not website
		{"", "export/path", "alt/metadata", "alt/api", "testdata/website/subdir", "..", "export/path", "alt/metadata", "alt/api", false},
		{"", Paths.Export, Paths.MetaData, Paths.API, "testdata/website", "", defaultRelativeExportPath, defaultRelativeMetadataPath, defaultRelativeAPIPath, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("(website: %s), (export: %s), (metadata: %s), (api: %s) in [%s]",
			tc.websitePath, tc.exportPath, tc.metadataPath, tc.apiPath, tc.cwd), func(t *testing.T) {
			// Setup/Teardown
			defer chdir(t, tc.cwd)()
			defer changePathObject(tc.websitePath, tc.exportPath, tc.metadataPath, tc.apiPath)()
			tc.wantWebsitePath = absPath(t, tc.wantWebsitePath)
			tc.wantExportPath = absPath(t, tc.wantExportPath)
			tc.wantMetaDataPath = absPath(t, tc.wantMetaDataPath)
			tc.wantAPIPath = absPath(t, tc.apiPath)

			// Test
			err := DetectPaths()

			// Error checking
			if err != nil && !tc.errExpected {
				t.Errorf("DetectPaths errored out unexpectedly: %s", err)
			}
			if err == nil && tc.errExpected {
				t.Error("DetectPaths expected an error and didn't")
			}
			if err != nil && tc.errExpected {
				// Error is fatal, we don't care about paths
				// We don't disable path checking if the error wasn't expected to help us diving into any issue
				return
			}

			// Paths checks
			if Paths.Website != tc.wantWebsitePath {
				t.Errorf("Website: got %s; want %s", Paths.Website, tc.wantWebsitePath)
			}
			if Paths.Export != tc.wantExportPath {
				t.Errorf("Export: got %s; want %s", Paths.Export, tc.wantExportPath)
			}
			if Paths.MetaData != tc.wantMetaDataPath {
				t.Errorf("Metadata: got %s; want %s", Paths.MetaData, tc.wantMetaDataPath)
			}
			if Paths.API != tc.wantAPIPath {
				t.Errorf("API: got %s; want %s", Paths.API, tc.wantAPIPath)
			}
		})
	}
}

func TestImportTutorialPaths(t *testing.T) {
	website := "/ws/"
	testCases := []struct {
		paths         []string
		expectedPaths []string
	}{
		{nil, []string{website + defaultTutorialPath}},
		{[]string{"/rep1", "/rep2/tut1.md", "/rep3/rep5"}, []string{"/rep1", "/rep2/tut1.md", "/rep3/rep5"}},
		{[]string{"rep1", "../rep2/tut1.md", "rep3/rep5"}, []string{"rep1", "../rep2/tut1.md", "rep3/rep5"}},
		{[]string{"/foo/rep1"}, []string{"/foo/rep1"}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path argument: %+v", tc.paths), func(t *testing.T) {
			// Setup/Teardown
			p := P{
				Website: website,
			}
			for i, expected := range tc.expectedPaths {
				tc.expectedPaths[i] = absPath(t, expected)
			}

			// Test
			err := p.ImportTutorialPaths(tc.paths)
			if err != nil {
				t.Errorf("err: %s", err)
			}

			if !reflect.DeepEqual(p.TutorialInputs, tc.expectedPaths) {
				t.Errorf("Import path: got %+v; want %+v", p.TutorialInputs, tc.expectedPaths)
			}
		})
	}
}

func TestCreateTempPathHandling(t *testing.T) {
	p := P{}

	// Create temp dir
	if err := p.CreateTempOutPath(); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.API == "" || p.Export == "" {
		t.Errorf("one of API (%s) or Export (%s) is empty", p.API, p.Export)
		return
	}
	tmpdir := p.API[:len(p.API)-len(defaultRelativeAPIPath)]
	if !strings.HasPrefix(p.Export, tmpdir) {
		t.Errorf("API (%s) and Export (%s) don't have the same temporary prefix", p.API, p.Export)
	}
	if _, err := os.Stat(tmpdir); os.IsNotExist(err) {
		t.Errorf("%s doesn't exists", tmpdir)
		return
	}

	// Remove temp dir
	if err := p.CleanTempPath(); err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := os.Stat(tmpdir); err == nil {
		t.Errorf("%s still exists", tmpdir)
	}
	if p.API != "" || p.Export != "" {
		t.Errorf("API (%s) and Export (%s) should now be empty", p.API, p.Export)
	}
}

func TestTryCleanNonTempDir(t *testing.T) {
	p := P{}

	if err := p.CleanTempPath(); err == nil {
		t.Errorf("Cleaning a non temporary path object should have returned an error: %+v", p)
	}
}

func chdir(t *testing.T, dir string) func() {
	if dir == "" {
		return func() {}
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %s", err)
	}
	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

func absPath(t *testing.T, path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return path
}

func changePathObject(w, e, m, a string) func() {
	oldPath := Paths
	Paths = P{
		Website:  w,
		Export:   e,
		MetaData: m,
		API:      a,
	}

	return func() {
		Paths = oldPath
	}
}
