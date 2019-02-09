package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/blackchip-org/retro-cs/app"
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
)

const (
	defaultWidth  = 1024
	defaultHeight = 786
)

var (
	optProfC   bool
	optSystem  string
	optMonitor bool
	optNoAudio bool
	optNoVideo bool
	optTrace   bool
	optWait    bool
)

func init() {
	flag.BoolVar(&optProfC, "profc", false, "enable cpu profiling")
	flag.BoolVar(&optNoAudio, "no-audio", false, "disable audio")
	flag.BoolVar(&optNoVideo, "no-video", false, "disable video")
	flag.BoolVar(&optMonitor, "m", false, "enable monitor")
	flag.StringVar(&optSystem, "s", "c64", "start this system")
	flag.BoolVar(&optTrace, "t", false, "enable tracing")
	flag.BoolVar(&optWait, "w", false, "wait for go command")
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	config.DataDir = filepath.Join(config.ResourceDir(), "data")

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
	if optNoVideo || optTrace || optWait {
		optMonitor = true
	}

	newMachine, ok := app.Systems[optSystem]
	if !ok {
		log.Fatalf("no such system: %v", optSystem)
	}
	config.ROMDir = filepath.Join(config.DataDir, optSystem)
	config.VarDir = filepath.Join(config.ResourceDir(), "var", optSystem)

	if err := os.MkdirAll(config.VarDir, 0755); err != nil {
		log.Fatalf("unable to create directory %v: %v", config.VarDir, err)
	}

	ctx := rcs.SDLContext{}
	if !optNoVideo {
		if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
			log.Fatalf("unable to initialize video: %v", err)
		}
		fullScreen := uint32(0)
		if !optMonitor {
			fullScreen = sdl.WINDOW_FULLSCREEN_DESKTOP
		}
		window, err := sdl.CreateWindow(
			"retro-cs",
			sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			defaultWidth, defaultHeight,
			sdl.WINDOW_SHOWN|fullScreen,
		)
		if err != nil {
			log.Fatalf("unable to initialize window: %v", err)
		}

		r, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
		if err != nil {
			log.Fatalf("unable to initialize renderer: %v", err)
		}
		if err = sdl.GLSetSwapInterval(1); err != nil {
			log.Printf("unable to set swap interval: %v", err)
		}
		ctx.Window = window
		ctx.Renderer = r
	}

	if !optNoAudio {
		requestSpec := sdl.AudioSpec{
			Freq:     22050,
			Format:   sdl.AUDIO_S16LSB,
			Channels: 2,
			Samples:  367,
		}
		if err := sdl.OpenAudio(&requestSpec, &ctx.AudioSpec); err != nil {
			log.Fatalf("unable to initialize audio: %v", err)
		}
		sdl.PauseAudio(false)
	}

	mach, err := newMachine(ctx)
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
	if optTrace {
		mach.Command(rcs.MachTrace)
	}
	mach.Run()
}
