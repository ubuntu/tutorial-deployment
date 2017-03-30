package internaltools

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUnique(t *testing.T) {
	testCases := []struct {
		elems    []string
		expected []string
	}{
		{[]string{"/foo/bar", "/foo/baz", "ta"}, []string{"/foo/bar", "/foo/baz", "ta"}},
		{[]string{"/foo/bar", "/foo/baz", "/foo/bar", "ta"}, []string{"/foo/bar", "/foo/baz", "ta"}},
		{nil, nil},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with %s", tc.elems), func(t *testing.T) {
			elems := UniqueStrings(tc.elems)
			if !reflect.DeepEqual(elems, tc.expected) {
				t.Errorf("got %+v; want %+v", elems, tc.expected)
			}
		})
	}
}
