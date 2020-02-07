package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidParam  = errors.New("invalid param(s)")
	ErrFile          = errors.New("invalid file")
	ErrPath          = errors.New("path error")
	ErrNotAPath      = errors.New("invalid path, please select a path not a file")
	ErrInvalidTypes  = errors.New("invalid file types")
	ErrInvalidScript = errors.New("empty script")
	ErrMaxFiles      = errors.New("MAX files limit reached, please consider new limit")
)

// Configs struct
type Configs struct {
	Path             string
	FileTypes        []string
	FileTypeNone     bool
	ExcludeDotDirs   bool
	IncludeFileNames []string
	ExcludeFileNames []string
	Script           string
	MaxFiles         uint
	Interval         uint // seconds
	ScannedDirs      uint
	Files            uint
}

// Store struct for each file monitoring
type Store struct {
	ID       uint
	Filename string
	FileType string
	Path     string
	ModTime  time.Time
	Info     os.FileInfo
}

// Storage is the global slice of stores for files
type Storage []Store

// Execution:
// 	$ reloadtab --path . --filetypes htm html css js --script "osascript -e 'tell application "Brave" to tell the active tab of its first window to reload'"
func main() {

	var err error

	config := &Configs{
		Path:             "",
		FileTypes:        []string{},
		FileTypeNone:     false,
		ExcludeDotDirs:   true,
		IncludeFileNames: []string{}, //TODO
		ExcludeFileNames: []string{}, //TODO
		Script:           "",
		MaxFiles:         0,
		Interval:         2,
		ScannedDirs:      0,
		Files:            0,
	}

	// flags
	var flagPath string
	var flagFileTypes string
	var flagFileTypeNone bool   // for files with no extension
	var flagExcludeDotDirs bool // exclude .dirs (dot dirs like .git)
	var flagMaxFiles uint
	var flagInterval uint
	var flagScript string

	flag.StringVar(&flagPath, "path", ".", "path to monitor")
	flag.StringVar(&flagPath, "p", "", "(shorthand for path)")
	flag.StringVar(&flagFileTypes, "filetypes", "htm html css js", "file types to be monitored for changes")
	flag.StringVar(&flagFileTypes, "f", "htm html css js", "(shorthand for filetypes)")
	flag.BoolVar(&flagFileTypeNone, "none", false, "file types without extension (boolean, set to true to activate)")
	flag.BoolVar(&flagExcludeDotDirs, "no-dot", true, "exclude (dot) dirs like .git (boolean, set to false to enable entering them)")
	flag.StringVar(&flagScript, "script", "", "comand to be called upon change detection")
	flag.StringVar(&flagScript, "s", "", "(shorthand for script)")
	flag.UintVar(&flagMaxFiles, "max", 200, "max number of files to monitor")
	flag.UintVar(&flagInterval, "i", 2, "interval in seconds for monitor changes")

	flag.Parse()

	if flag.NFlag() < 1 {
		flag.PrintDefaults()
		log.Fatal("please specify params")
	}

	config.Path, err = validPath(flagPath)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}
	config.FileTypes, err = validFileTypes(flagFileTypes)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}
	config.FileTypeNone = flagFileTypeNone
	config.ExcludeDotDirs = flagExcludeDotDirs
	config.MaxFiles = flagMaxFiles
	config.Interval = flagInterval
	config.Script, err = validScript(flagScript)
	if err != nil {
		log.Printf("use -h for help")
		log.Fatalf("*** Error: %s", err)
	}

	storage := Storage{}
	config.ScannedDirs, config.Files, err = storage.New(*config)
	if err != nil {
		log.Fatalf("*** Error: %s", err)
	}

	log.Print("************************************************")
	log.Printf("Root path: %s", config.Path)
	log.Printf("File types: %s", config.FileTypes)
	log.Printf("File types with no extension ? %t", config.FileTypeNone)
	log.Printf("Exclude dot dirs ? %t", config.ExcludeDotDirs)
	log.Printf("Max number of files: %d", config.MaxFiles)
	log.Printf("Interval: %d seconds", config.Interval)
	log.Printf("Script: %s", config.Script)
	log.Printf("Number of directories scanned: %d", config.ScannedDirs)
	log.Printf("Number of files added and being monitored: %d", config.Files)
	log.Print("************************************************")
	fmt.Println()
	fmt.Print("> ")

	// user interface
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Print("> ")
		line := scanner.Text()
		parser(line, storage, *config)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

// parser is the function parses and interprets the command
func parser(cmd string, storage Storage, config Configs) {
	fmt.Println("")
	switch cmd {
	case "quit":
		os.Exit(1)
	case "?", "help", "h":
		fmt.Println("available commands: quit help moo count list configs start stop")
	case "moo":
		fmt.Println("^__^ \n(oo)\\_______ \n(__)\\       )\\/\\ \n    ||----w | \n    ||     ||")
	case "count":
		fmt.Printf("%d files on store \n", len(storage))
	case "configs":
		fmt.Println("configs:")
		fmt.Printf("   Root path: %s \n", config.Path)
		fmt.Printf("   File types: %s \n", config.FileTypes)
		fmt.Printf("   File types with no extension ? %t \n", config.FileTypeNone)
		fmt.Printf("   Exclude dot dirs ? %t \n", config.ExcludeDotDirs)
		fmt.Printf("   Max number of files: %d \n", config.MaxFiles)
		fmt.Printf("   Interval: %d seconds \n", config.Interval)
		fmt.Printf("   Script: %s \n", config.Script)
		fmt.Printf("   Number of directories scanned: %d \n", config.ScannedDirs)
		fmt.Printf("   Number of files added and being monitored: %d \n", config.Files)
	case "list":
		for _, s := range storage {
			fmt.Printf("%d %s last modified at %v \n", s.ID, s.Path, s.ModTime)
		}
	case "start":
		// start monitoring
		// WIP
		fmt.Printf("ok, monitoring %d files at interval %d seconds \n", config.Files, config.Interval)
	case "stop":
		// stop monitoring
		// WIP
		fmt.Printf("ok, stopped monitoring %d files \n", config.Files)
	default:
		fmt.Println("unknown command...")

	}
	fmt.Print("> ")
}

// New storage preloads all files in the storage structure.
// Returns:
// - number of scanned dirs
// - number of added files
// - error
func (s *Storage) New(config Configs) (uint, uint, error) {
	var nd, nf uint = 0, 0

	err := filepath.Walk(config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if nf > config.MaxFiles {
			return ErrMaxFiles
		}
		if info.IsDir() {
			// directory exclusion
			if (info.Name()[0:1] == ".") && config.ExcludeDotDirs {
				return filepath.SkipDir
			} else {
				nd += 1
				log.Printf("* entering directory: %s", info.Name())
			}
		}
		// file picking
		if !info.IsDir() {
			log.Printf("   > checking %s", info.Name())
			if !info.Mode().IsRegular() {
				return filepath.SkipDir
			}
			// check extension size and if allowed
			ext_size := len(filepath.Ext(info.Name()))
			if ext_size == 0 && !config.FileTypeNone {
				return filepath.SkipDir
			}
			// check if extension inside slice of valid ones
			ext := filepath.Ext(info.Name())[1:]
			i := sort.SearchStrings(config.FileTypes, ext)
			if ext_size > 0 && i < len(config.FileTypes) && config.FileTypes[i] == ext {
				// add the file
				log.Printf("   + adding %s (%v)", info.Name(), info.ModTime())
				nf += 1
				f := Store{
					ID:       nf,
					Filename: info.Name(),
					FileType: ext,
					Path:     path,
					ModTime:  info.ModTime(),
					Info:     info,
				}
				*s = append(*s, f)
			}
		}
		return nil
	})

	return nd, nf, err
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
