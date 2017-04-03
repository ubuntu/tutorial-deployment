// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file is slightly modified from https://github.com/googlecodelabs/tools/blob/master/claat/auth.go

package claattools

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_driveClient(t *testing.T) {
	tests := []struct {
		name    string
		want    *http.Client
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DriveClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("driveClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("driveClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
