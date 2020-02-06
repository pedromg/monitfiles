package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// MAX_FILES sets a limit for max number of files to monitor
	MAX_FILES = 100
)

var (
	ErrInvalidParam  = errors.New("invalid param(s)")
	ErrFile          = errors.New("invalid file")
	ErrPath          = errors.New("path error")
	ErrNotAPath      = errors.New("invalid path, please select a path not a file")
	ErrInvalidTypes  = errors.New("invalid file types")
	ErrInvalidScript = errors.New("empty script")
)

type Configs struct {
	Path string
	FileTypes []string


}

type Store struct {
	Filename string
	FileType string
	Info     os.FileInfo
}

type Storage []Store

// Execution:
// 	$ reloadtab --path . --filetypes htm html css js --script "osascript -e 'tell application "Brave" to tell the active tab of its first window to reload'"
func main() {

	// flags
	var flagPath string
	var flagFileTypes string
	var flagFileTypesZero bool  // for files with no extension
	var flagExcludeDotDirs bool // exclude .dirs (dot dirs like .git)
	var flagScript string

	flag.StringVar(&flagPath, "path", ".", "path to monitor")
	flag.StringVar(&flagPath, "p", "", "(shorthand for path)")
	flag.StringVar(&flagFileTypes, "filetypes", "htm html css js", "file types to be monitored for changes")
	flag.StringVar(&flagFileTypes, "f", "htm html css js", "(shorthand for filetypes)")
	flag.BoolVar(&flagFileTypesZero, "z", false, "file types without extension (boolean, set to true to activate)")
	flag.BoolVar(&flagExcludeDotDirs, "z", true, "exclude (dot) dirs like .git (boolean, set to false to enable entering them)")
	flag.StringVar(&flagScript, "script", "", "comand to be called upon change detection")
	flag.StringVar(&flagScript, "s", "", "(shorthand for script)")

	flag.Parse()

	if flag.NFlag() < 1 {
		flag.PrintDefaults()
		log.Fatal("please specify params")
	}

	path, err := validPath(flagPath)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}
	fileTypes, err := validFileTypes(flagFileTypes)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}
	script, err := validScript(flagScript)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}

	var storage Storage
	err = storage.New(path, fileTypes)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}

	log.Printf("Root: %s", path)
	log.Printf("File Types: %s", fileTypes)
	log.Printf("Script: %s", script)

	err = storage.Preload(path, fileTypes, flagFileTypesZero)
	if err != nil {
		log.Fatalf("*** Error: %s", err)
	}

}

// Preload load all files in the storage structure.
// Parameters are:
// - initial root
// - file type extensions
// - include files with no extension
// TODO: add a flag for specific filenames
func (s Storage) Preload(p string, ft []string, z bool) error {
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() &&  {
			log.Printf("* entering directory: %s", info.Name())
		}
		if !info.IsDir() {
			log.Printf("   > checking %s", info.Name())
			if sort.SearchStrings(ft, info.Name()[1:]) < len(ft) {
				log.Printf("   + adding %s (%v)", info.Name(), info.ModTime())
				// TODO Add the file to the slice
			}
		}
		return nil
	})

	return err
}

// New storage for the specified path and files.
// Will filter specified file types. Makes us of https://golang.org/pkg/path/filepath/#Walk
func (s Storage) New(root string, fileTypes []string) error {
	var err error

	return err
}

// validPath check the existance of a path and converts to Absolute path.
func validPath(p string) (string, error) {

	stat, err := os.Stat(p)
	if err != nil {
		return "", ErrPath
	}

	if !stat.IsDir() {
		return "", ErrNotAPath
	}

	res, err := filepath.Abs(p)
	if err != nil {
		return "", ErrPath
	}

	return res, err
}

// validFileTypes checks if not empty, creates a slice of file types and converts to lower case
// saving n convertions that way later. Slice is sorted to ease comparisions later.
func validFileTypes(ft string) ([]string, error) {
	var err error

	res := strings.Fields(strings.ToLower(ft))
	if len(res) < 1 {
		return res, ErrInvalidTypes
	}

	sort.StringSlice(res).Sort()
	return res, err
}

func validScript(s string) (string, error) {
	var err error

	if len(s) == 0 {
		return "", ErrInvalidScript
	}

	return s, err
}
