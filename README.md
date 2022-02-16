# monitfiles

Compact CLI executable that uses go routines to monitor file changes and execute a script.
The sub directories will also be searched recursivelly for the file types being monitored.

### build

```bash
$ go build
```


### test

```bash
$ go test -v
```

### Commands

there is a command line to interact with the storage. 
New features will be added like simple file management to add/remove files to/from storage.

 - __help__
 - __quit__ 
 - __configs__: show all current configurations
 - __count__: number of files in store
 - __list__: list the files in store being monitored
 - __fire__: trigger the script
 - __stop__: pause monitoring
 - __start__: restart monitoring 
 - __moo__: a cow, why not


### Syntax

```bash
$ ./monitfiles 

  -b	
    	blocks script execution, waits to finish, defaults to false
  -f string
    	(shorthand for filetypes) (default "htm html css js")
  -filenames string
    	file names to be monitored for changes
  -filetypes string
    	file types to be monitored for changes (default "htm html css js")
  -i uint
    	interval in seconds for monitor changes (default 2)
  -max uint
    	max number of files to monitor (default 200)
  -n string
    	(shorthand for filenames)
  -no-dot
    	exclude (dot) dirs like .git (boolean, set to false to enable entering them) (default true)
  -none
    	file types without extension (boolean, set to true to activate)
  -p string
    	(shorthand for path)
  -path string
    	path to monitor (default ".")
  -s string
    	(shorthand for script)
  -script string
    	comand to be called upon change detection
  -v	verbose output
  -w	opens the script URL in predefined browser

```


### Example

Monitor a path for "*.md" file changes, and 2 specific files "main.go" and "main_test.go".

```bash
$ ./monitfiles -filenames "main.go main_test.go" -f md -p . -s "./scripts/say_hi.sh" -v

2022/02/16 13:25:56 * entering directory: monitfiles
2022/02/16 13:25:56    > checking .README.md.swp
2022/02/16 13:25:56    > checking .gitignore
2022/02/16 13:25:56    > checking .main.go.swp
2022/02/16 13:25:56    > checking README.md
2022/02/16 13:25:56    + adding README.md (2022-02-16 13:25:28.868812804 +0000 WET)
2022/02/16 13:25:56    > checking go.mod
2022/02/16 13:25:56    > checking go.sum
2022/02/16 13:25:56    > checking main.go
2022/02/16 13:25:56    > checking main_test.go
2022/02/16 13:25:56    > checking monitfiles
2022/02/16 13:25:56 * entering directory: scripts
2022/02/16 13:25:56    > checking brave_reload.sh
2022/02/16 13:25:56    > checking chrome_reload.sh
2022/02/16 13:25:56    > checking firefox_reload.sh
2022/02/16 13:25:56    > checking say_hi.sh
2022/02/16 13:25:56    + adding main.go (2022-02-16 13:05:13.757952126 +0000 WET)
2022/02/16 13:25:56    + adding main_test.go (2021-10-26 16:15:13.19038135 +0100 WEST)
2022/02/16 13:25:56 ************************************************
2022/02/16 13:25:56 Root path: .../scripts/monitfiles
2022/02/16 13:25:56 File types: [md]
2022/02/16 13:25:56 File names: [main.go main_test.go]
2022/02/16 13:25:56 File types with no extension ? false
2022/02/16 13:25:56 Exclude dot dirs ? true
2022/02/16 13:25:56 Max number of files: 200
2022/02/16 13:25:56 Blocking ? false
2022/02/16 13:25:56 Verbose ? true
2022/02/16 13:25:56 Interval: 2 seconds
2022/02/16 13:25:56 Script: ./scripts/say_hi.sh
2022/02/16 13:25:56 Web: false
2022/02/16 13:25:56 Number of directories scanned: 2
2022/02/16 13:25:56 Number of files added and being monitored: 3
2022/02/16 13:25:56 ************************************************


> list
 
1 (0 updates) .../scripts/monitfiles/README.md last modified at 2022-02-16 13:25:28.868812804 +0000 WET 
2 (0 updates) main.go last modified at 2022-02-16 13:05:13.757952126 +0000 WET 
3 (0 updates) main_test.go last modified at 2021-10-26 16:15:13.19038135 +0100 WEST 

 
> configs
 
configs:
   Root path: /home/ubuntu/scripts/monitfiles 
   File types: [md] 
   File names: [main.go main_test.go] 
   File types with no extension ? false 
   Exclude dot dirs ? true 
   Blocking ? false 
   Verbose ? true 
   Max number of files: 200 
   Interval: 2 seconds 
   Script: ./scripts/say_hi.sh 
   Web: false 
   Number of directories scanned: 2 
   Number of files added and being monitored: 3 
 

> moo 
 
^__^ 
(oo)\_______ 
(__)\       )\/\ 
    ||----w | 
    ||     ||
 
 
> quit
 
👋  bye!
``` 





