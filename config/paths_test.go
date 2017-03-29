package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	for _, c := range testCases {
		t.Run(fmt.Sprintf("website: %s, export: %s, metadata: %s in [%s]",
			c.websitePath, c.exportPath, c.metadataPath, c.cwd), func(t *testing.T) {
			// Setup/Teardown
			defer chdir(t, c.cwd)()
			defer changePathObject(c.websitePath, c.exportPath, c.metadataPath)()
			c.wantWebsitePath = absPath(t, c.wantWebsitePath)
			c.wantExportPath = absPath(t, c.wantExportPath)
			c.wantMetaDataPath = absPath(t, c.wantMetaDataPath)

			// Test
			err := DetectPaths()

			// Error checking
			if err != nil && !c.errExpected {
				t.Errorf("DetectPaths errored out unexpectidely: %s", err)
			}
			if err == nil && c.errExpected {
				t.Error("DetectPaths expected an error and didn't")
			}
			if err != nil {
				return // Error is fatal, we don't care about paths
			}

			// Paths checks
			if Paths.Website != c.wantWebsitePath {
				t.Errorf("Website: got %s; want %s", Paths.Website, c.wantWebsitePath)
			}
			if Paths.Export != c.wantExportPath {
				t.Errorf("Export: got %s; want %s", Paths.Export, c.wantExportPath)
			}
			if Paths.MetaData != c.wantMetaDataPath {
				t.Errorf("Metadata: got %s; want %s", Paths.MetaData, c.wantMetaDataPath)
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
