package pacman

import "github.com/blackchip-org/retro-cs/rcs"

var ROM = map[string][]rcs.ROM{
	"pacman": []rcs.ROM{
		rcs.NewROM("code    ", "pacman.6e", "e87e059c5be45753f7e9f33dff851f16d6751181"),
		rcs.NewROM("code    ", "pacman.6f", "674d3a7f00d8be5e38b1fdc208ebef5a92d38329"),
		rcs.NewROM("code    ", "pacman.6h", "8e47e8c2c4d6117d174cdac150392042d3e0a881"),
		rcs.NewROM("code    ", "pacman.6j", "d4a70d56bb01d27d094d73db8667ffb00ca69cb9"),
		rcs.NewROM("tile    ", "pacman.5e", "06ef227747a440831c9a3a613b76693d52a2f0a9"),
		rcs.NewROM("sprite  ", "pacman.5f", "4a937ac02216ea8c96477d4a15522070507fb599"),
		rcs.NewROM("color   ", "82s123.7f", "8d0268dee78e47c712202b0ec4f1f51109b1f2a5"),
		rcs.NewROM("palette ", "82s126.4a", "19097b5f60d1030f8b82d9f1d3a241f93e5c75d6"),
		rcs.NewROM("waveform", "82s126.1m", "bbcec0570aeceb582ff8238a4bc8546a23430081"),
		rcs.NewROM("waveform", "82s126.3m", "0c4d0bee858b97632411c440bea6948a74759746"),
	},
	"mspacman": []rcs.ROM{
		rcs.NewROM("code    ", "boot1    ", "bc2247ec946b639dd1f00bfc603fa157d0baaa97"),
		rcs.NewROM("code    ", "boot2    ", "13ea0c343de072508908be885e6a2a217bbb3047"),
		rcs.NewROM("code    ", "boot3    ", "5ea4d907dbb2690698db72c4e0b5be4d3e9a7786"),
		rcs.NewROM("code    ", "boot4    ", "3022a408118fa7420060e32a760aeef15b8a96cf"),
		rcs.NewROM("code2   ", "boot5    ", "fed6e9a2b210b07e7189a18574f6b8c4ec5bb49b"),
		rcs.NewROM("code2   ", "boot6    ", "387010a0c76319a1eab61b54c9bcb5c66c4b67a1"),
		rcs.NewROM("tile    ", "5e       ", "5e8b472b615f12efca3fe792410c23619f067845"),
		rcs.NewROM("sprite  ", "5f       ", "fd6a1dde780b39aea76bf1c4befa5882573c2ef4"),
		rcs.NewROM("color   ", "82s123.7f", "8d0268dee78e47c712202b0ec4f1f51109b1f2a5"),
		rcs.NewROM("palette ", "82s126.4a", "19097b5f60d1030f8b82d9f1d3a241f93e5c75d6"),
		rcs.NewROM("waveform", "82s126.1m", "bbcec0570aeceb582ff8238a4bc8546a23430081"),
		rcs.NewROM("waveform", "82s126.3m", "0c4d0bee858b97632411c440bea6948a74759746"),
	},
}
