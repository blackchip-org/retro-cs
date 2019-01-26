package galaga

import "github.com/blackchip-org/retro-cs/rcs"

var ROM = map[string][]rcs.ROM{
	"galaga": []rcs.ROM{
		rcs.NewROM("code1  ", "04m_g01.bin", "6907773db7c002ecde5e41853603d53387c5c7cd"),
		rcs.NewROM("code1  ", "04k_g02.bin", "666975aed5ce84f09794c54b550d64d95ab311f0"),
		rcs.NewROM("code1  ", "04j_g03.bin", "481f443aea3ed3504ec2f3a6bfcf3cd47e2f8f81"),
		rcs.NewROM("code1  ", "04h_g04.bin", "366cb0dbd31b787e64f88d182108b670d03b393e"),
		rcs.NewROM("code2  ", "04e_g05.bin", "d29b68d6aab3217fa2106b3507b9273ff3f927bf"),
		rcs.NewROM("code3  ", "04d_g06.bin", "d6cb439de0718826d1a0363c9d77de8740b18ecf"),
		rcs.NewROM("tile   ", "07m_g08.bin", "62f1279a784ab2f8218c4137c7accda00e6a3490"),
		rcs.NewROM("sprite ", "07e_g10.bin", "e697c180178cabd1d32483c5d8889a40633f7857"),
		rcs.NewROM("sprite ", "07h_g09.bin", "c340ed8c25e0979629a9a1730edc762bd72d0cff"),
		rcs.NewROM("palette", "5n.bin     ", "1a6dea13b4af155d9cb5b999a75d4f1eb9c71346"),
		rcs.NewROM("color  ", "2n.bin     ", "7323084320bb61ae1530d916f5edd8835d4d2461"),
	},
}
