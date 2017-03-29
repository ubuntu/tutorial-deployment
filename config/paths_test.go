package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectPaths(t *testing.T) {
	testCases := []struct {
		websitePath  string
		exportPath   string
		metadataPath string
		cwd          string

		// the wanted paths are relative to cwd defined above
		wantWebsitePath  string
		wantExportPath   string
		wantMetaDataPath string

		errExpected bool
	}{
		{"/defined/website", "/other/export", "/metadata", "", "/defined/website", "/other/export", "/metadata", false},
		{"/defined/website", "export/path", "metadata", "", "/defined/website", "export/path", "metadata", false},
		{"", "export/path", "metadata", "testdata/nosite", "", "", "", true},                      // Error due to no site detected
		{"", "export/path", "metadata", "testdata/partialwebsite", "", "", "", true},              // Error due to no site detected
		{"", "export/path", "metadata", "testdata/website", "", "export/path", "metadata", false}, // Defined path are always relative to cwd, not website
		{"", "export/path", "metadata", "testdata/website/subdir", "..", "export/path", "metadata", false},
		{"", Paths.Export, Paths.MetaData, "testdata/website", "", defaultRelativeExportPath, defaultRelativeMetadataPath, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("website: %s, export: %s, metadata: %s in [%s]",
			tc.websitePath, tc.exportPath, tc.metadataPath, tc.cwd), func(t *testing.T) {
			// Setup/Teardown
			defer chdir(t, tc.cwd)()
			defer changePathObject(tc.websitePath, tc.exportPath, tc.metadataPath)()
			tc.wantWebsitePath = absPath(t, tc.wantWebsitePath)
			tc.wantExportPath = absPath(t, tc.wantExportPath)
			tc.wantMetaDataPath = absPath(t, tc.wantMetaDataPath)

			// Test
			err := DetectPaths()

			// Error checking
			if err != nil && !tc.errExpected {
				t.Errorf("DetectPaths errored out unexpectidely: %s", err)
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
		})
	}
}

func TestImportTutorialPaths(t *testing.T) {
	mp := "/foo/bar/"
	testCases := []struct {
		paths         []string
		expectedPaths []string
	}{
		{nil, []string{mp + defaultTutorialPathInMeta}},
		{[]string{"/rep1", "/rep2/tut1.md", "/rep3/rep5"}, []string{"/rep1", "/rep2/tut1.md", "/rep3/rep5"}},
		{[]string{"rep1", "../rep2/tut1.md", "rep3/rep5"}, []string{"rep1", "../rep2/tut1.md", "rep3/rep5"}},
		{[]string{"/foo/rep1"}, []string{"/foo/rep1"}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path argument: %+v", tc.paths), func(t *testing.T) {
			// Setup/Teardown
			p := P{
				MetaData: "/foo/bar",
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

func changePathObject(w, e, m string) func() {
	oldPath := Paths
	Paths = P{
		Website:  w,
		Export:   e,
		MetaData: m,
	}

	return func() {
		Paths = oldPath
	}
}
