// monitfiles is a small exercise cli app that uses go routines to monitor file changes and execute
// a script. The sub directories will also be searched for the file types being monitores.
// Build with:
// 	$ go build
// Execute with:
// 	$ monitfiles -f "htm html css js" -i 1 -p "/path/to/root" -s scripts/brave_reload.sh -b -v
// 	$ monitfiles --path . --filetypes "htm html css js" --script "/path/to/script" -v
// 	$ monitfiles --path . --filetypes "htm html css js" -w -s "/path/to/file"
//
// There is a comand line associated, try:
// > help
// > moo
// > list
//
// TODO: if several files change at a time dump the multiple requests with a timeout.
//
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/browser"
)

var (
	errInvalidParam  = errors.New("invalid param(s)")
	errFile          = errors.New("invalid file")
	errName          = errors.New("invalid file name")
	errPath          = errors.New("path error")
	errNotAPath      = errors.New("invalid path, please select a path not a file")
	errInvalidTypes  = errors.New("invalid file types")
	errInvalidScript = errors.New("empty script")
	errMaxFiles      = errors.New("MAX files limit reached, please consider new limit")
	errUnsuported    = errors.New("unsuported platform")
)

// Configs struct where the flags are passed into
type Configs struct {
	Path             string
	FileTypes        []string
	FileNames        []string
	FileTypeNone     bool
	ExcludeDotDirs   bool
	IncludeFileNames []string
	ExcludeFileNames []string
	Script           string
	Web              bool // opend the URL in Script in the default browser
	Blocking         bool // blocks script execution, waits to finish, defaults to false
	Verbose          bool
	MaxFiles         uint
	Interval         uint // seconds
	ScannedDirs      uint
	Files            uint
	Trigger          chan bool
}

// Store struct for each file monitoring. The State and Done channels serve to communicate with the main thread.
// The ticker guarantees independent go routine execution.
type Store struct {
	ID       uint
	Filename string
	FileType string
	Path     string
	ModTime  time.Time
	Info     os.FileInfo
	Updated  uint
	Ticker   *time.Ticker
	State    chan bool
	Done     chan bool
}

// Storage is the global slice of stores for files
type Storage []Store

func main() {
	exit := exitErrors[run(os.Args, os.Stdout)]
	os.Exit(exit.Exit)
}

func run(ags []string, stdout io.Writer) int {
	var err error

	config := &Configs{
		Path:             "",
		FileTypes:        []string{},
		FileNames:        []string{},
		FileTypeNone:     false,
		ExcludeDotDirs:   true,
		IncludeFileNames: []string{}, //TODO
		ExcludeFileNames: []string{}, //TODO
		Script:           "",
		Web:              false, // opens the URL in Script in the active browser
		Blocking:         false,
		Verbose:          false,
		MaxFiles:         0,
		Interval:         2,
		ScannedDirs:      0,
		Files:            0,
	}

	// flags
	var flagPath string
	var flagFileTypes string
	var flagFileNames string
	var flagFileTypeNone bool   // for files with no extension
	var flagExcludeDotDirs bool // exclude .dirs (dot dirs like .git)
	var flagBlocking bool
	var flagVerbose bool
	var flagMaxFiles uint
	var flagInterval uint
	var flagScript string
	var flagWeb bool
	var flagVersion bool

	flag.StringVar(&flagPath, "path", ".", "path to monitor")
	flag.StringVar(&flagPath, "p", "", "(shorthand for path)")
	flag.StringVar(&flagFileTypes, "filetypes", "htm html css js", "file types to be monitored for changes")
	flag.StringVar(&flagFileTypes, "f", "htm html css js", "(shorthand for filetypes)")
	flag.StringVar(&flagFileNames, "filenames", "", "file names to be monitored for changes")
	flag.StringVar(&flagFileNames, "n", "", "(shorthand for filenames)")
	flag.BoolVar(&flagFileTypeNone, "none", false, "file types without extension (boolean, set to true to activate)")
	flag.BoolVar(&flagExcludeDotDirs, "no-dot", true, "exclude (dot) dirs like .git (boolean, set to false to enable entering them)")
	flag.StringVar(&flagScript, "script", "", "comand to be called upon change detection")
	flag.StringVar(&flagScript, "s", "", "(shorthand for script)")
	flag.BoolVar(&flagWeb, "w", false, "opens the script URL in predefined browser")
	flag.BoolVar(&flagBlocking, "b", false, "blocks script execution, waits to finish, defaults to false")
	flag.BoolVar(&flagVerbose, "v", false, "verbose output")
	flag.UintVar(&flagMaxFiles, "max", 200, "max number of files to monitor")
	flag.UintVar(&flagInterval, "i", 2, "interval in seconds for monitor changes")
	flag.BoolVar(&flagVersion, "version", false, "version of the app")

	flag.Parse() // TODO: catch it here, kills the program flow

	if flag.NFlag() < 1 {
		flag.PrintDefaults()
		return errNoParams
	}

	config.Verbose = flagVerbose

	if flagVersion {
		fmt.Println("Version: ", string(version))
		if config.Verbose {
			fmt.Println("Release: ", string(goToolChainRev))
			fmt.Println("Compiled and built with <3 from Gophers:", string(goToolChainVer))
		}
		return errOK
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
	config.FileNames, err = validFileNames(flagFileNames)
	if err != nil {
		log.Printf("Warning, some filenames were discarded...")
	}
	config.FileTypeNone = flagFileTypeNone
	config.ExcludeDotDirs = flagExcludeDotDirs
	config.Blocking = flagBlocking
	config.MaxFiles = flagMaxFiles
	config.Interval = flagInterval
	config.Web = flagWeb
	config.Script, err = validScript(flagScript)
	if err != nil {
		log.Printf("use -h for help")
		return errScript
	}
	// channels
	config.Trigger = make(chan bool)

	storage := &Storage{}
	config.ScannedDirs, config.Files, err = storage.New(*config)
	if err != nil {
		return errStorage
	}

	// start monitoring
	for i := range *storage {
		(*storage)[i].Monitor(config)
	}

	log.Print("************************************************")
	log.Printf("Root path: %s", config.Path)
	log.Printf("File types: %s", config.FileTypes)
	log.Printf("File names: %s", config.FileNames)
	log.Printf("File types with no extension ? %t", config.FileTypeNone)
	log.Printf("Exclude dot dirs ? %t", config.ExcludeDotDirs)
	log.Printf("Max number of files: %d", config.MaxFiles)
	log.Printf("Blocking ? %t", config.Blocking)
	log.Printf("Verbose ? %t", config.Verbose)
	log.Printf("Interval: %d seconds", config.Interval)
	log.Printf("Script: %s", config.Script)
	log.Printf("Web: %t", config.Web)
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
		if line == "" {
			continue
		}
		parser(line, storage, config)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return errUnknown
}

// parser is the function parses and interprets the command
func parser(cmd string, storage *Storage, config *Configs) {
	fmt.Println("")
	switch cmd {
	case "quit":
		for i := range *storage {
			(*storage)[i].Done <- true
		}
		if config.Verbose {
			fmt.Println("\U0001F44B  bye!")
		}
		os.Exit(1) // TODO: replace with return a successful exitError
	case "?", "help", "h":
		fmt.Println("available commands: quit help moo count list fire configs start stop")
	case "moo":
		fmt.Println("^__^ \n(oo)\\_______ \n(__)\\       )\\/\\ \n    ||----w | \n    ||     ||\n ")
	case "count":
		fmt.Printf("%d files on store \n", len(*storage))
	case "configs":
		fmt.Println("configs:")
		fmt.Printf("   Root path: %s \n", config.Path)
		fmt.Printf("   File types: %s \n", config.FileTypes)
		fmt.Printf("   File names: %s \n", config.FileNames)
		fmt.Printf("   File types with no extension ? %t \n", config.FileTypeNone)
		fmt.Printf("   Exclude dot dirs ? %t \n", config.ExcludeDotDirs)
		fmt.Printf("   Blocking ? %t \n", config.Blocking)
		fmt.Printf("   Verbose ? %t \n", config.Verbose)
		fmt.Printf("   Max number of files: %d \n", config.MaxFiles)
		fmt.Printf("   Interval: %d seconds \n", config.Interval)
		fmt.Printf("   Script: %s \n", config.Script)
		fmt.Printf("   Web: %t \n", config.Web)
		fmt.Printf("   Number of directories scanned: %d \n", config.ScannedDirs)
		fmt.Printf("   Number of files added and being monitored: %d \n", config.Files)
	case "list":
		for _, s := range *storage {
			fmt.Printf("%d (%d updates) %s last modified at %v \n", s.ID, s.Updated, s.Path, s.ModTime)
		}
	case "fire":
		// exec script
		Exec(config)
	case "start":
		for i := range *storage {
			(*storage)[i].State <- true
		}
		if config.Verbose {
			log.Printf("+++ monitoring %d files at interval %d seconds \n", config.Files, config.Interval)
		}
	case "stop":
		for i := range *storage {
			(*storage)[i].State <- false
		}
		if config.Verbose {
			log.Printf("+++ stopped monitoring %d files \n", config.Files)
		}
	case "debug":
	default:
		if config.Verbose {
			fmt.Println("unknown command...")
		}

	}
	fmt.Print("> ")
}

// Exec the script
func Exec(config *Configs) {
	var out []byte
	var err error

	if config.Verbose {
		log.Printf("Script run: %s\n", config.Script)
	}
	if config.Web {
		// if -w flag and the script contains an URL to be opened in a browser
		err = browser.OpenURL(config.Script)
	} else {
		// script execution
		cmd := exec.Command(config.Script)

		if config.Blocking {
			out, err = cmd.Output()
			if config.Verbose {
				log.Printf("Output: %s \n", out)

			}
		} else {
			err = cmd.Start()
			if err == nil {
				err = cmd.Wait()
			}
		}
	}

	if config.Verbose && err != nil {
		log.Printf("Script error: %v\n", err)
	}
}

// Monitor a file for file changes every interval.
// On each tick, file changes are checked. If file check returns error a log warning is generated only.
// Upon Done, goroutines is returned, channels are closed.
// State channel activates/deactivates the monitoring action, yet trigger continues.
func (s *Store) Monitor(config *Configs) {

	go func(s *Store) {

		defer close(s.Done)
		defer s.Ticker.Stop()

		var state = true

		for {
			select {
			case <-s.Done:
				return
			case state = <-s.State:
			case <-s.Ticker.C:
				f, err := os.Stat(s.Path)
				if err != nil {
					if config.Verbose {
						log.Printf("file check error for (%d) %s (%s)", s.ID, s.Filename, err)
					}
				} else {
					if state && f.ModTime() != s.ModTime {
						if config.Verbose {
							log.Printf(" +++ file change: (%d) %s", s.ID, s.Filename)
						}
						// update the record with new information
						s.ModTime = f.ModTime()
						s.Info = f
						s.Updated++
						// exec script
						Exec(config)

					}
				}
			}
		}
	}(s)

}

// New storage preloads all files in the storage structure.
// Returns:
// - number of scanned dirs
// - number of added files
// - error
// A ticker channel is added so that there's independence per store.
func (s *Storage) New(config Configs) (uint, uint, error) {
	var nd, nf uint = 0, 0

	// per dir
	err := filepath.Walk(config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if nf > config.MaxFiles {
			return errMaxFiles
		}
		if info.IsDir() {
			// directory exclusion
			if (info.Name()[0:1] == ".") && config.ExcludeDotDirs {
				return filepath.SkipDir
			} else {
				nd++
				if config.Verbose {
					log.Printf("* entering directory: %s", info.Name())
				}
			}
		}
		// file picking
		if !info.IsDir() {
			if config.Verbose {
				log.Printf("   > checking %s", info.Name())
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			// check extension size and if allowed
			extSize := len(filepath.Ext(info.Name()))
			if extSize == 0 && !config.FileTypeNone {
				return nil
			}
			// check if extension inside slice of valid ones
			ext := ""
			if extSize != 0 {
				ext = filepath.Ext(info.Name())[1:]
			}
			i := sort.SearchStrings(config.FileTypes, ext)
			caseExtListed := extSize > 0 && i < len(config.FileTypes) && config.FileTypes[i] == ext
			caseNoExt := extSize == 0 && config.FileTypeNone
			if caseExtListed || caseNoExt {
				// add the file
				if config.Verbose {
					log.Printf("   + adding %s (%v)", info.Name(), info.ModTime())
				}
				nf++
				f := Store{
					ID:       nf,
					Filename: info.Name(),
					FileType: ext,
					Path:     path,
					ModTime:  info.ModTime(),
					Updated:  0,
					Info:     info,
					Done:     make(chan bool),
					State:    make(chan bool),
					Ticker:   time.NewTicker(time.Duration(config.Interval) * time.Second),
				}
				*s = append(*s, f)
			}
		}
		return nil
	})

	// per name (-filenames)
	for _, n := range config.FileNames {
		info, err := os.Stat(n)
		if err != nil {
			log.Printf("## %s %s %v", n, errName, err)
			continue
		}
		// add the file
		if config.Verbose {
			log.Printf("   + adding %s (%v)", info.Name(), info.ModTime())
		}
		nf++
		f := Store{
			ID:       nf,
			Filename: info.Name(),
			FileType: filepath.Ext(info.Name())[1:],
			Path:     n,
			ModTime:  info.ModTime(),
			Updated:  0,
			Info:     info,
			Done:     make(chan bool),
			State:    make(chan bool),
			Ticker:   time.NewTicker(time.Duration(config.Interval) * time.Second),
		}
		*s = append(*s, f)
	}

	return nd, nf, err
}

// validPath check the existance of a path and converts to Absolute path.
func validPath(p string) (string, error) {

	stat, err := os.Stat(p)
	if err != nil {
		return "", errPath
	}

	if !stat.IsDir() {
		return "", errNotAPath
	}

	res, err := filepath.Abs(p)
	if err != nil {
		return "", errPath
	}

	return res, err
}

// validFileTypes checks if not empty, creates a slice of file types and converts to lower case
// saving n convertions that way later. Slice is sorted to ease comparisions later.
func validFileTypes(ft string) ([]string, error) {
	var err error

	res := strings.Fields(strings.ToLower(ft))
	if len(res) < 1 {
		return res, errInvalidTypes
	}

	sort.StringSlice(res).Sort()
	return res, err
}

// validFileNames checks filenames and removes invalid ones.
func validFileNames(fn string) ([]string, error) {
	var err error

	names := strings.Fields(fn)
	if len(names) < 1 {
		return nil, nil
	}

	res := make([]string, len(names))

	for i, n := range names {
		if _, err := os.Open(n); err == nil {
			res[i] = n
		}
	}

	sort.StringSlice(res).Sort()
	return res, err
}

func validScript(s string) (string, error) {
	var err error

	if len(s) == 0 {
		return "", errInvalidScript
	}

	return s, err
}
