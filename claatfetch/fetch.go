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

// Most of this file is heavily inspired from https://github.com/googlecodelabs/tools/blob/master/claat/fetch.go

package claatfetch

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ubuntu/tutorial-deployment/consts"
)

// Resource is a codelab Resource, loaded from local file
// or fetched from remote location.
type Resource struct {
	body io.ReadCloser // resource body
	mod  time.Time     // last update of content
}

// driveAPI is a base URL for Drive API
const driveAPI = "https://www.googleapis.com/drive/v3"

// Fetch retrieves codelab doc either from local disk
// or a remote location.
// The caller is responsible for closing returned stream.
func Fetch(name string) (*Resource, error) {
	fi, err := os.Stat(name)
	if os.IsNotExist(err) {
		return fetchRemote(name, false)
	}
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return &Resource{
		body: r,
		mod:  fi.ModTime(),
	}, nil
}

// fetchRemote retrieves resource r from the network.
//
// If urlStr is not a URL, i.e. does not have the host part and is prepended by gdoc:, it is considered to be
// a Google Doc ID and fetched accordingly. Otherwise, a simple GET request
// is used to retrieve the contents.
//
// The caller is responsible for closing returned stream.
// If nometa is true, Resource.mod may have zero value.
func fetchRemote(urlStr string, nometa bool) (*Resource, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if (u.Host == "" && strings.HasPrefix(urlStr, consts.GdocPrefix)) || u.Host == "docs.google.com" {
		return fetchDriveFile(strings.TrimPrefix(urlStr, consts.GdocPrefix), nometa)
	}
	return fetchRemoteFile(urlStr)
}

// fetchRemoteFile retrieves codelab resource from url.
// It is a special case of fetchRemote function.
func fetchRemoteFile(url string) (*Resource, error) {
	res, err := retryGet(nil, url, 3)
	if err != nil {
		return nil, err
	}
	t, err := http.ParseTime(res.Header.Get("last-modified"))
	if err != nil {
		t = time.Now()
	}
	return &Resource{
		body: res.Body,
		mod:  t,
	}, nil
}

// fetchDriveFile uses Drive API to retrieve HTML representation of a Google Doc.
// See https://developers.google.com/drive/web/manage-downloads#downloading_google_documents
// for more details.
//
// If nometa is true, resource.mod will have zero value.
func fetchDriveFile(id string, nometa bool) (*Resource, error) {
	id = gdocID(id)
	exportURL := gdocExportURL(id)
	client, err := driveClient()
	if err != nil {
		return nil, err
	}

	if nometa {
		res, err := retryGet(client, exportURL, 7)
		if err != nil {
			return nil, err
		}
		return &Resource{body: res.Body}, nil
	}

	u := fmt.Sprintf("%s/files/%s?fields=id,mimeType,modifiedTime", driveAPI, id)
	res, err := retryGet(client, u, 7)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	meta := &struct {
		ID       string    `json:"id"`
		MimeType string    `json:"mimeType"`
		Modified time.Time `json:"modifiedTime"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(meta); err != nil {
		return nil, err
	}
	if meta.MimeType != "application/vnd.google-apps.document" {
		return nil, fmt.Errorf("%s: invalid mime type: %s", id, meta.MimeType)
	}

	if res, err = retryGet(client, exportURL, 7); err != nil {
		return nil, err
	}
	return &Resource{
		body: res.Body,
		mod:  meta.Modified,
	}, nil
}

// retryGet tries to GET specified url up to n times.
// Default client will be used if not provided.
func retryGet(client *http.Client, url string, n int) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}
	for i := 0; i <= n; i++ {
		if i > 0 {
			t := time.Duration((math.Pow(2, float64(i)) + rand.Float64()) * float64(time.Second))
			time.Sleep(t)
		}
		res, err := client.Get(url)
		// return early with a good response
		// the rest is error handling
		if err == nil && res.StatusCode == http.StatusOK {
			return res, nil
		}

		// sometimes Drive API wouldn't even start a response,
		// we get net/http: TLS handshake timeout instead:
		// consider this a temporary failure and retry again
		if err != nil {
			continue
		}
		// otherwise, decode error response and check for "rate limit"
		defer res.Body.Close()
		var erres struct {
			Error struct {
				Errors []struct{ Reason string }
			}
		}
		b, _ := ioutil.ReadAll(res.Body)
		json.Unmarshal(b, &erres)
		var rateLimit bool
		for _, e := range erres.Error.Errors {
			if e.Reason == "rateLimitExceeded" || e.Reason == "userRateLimitExceeded" {
				rateLimit = true
				break
			}
		}
		// this is neither a rate limit error, nor a server error:
		// retrying is useless
		if !rateLimit && res.StatusCode < http.StatusInternalServerError {
			return nil, fmt.Errorf("fetch %s: %s; %s", url, res.Status, b)
		}
	}
	return nil, fmt.Errorf("%s: failed after %d retries", url, n)
}

func gdocID(url string) string {
	const s = "/document/d/"
	if i := strings.Index(url, s); i >= 0 {
		url = url[i+len(s):]
	}
	if i := strings.IndexRune(url, '/'); i > 0 {
		url = url[:i]
	}
	return url
}

func gdocExportURL(id string) string {
	return fmt.Sprintf("%s/files/%s/export?mimeType=text/html", driveAPI, id)
}
