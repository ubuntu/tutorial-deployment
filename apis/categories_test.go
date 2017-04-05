package apis

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ubuntu/tutorial-deployment/paths"
)

func TestNewCategories(t *testing.T) {
	testCases := []struct {
		categoriesDir string

		wantCategories Categories
		wantErr        bool
	}{
		{"testdata/categories/valid",

			Categories{"snap": category{Lightcolor: "var(--paper-indigo-300)", Maincolor: "var(--paper-indigo-500)", Secondarycolor: "var(--paper-indigo-700)"},
				"snapcraft": category{Lightcolor: "var(--paper-teal-300)", Maincolor: "var(--paper-teal-500)", Secondarycolor: "var(--paper-teal-700)"},
				"unknown":   category{Lightcolor: "#444", Maincolor: "#444", Secondarycolor: "#444"},
			},
			false},
		{"doesnt/exist", nil, true},
		{"testdata/categories/no-categories", Categories{}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("create categories for: %+v", tc.categoriesDir), func(t *testing.T) {
			// Setup/Teardown
			p, teardown := paths.MockPath()
			defer teardown()
			p.MetaData = tc.categoriesDir

			// Test
			c, err := NewCategories()

			if (err != nil) != tc.wantErr {
				t.Errorf("NewCategories() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(*c, tc.wantCategories) {
				t.Errorf("Generated categories: got %+v; want %+v", *c, tc.wantCategories)
			}
		})
	}
}
