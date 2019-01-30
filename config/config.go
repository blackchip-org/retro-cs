package config

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	Home      string // use this as the home directory
	Monitor   bool   // enable the machine monitor
	NoVideo   bool   // disable video
	NoAudio   bool   // disable audio
	ProfC     bool   // enable CPU profiling
	Restore   bool   // restore from last export
	System    string // name of the system running
	Trace     bool   // enable tracing on start
	SlowStart bool   // skip any POST bypass
	Wait      bool   // wait for go command

	DataDir string // data directory
	ROMDir  string // ROM directory
)

// Root returns the root directory where RCS data can be found. Locations
// for the root directory are checked in this order: 1) The value of the
// Home variable in the configuration, 2) The value of the RCS_HOME
// environmental variable, 3) The "retro-cs" directory in the user's home
// directory.
func Root() string {
	userHome := "."
	u, err := user.Current()
	if err != nil {
		log.Printf("unable to find home directory: %v", err)
	} else {
		userHome = u.HomeDir
	}

	root := Home
	if root == "" {
		root = os.Getenv("RCS_HOME")
	}
	if root == "" {
		root = filepath.Join(userHome, "retro-cs")
	}
	return root
}
