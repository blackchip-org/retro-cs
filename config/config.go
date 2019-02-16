package config

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	UserHome string // the user's home directory
	RCSDir   string // use this as the home directory
	UserDir  string
	DataDir  string // data directory
	VarDir   string // Directory where runtime variable data is stored
	System   string
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Printf("unable to get home directory: %v", err)
	}
	UserHome = usr.HomeDir
}

// ResourceDir returns the root directory where RCS data can be found. Locations
// for the root directory are checked in this order: 1) The value of the
// Home variable in the configuration, 2) The value of the RCS_HOME
// environmental variable, 3) The "retro-cs" directory in the user's home
// directory.
func ResourceDir() string {
	userHome := "."
	u, err := user.Current()
	if err != nil {
		log.Printf("unable to find home directory: %v", err)
	} else {
		userHome = u.HomeDir
	}

	root := RCSDir
	if root == "" {
		root = os.Getenv("RCS_HOME")
	}
	if root == "" {
		root = filepath.Join(userHome, "rcs")
	}
	return root
}
