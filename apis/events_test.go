package apis

import (
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"github.com/ubuntu/tutorial-deployment/paths"

	"os"
)

func TestNewEvents(t *testing.T) {
	testCases := []struct {
		eventsDir string

		wantEvents Events
		wantErr    bool
	}{
		{"testdata/events/valid",
			Events{"event-1": event{Name: "Event 1", Logo: "img/event1.jpg", Description: "This workshop is taking place at Event 1."},
				"event-2": event{Name: "Event 2", Logo: "event2.jpg", Description: "This workshop is taking place at Event 2."},
			},
			false},
		{"doesnt/exist", nil, true},
		{"testdata/events/valid-missing-image", // we still load correctly, we don't touch images at this stage
			Events{"event-1": event{Name: "Event 1", Logo: "img/event1.jpg", Description: "This workshop is taking place at Event 1."},
				"event-2": event{Name: "Event 2", Logo: "event2.jpg", Description: "This workshop is taking place at Event 2."},
			},
			false},
		{"testdata/events/no-events", Events{}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("create event for: %+v", tc.eventsDir), func(t *testing.T) {
			// Setup/Teardown
			p, teardown := paths.MockPath()
			defer teardown()
			p.MetaData = tc.eventsDir

			// Test
			e, err := NewEvents()

			if (err != nil) != tc.wantErr {
				t.Errorf("NewEvents() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(*e, tc.wantEvents) {
				t.Errorf("Generated events: got %+v; want %+v", *e, tc.wantEvents)
			}
		})
	}
}

func TestSaveImages(t *testing.T) {
	testCases := []struct {
		eventsDir string
		eventsObj Events

		wantEvents Events
		wantErr    bool
	}{
		{"testdata/events/valid",
			Events{"event-1": event{Name: "Event 1", Logo: "img/event1.jpg", Description: "This workshop is taking place at Event 1."},
				"event-2": event{Name: "Event 2", Logo: "event2.jpg", Description: "This workshop is taking place at Event 2."},
			},
			Events{"event-1": event{Name: "Event 1", Logo: "assets/event1.jpg", Description: "This workshop is taking place at Event 1."},
				"event-2": event{Name: "Event 2", Logo: "assets/event2.jpg", Description: "This workshop is taking place at Event 2."},
			},
			false},
		{"testdata/events/valid-missing-image",
			Events{"event-1": event{Name: "Event 1", Logo: "img/event1.jpg", Description: "This workshop is taking place at Event 1."},
				"event-2": event{Name: "Event 2", Logo: "event2.jpg", Description: "This workshop is taking place at Event 2."},
			},
			nil,
			true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("save events: %+v", tc.eventsDir), func(t *testing.T) {
			// Setup/Teardown
			out, teardown := tempDir(t)
			defer teardown()
			p, teardown := paths.MockPath()
			defer teardown()
			p.MetaData = tc.eventsDir
			p.API = out

			// Test
			err := tc.eventsObj.SaveImages()

			if (err != nil) != tc.wantErr {
				t.Errorf("SaveImages() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(tc.eventsObj, tc.wantEvents) {
				t.Errorf("Image paths not changed in event: got %+v; want %+v", tc.eventsObj, tc.wantEvents)
			}
			for _, e := range tc.wantEvents {
				imgP := path.Join(p.API, e.Logo)
				if _, err := os.Stat(imgP); os.IsNotExist(err) {
					t.Errorf("%s doesn't exist when we wanted it", imgP)
				}
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
