package config

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// P is used for the main Paths object
type P struct {
	Website        string
	Export         string
	MetaData       string
	TutorialInputs []string
}

// Paths encapsulate global path properties of the project
var Paths = P{
	Export:   defaultRelativeExportPath,
	MetaData: defaultRelativeMetadataPath,
}

const (
	defaultRelativeExportPath   = "src/codelabs"
	defaultRelativeMetadataPath = "content"

	defaultTutorialPathInMeta = "tutorials"
	// GdocFilename for tutorials in google doc format.
	GdocFilename = "gdoc.def"
)

func init() {
	flag.StringVar(&(Paths.Website), "w", "", "website root path directory where main index.html is located. Will "+
		"autodetect if current directory is within the website repository")
	flag.StringVar(&(Paths.Export), "e", defaultRelativeExportPath,
		fmt.Sprintf("export path for generated tutorials. Default is [WEBSITE_PATH]/%s", defaultRelativeExportPath))
	flag.StringVar(&(Paths.MetaData), "i", defaultRelativeMetadataPath,
		fmt.Sprintf("import path for metadata and default tutorials. Default is [WEBSITE_PATH]/%s", defaultRelativeMetadataPath))
}

// ImportTutorialPaths sanitizes relative paths, adding default if none provided
func (p *P) ImportTutorialPaths(tps []string) (err error) {
	// default: tutorial and google doc reference path
	if len(tps) == 0 {
		tps = []string{path.Join(p.MetaData, defaultTutorialPathInMeta),
			path.Join(p.MetaData, GdocFilename)}
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

// DetectPaths search for paths and load them accordingly to flags
func DetectPaths() (err error) {
	if Paths.Website == "" {
		Paths.Website, err = detectWebsitePath()
		if err != nil {
			return err
		}
	}

	if Paths.Export == defaultRelativeExportPath {
		Paths.Export = path.Join(Paths.Website, Paths.Export)
	}
	Paths.Export, err = filepath.Abs(Paths.Export)
	if err != nil {
		return err
	}
	if Paths.MetaData == defaultRelativeMetadataPath {
		Paths.MetaData = path.Join(Paths.Website, defaultRelativeMetadataPath)
	}
	Paths.MetaData, err = filepath.Abs(Paths.MetaData)
	if err != nil {
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
