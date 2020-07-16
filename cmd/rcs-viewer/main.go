package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	scale    int
	hscan    bool
	vscan    bool
	filename string
)

func init() {
	flag.IntVar(&scale, "scale", 1, "image `scale`")
	flag.StringVar(&config.RCSDir, "home", "", "set the RCS `home` directory")
	flag.BoolVar(&hscan, "hscan", false, "add horizontal scan lines")
	flag.BoolVar(&vscan, "vscan", false, "add vertical scan lines")
	flag.StringVar(&filename, "out", "", "output to `filename`")

	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Usage: rcs-viewer [options] <view>\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(o, "\nAvailable values for <view>:\n\n")
		list := []string{}
		for key := range views {
			list = append(list, key)
		}
		sort.Strings(list)
		fmt.Fprintln(o, strings.Join(list, "\n"))
		fmt.Fprintln(o)
	}
}

type view struct {
	system string
	roms   []rcs.ROM
	render func(*sdl.Renderer, map[string][]byte) (rcs.TileSheet, error)
}

func main() {
	log.SetFlags(0)

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	v, ok := views[flag.Arg(0)]
	if !ok {
		log.Fatalln("no such view")
	}

	var roms map[string][]byte
	if v.roms != nil {
		dir := filepath.Join(config.ResourceDir(), "data", v.system)
		r, err := rcs.LoadROMs(dir, v.roms)
		if err != nil {
			log.Fatalf("unable to load roms:\n%v\n", err)
		}
		roms = r
	}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintf(os.Stderr, "unable to initialize sdl: %v\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		flag.Arg(0),
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		100, 100,
		sdl.WINDOW_HIDDEN,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to initialize window: %v", err)
		os.Exit(1)
	}

	r, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("unable to initialize renderer: %v", err)
	}

	sheet, err := v.render(r, roms)
	if err != nil {
		log.Fatalf("unable to create sheet: %v", err)
	}
	winX, winY := sheet.TextureW*int32(scale), sheet.TextureH*int32(scale)
	window.SetSize(winX, winY)
	window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
	window.Show()

	// err = sdl.GLSetSwapInterval(1)
	// if err != nil {
	// 	fmt.Printf("unable to set swap interval: %v\n", err)
	// }

	var scanlines *sdl.Texture
	// Now that the window has been shown, the texture needs to be rerendered.
	sheet, _ = v.render(r, roms)
	slwidth := int32(scale / 2)
	if slwidth == 0 {
		slwidth = 1
	}
	if hscan {
		scanlines, err = rcs.NewScanLinesH(r, winX, winY, slwidth)
		if err != nil {
			log.Fatal(err)
		}
	}
	if vscan {
		scanlines, err = rcs.NewScanLinesV(r, winX, winY, slwidth)
		if err != nil {
			log.Fatal(err)
		}
	}

	if filename != "" {
		surf, err := sdl.CreateRGBSurface(0, winX, winY, 32, 0, 0, 0, 0)
		if err != nil {
			log.Fatal(err)
		}
		r.SetRenderTarget(nil)
		r.SetDrawColor(0, 0, 0, 0)
		r.Clear()
		r.Copy(sheet.Texture, nil, nil)

		pixels := surf.Pixels()
		ptr := unsafe.Pointer(&pixels[0])
		r.ReadPixels(nil, surf.Format.Format, ptr, int(surf.Pitch))

		if err := img.SavePNG(surf, filename); err != nil {
			log.Fatal(err)
		}
		return
	}

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if _, ok := event.(*sdl.QuitEvent); ok {
				os.Exit(0)
			}
		}

		r.SetRenderTarget(nil)
		r.SetDrawColor(0, 0, 0, 0)
		r.Clear()
		r.Copy(sheet.Texture, nil, nil)
		if scanlines != nil {
			r.Copy(scanlines, nil, nil)
		}
		sdl.Delay(250)
		r.Present()
	}

}
