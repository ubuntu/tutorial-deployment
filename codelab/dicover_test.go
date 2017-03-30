package codelab

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ubuntu/tutorial-deployment/paths"
)

func TestDiscover(t *testing.T) {
	testCases := []struct {
		tutorialPaths     []string
		expectedTutorials []string
		errExpected       bool
	}{
		{[]string{}, nil, false},
		{[]string{"/doesnt/exist"}, nil, true},
		{[]string{"testdata/nothing"}, nil, false},
		{[]string{"testdata/flat"}, []string{"testdata/flat/tut1.md", "testdata/flat/tut2.md"}, false},
		{[]string{"testdata/nested"}, []string{"testdata/nested/subdir1/subsub/tut1.md", "testdata/nested/subdir1/subsub/tut2.md", "testdata/nested/subdir2/tut1.md", "testdata/nested/subdir2/tut2.md"}, false},
		{[]string{"testdata/flat", "testdata/flat2"}, []string{"testdata/flat/tut1.md", "testdata/flat/tut2.md", "testdata/flat2/tut1.md"}, false},
		{[]string{"testdata/withgdoc"}, []string{"gdoc:mytut1", "gdoc:mytut2"}, false},
		{[]string{"testdata/withignored"}, []string{"testdata/withignored/tut1.md"}, false},
		{[]string{"testdata/flat", "testdata/withgdoc", "testdata/withignored"}, []string{"testdata/flat/tut1.md", "testdata/flat/tut2.md", "gdoc:mytut1", "gdoc:mytut2", "testdata/withignored/tut1.md"}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("scanning %s", tc.tutorialPaths), func(t *testing.T) {
			// Setup/Teardown
			defer func(i []string) func() {
				p := paths.New()
				origInputPaths := p.TutorialInputs
				p.TutorialInputs = i
				return func() {
					p.TutorialInputs = origInputPaths
				}
			}(tc.tutorialPaths)()

			tutorials, err := Discover()

			if err != nil && !tc.errExpected {
				t.Errorf("Discover errored out unexpectedly: %s", err)
			}
			if err == nil && tc.errExpected {
				t.Error("Discvoer expected an error and didn't")
			}
			if !reflect.DeepEqual(tutorials, tc.expectedTutorials) {
				t.Errorf("got %+v; want %+v", tutorials, tc.expectedTutorials)
			}
		})
	}
}
