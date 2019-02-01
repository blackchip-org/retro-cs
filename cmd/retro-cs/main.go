package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/blackchip-org/retro-cs/app"
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
)

var (
	optProfC   bool
	optWait    bool
	optSystem  string
	optMonitor bool
)

func init() {
	flag.BoolVar(&optProfC, "profc", false, "enable cpu profiling")
	flag.StringVar(&optSystem, "s", "c64", "start this system")
	flag.BoolVar(&optWait, "w", false, "wait for go command")
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	config.DataDir = filepath.Join(config.Root(), "data")

	if optProfC {
		f, err := os.Create("./cpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		log.Println("starting profile")
		defer func() {
			pprof.StopCPUProfile()
			log.Println("profile saved")
		}()
	}

	optMonitor = true // always for now!

	newMachine, ok := app.Systems[optSystem]
	if !ok {
		log.Fatalf("no such system: %v", optSystem)
	}
	config.ROMDir = filepath.Join(config.DataDir, optSystem)
	mach, err := newMachine()
	if err != nil {
		log.Fatalf("unable to create machine: \n%v", err)
	}

	var mon *app.Monitor
	if optMonitor {
		mon = app.NewMonitor(mach)
		defer func() {
			mon.Close()
		}()
		go func() {
			err := mon.Run()
			if err != nil {
				log.Fatalf("monitor error: %v", err)
			}
		}()
	}
	mach.Status = rcs.Run
	if optWait {
		mach.Status = rcs.Pause
	}
	mach.Run()
}
