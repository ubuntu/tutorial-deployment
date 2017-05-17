package codelab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/didrocks/codelab-ubuntu-tools/claat/parser"
	"github.com/didrocks/codelab-ubuntu-tools/claat/render"
	"github.com/didrocks/codelab-ubuntu-tools/claat/types"
	"github.com/ubuntu/tutorial-deployment/claattools"
	"github.com/ubuntu/tutorial-deployment/consts"

	// allow parsers to register themselves
	_ "github.com/didrocks/codelab-ubuntu-tools/claat/parser/gdoc"
	_ "github.com/didrocks/codelab-ubuntu-tools/claat/parser/md"
)

const (
	appspotPreviewURL = "https://codelabs-preview.appspot.com/?file_id="
	relativeImgDir    = "img" // img relative directory in codelab
	metaFilename      = "codelab.json"
)

// Codelab augments claat Codelab object by owning all Codelab Metadata and last updated time
type Codelab struct {
	RefURI string `json:"-"` // Reference uri path
	types.Codelab
	FilesWatched []string  `json:"-"`               // Path to asset files to watch
	HideSteps    *struct{} `json:"Steps,omitempty"` // Hide the Steps json export from types.Codelab with this nil object

	watch    bool   // We will need to watch files
	dir      string // path where the codelab is stored
	template string // template path used
}

// New retrieves and parses codelab source.
func New(codelabRef, dest, template string, watch bool) (*Codelab, error) {
	c := Codelab{
		RefURI:   codelabRef,
		template: template,
		watch:    watch,
	}
	if err := c.download(); err != nil {
		return nil, err
	}
	c.dir = filepath.Join(dest, c.ID)
	if err := c.downloadAssets(); err != nil {
		return nil, err
	}
	if err := c.writeCodelab(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Refresh content and assets of given codelab
func (c *Codelab) Refresh() error {
	if err := c.wipe(); err != nil {
		return err
	}
	c.FilesWatched = nil
	if err := c.download(); err != nil {
		return err
	}
	if err := c.downloadAssets(); err != nil {
		return err
	}
	return c.writeCodelab()
}

// download and parse codelab content
// The function will also fetch, parse and integrate its imports
func (c *Codelab) download() error {
	res, err := claattools.Fetch(c.RefURI)
	if err != nil {
		return fmt.Errorf("failed getting: %v", err)
	}
	defer res.Body.Close()
	clab, err := parser.Parse(res.Type, res.Body)
	if err != nil {
		return err
	}

	// fetch imports and parse them as fragments
	var imports []*types.ImportNode
	for _, st := range clab.Steps {
		imports = append(imports, claattools.GetImportNodes(st.Content.Nodes)...)
	}
	ch := make(chan error, len(imports))
	defer close(ch)
	for _, imp := range imports {
		go func(n *types.ImportNode) {
			frag, err := getFragment(n.URL)
			if err != nil {
				ch <- fmt.Errorf("%s from import: %s", n.URL, err)
				return
			}
			n.Content.Nodes = frag
			c.appendResourceToWatchFile(n.URL)
			ch <- nil
		}(imp)
	}
	for _ = range imports {
		if err := <-ch; err != nil {
			return err
		}
	}

	c.Codelab = *clab
	c.appendResourceToWatchFile(c.RefURI)
	return nil
}

var crcTable = crc64.MakeTable(crc64.ECMA)

// downloadAssets get images and other assets associated to the codelab
func (c *Codelab) downloadAssets() (err error) {
	imgDir := path.Join(c.dir, relativeImgDir)
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return err
	}

	// Handle google drive download: all images will use the same driveClient element.
	var client *http.Client
	if strings.HasPrefix(c.RefURI, consts.GdocPrefix) {
		client, err = claattools.DriveClient()
		if err != nil {
			return err
		}
	}

	type res struct {
		src  string
		dest string
		err  error
	}
	ch := make(chan res)
	defer close(ch)
	var nImages int
	for _, st := range c.Steps {
		nodes := claattools.GetImageNodes(st.Content.Nodes)
		nImages += len(nodes)
		for _, n := range nodes {
			go func(n *types.ImageNode) {
				// src can be remote or local
				imgURL := n.Src
				u, err := url.Parse(imgURL)
				if err != nil {
					ch <- res{imgURL, "", err}
					return
				}
				var b []byte
				var ext string
				// read (optionally download) image filename
				if u.Host == "" {
					imgURL = path.Join(path.Dir(c.RefURI), imgURL)
					b, err = ioutil.ReadFile(imgURL)
					ext = path.Ext(imgURL)
				} else {
					b, err = claattools.FetchRemoteBytes(client, imgURL, 5)
					ext = ".png"
				}
				if err != nil {
					ch <- res{imgURL, "", err}
					return
				}

				// compute checksum which will be new file name and write it
				crc := crc64.Checksum(b, crcTable)
				name := fmt.Sprintf("%x%s", crc, ext)
				n.Src = fmt.Sprintf("CODELABURL/%s/%s", relativeImgDir, name)
				dest := filepath.Join(imgDir, name)
				if err = ioutil.WriteFile(dest, b, 0644); err != nil {
					ch <- res{imgURL, dest, err}
					return
				}

				ch <- res{imgURL, dest, nil}
			}(n)
		}
	}

	// fetch possible errors
	var errs bytes.Buffer
	for i := 0; i < nImages; i++ {
		r := <-ch
		if r.err != nil {
			errs.WriteString(fmt.Sprintf("Couldn't copy %s => %s: %v\n", r.src, r.dest, r.err))
			continue
		}
		c.appendResourceToWatchFile(r.src)
	}

	if errs.Len() > 0 {
		err = errors.New(errs.String())
	}
	return err
}

// write codelab itself to disk: html content and json metadata
func (c *Codelab) writeCodelab() error {
	// make sure codelab dir exists
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}

	// write metadata
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(c.dir, metaFilename), b, 0644); err != nil {
		return err
	}

	// main content file(s)
	f, err := os.Create(filepath.Join(c.dir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return render.Execute(f, c.template, c)
}

// wipe output directory content for codelab
// Used when autorefreshing
func (c *Codelab) wipe() error {
	return os.RemoveAll(c.dir)
}

func (c *Codelab) appendResourceToWatchFile(refPath string) error {
	if !c.watch {
		return nil
	}

	u, err := url.Parse(refPath)
	if err != nil {
		return fmt.Errorf("Couldn't parse url to append resource watch file: %s", refPath)
	}

	if u.Host == "" && strings.HasPrefix(refPath, consts.GdocPrefix) {
		gdocID := strings.TrimPrefix(refPath, consts.GdocPrefix)
		log.Printf("Can't track %s for changes as it refers to a google doc. You can head over "+
			"to %s%s to preview it with a default template dynamically.", gdocID, appspotPreviewURL, gdocID)
		return nil
	}

	if u.Host != "" {
		log.Printf("%s: Can't track %s for changes as it's a remote resource. You will need to rerun this binary and "+
			"refresh your page to preview the changes", c.RefURI, refPath)
		return nil
	}

	// don't add duplicates
	for _, f := range c.FilesWatched {
		if f == refPath {
			return nil
		}
	}
	c.FilesWatched = append(c.FilesWatched, refPath)
	return nil
}

func getFragment(url string) ([]types.Node, error) {
	res, err := claattools.FetchRemote(url, true)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return parser.ParseFragment(res.Type, res.Body)
}
