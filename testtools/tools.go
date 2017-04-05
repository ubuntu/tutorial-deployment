package testtools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TempDir is a setup/teardown test helper for creating a temporary directory
func TempDir(t *testing.T) (string, func()) {
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

// StringContains check an element is contained in a slice of string
func StringContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// AbsPath is a simple abspath for absolute path test with test failure
func AbsPath(t *testing.T, path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return path
}

// Chdir is a setup/teadown test helper for changing current directory
func Chdir(t *testing.T, dir string) func() {
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
