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

func init() {
	flag.BoolVar(&config.ProfC, "profc", false, "enable cpu profiling")
	flag.StringVar(&config.System, "s", "c64", "start this system")
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	config.DataDir = filepath.Join(config.Root(), "data")

	if config.ProfC {
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

	if config.NoVideo || config.Trace || config.Wait {
		config.Monitor = true
	}
	config.Monitor = true // always for now!

	newMachine, ok := app.Systems[config.System]
	if !ok {
		log.Fatalf("no such system: %v", config.System)
	}
	config.ROMDir = filepath.Join(config.DataDir, config.System)
	mach, err := newMachine()
	if err != nil {
		log.Fatalf("unable to create machine: \n%v", err)
	}

	var mon *app.Monitor
	if config.Monitor {
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
	if !config.Wait {
		mach.Command(rcs.MachStart{})
	}
	mach.Run()
}
