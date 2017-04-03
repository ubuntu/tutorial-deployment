package paths

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

const (
	defaultRelativeExportPath   = "src/codelabs"
	defaultRelativeMetadataPath = "metadata"
	defaultRelativeAPIPath      = "api"
	defaultTutorialPath         = "tutorials"

	// GdocFilename for tutorials in google doc format.
	GdocFilename = "gdoc.def"
)

func init() {
	p := New()
	flag.StringVar(&p.Website, "w", "", "website root path directory where main index.html is located. Will "+
		"autodetect if current directory is within the website repository")
	flag.StringVar(&p.Export, "e", defaultRelativeExportPath,
		fmt.Sprintf("export path for generated tutorials. Default is [WEBSITE_PATH]/%s", defaultRelativeExportPath))
	flag.StringVar(&p.MetaData, "i", defaultRelativeMetadataPath,
		fmt.Sprintf("import path for metadata as template and events definition. Default is [WEBSITE_PATH]/%s", defaultRelativeMetadataPath))
	flag.StringVar(&p.API, "a", defaultRelativeAPIPath,
		fmt.Sprintf("exported apis for generated tutorials. Default is [WEBSITE_PATH]/%s", defaultRelativeAPIPath))
}

var (
	// paths encapsulate global path properties of the project
	paths   Path
	onePath sync.Once
)

// Path is used for the main Paths object
type Path struct {
	Website        string
	Export         string
	MetaData       string
	API            string
	TutorialInputs []string

	// are out paths export and api temporary? (prevent accidental deletion)
	tempRootPath string
}

// New get access to the singleton path and create it if necessary (multi-threading safe)
func New() *Path {
	onePath.Do(func() {
		paths = Path{
			Export:   defaultRelativeExportPath,
			MetaData: defaultRelativeMetadataPath,
		}
	})
	return &paths
}

// ImportTutorialPaths sanitizes relative paths, adding default if none provided
func (p *Path) ImportTutorialPaths(tps []string) (err error) {
	// default: tutorial and google doc reference path
	if len(tps) == 0 {
		tps = []string{path.Join(p.Website, defaultTutorialPath)}
	}
	for i, tp := range tps {
		if tp, err = filepath.Abs(tp); err != nil {
			if err != nil {
				return err
			}
		}
		tps[i] = tp
	}
	p.TutorialInputs = tps
	return nil
}

// CreateTempOutPath generate some temporary paths for API and export
func (p *Path) CreateTempOutPath() error {
	tmp, err := ioutil.TempDir("", "serve-tutorial-")
	if err != nil {
		return fmt.Errorf("Couldn't create temp path: %s", err)
	}
	p.tempRootPath = tmp
	p.API = path.Join(tmp, defaultRelativeAPIPath)
	p.Export = path.Join(tmp, defaultRelativeExportPath)
	return nil
}

// CleanTempPath removes all generated paths
func (p *Path) CleanTempPath() error {
	if p.tempRootPath == "" {
		return fmt.Errorf("No path in %+v corresponding to temporary paths", p)
	}
	p.API = ""
	p.Export = ""
	return os.RemoveAll(p.tempRootPath)
}

// DetectPaths search for paths and load them accordingly to flags
// this needs to be called after parsing the CLI args.
func (p *Path) DetectPaths() (err error) {
	// We only need Website if one of the values aren't defined
	if p.Website == "" &&
		(p.Export == defaultRelativeExportPath || p.MetaData == defaultRelativeMetadataPath || p.API == defaultRelativeAPIPath) {
		p.Website, err = detectWebsitePath()
		if err != nil {
			return err
		}
	}

	if err = sanitizeRelPathToWebsite(&p.Export, defaultRelativeExportPath, p.Website); err != nil {
		return err
	}
	if err = sanitizeRelPathToWebsite(&p.MetaData, defaultRelativeMetadataPath, p.Website); err != nil {
		return err
	}
	if err = sanitizeRelPathToWebsite(&p.API, defaultRelativeAPIPath, p.Website); err != nil {
		return err
	}
	return nil
}

func detectWebsitePath() (string, error) {
	rootDirFiles := [...]string{"index.html", "bower.json"}
	initdir := "."
	initdir, err := filepath.Abs(initdir)
	if err != nil {
		return "", err
	}
	dir := initdir

RootDirsLoop:
	for dir != "/" {
		for _, file := range rootDirFiles {
			if _, err := os.Stat(path.Join(dir, file)); os.IsNotExist(err) {
				dir = path.Join(dir, "..")
				dir, err = filepath.Abs(dir)
				if err != nil {
					return "", err
				}
				continue RootDirsLoop
			}
		}
		break
	}
	// Root can't be website directory
	if dir == "/" {
		return "", fmt.Errorf("Couldn't detect website directory from and in parent directories: %s", initdir)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func sanitizeRelPathToWebsite(p *string, defP, website string) (err error) {
	if *p == defP {
		*p = path.Join(website, defP)
	}
	if *p, err = filepath.Abs(*p); err != nil {
		return err
	}
	return nil
}
