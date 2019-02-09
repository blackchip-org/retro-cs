# pacman

Almost everything you need to know to write a Pac-Man Emulator can be found in Chris Lomont's [Pac-Man Emulation Guide](https://www.lomont.org/Software/Games/PacMan/PacmanEmulation.pdf). This document tries to fill in some of the areas that were not covered.

## Z80
A Z80 implementation that passes the [Zexdoc](http://mdfs.net/Software/Z80/Exerciser/) tests is sufficient. There is no need to add the undocumented instructions.

Accurate timing of the CPU is not necessary. This emulator runs 1,000 instructions per VBLANK and that number seems to work well.

## Memory
The source code in [MAME](https://github.com/mamedev/mame/blob/master/src/mame/drivers/pacman.cpp) notes that the most signfigant line in the addres bus (A15) is not attached. If this is not emulated, the attract screen will be missing the text for "High Score" and "Credits". This may be a copy protection feature.

When Pac-Man starts, it performs a series of initialization tasks and then executes a halt instruction to wait for the first interrupt. The stack pointer has not been set at this point and the interrupt will push the return address to 0xffff and 0xfffe which doesn't get used. Ms. Pac-Man also writes to 0xfffd and 0xfffc.

## Video
While there are only 64 palettes, color memory does contain garbage in the higher bits. Mask out the value by and-ing with 0x3f.

Figure 7 and Figure 8 which show the screen layout is a little difficult to read. This is a smaller table which shows the address values at the edges.

```
3df 3de | 3dd 3dc ... 3c3 3c2 | 3c1 3c0
3ff 3fe | 3fd 3fc ... 3e3 3e2 | 3e1 3e0
---------------------------------------
        | 3a0 380     060 040 |
        | 3a1 381     061 041 |
        | ... ...     ... ... |
        | 3be 39e     07e 05e |
        | 3bf 39f     07f 05f |
---------------------------------------
01f 01e | 01d 01c ... 003 002 | 001 000
03f 03e | 03d 03c ... 023 022 | 021 020
```

The layout of the tiles and sprites is confusing and easy to get wrong. There may also be a bug in the documentation. If those instructions do not produce correct images, the matrices below can be used instead. To fill the pixel in the target image at (X, Y) use the value found in matrix where a value of 0 is the first bit-plane in byte 0, 1 is the second bit-plane in byte 0, 4 is the first bit plane in byte 1, etc.

### Tiles
```
     0   1   2   3   4   5   6   7
     --  --  --  --  --  --  --  --
 0 | 63, 59, 55, 51, 47, 43, 39, 35
 1 | 62, 58, 54, 50, 46, 42, 38, 34
 2 | 61, 57, 53, 49, 45, 41, 37, 33
 3 | 60, 56, 52, 48, 44, 40, 36, 32
 4 | 31, 27, 23, 19, 15, 11,  7,  3
 5 | 30, 26, 22, 18, 14, 10,  6,  2
 6 | 29, 25, 21, 17, 13,  9,  5,  1
 7 | 28, 24, 20, 16, 12,  8,  4,  0
```
### Sprites
```
     0    1    2    3    4    5    6    7    8    9    10   11   12   13   14   15
     ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---  ---
 0 | 191, 187, 183, 179, 175, 171, 167, 163,  63,  59,  55,  51,  47,  43,  39,  35
 1 | 190, 186, 182, 178, 174, 170, 166, 162,  62,  58,  54,  50,  46,  42,  38,  34
 2 | 189, 185, 181, 177, 173, 169, 165, 161,  61,  57,  53,  49,  45,  41,  37,  33
 3 | 188, 184, 180, 176, 172, 168, 164, 160,  60,  56,  52,  48,  44,  40,  36,  32
 4 | 223, 219, 215, 211, 207, 203, 199, 195,  95,  91,  87,  83,  79,  75,  71,  67
 5 | 222, 218, 214, 210, 206, 202, 198, 194,  94,  90,  86,  82,  78,  74,  70,  66
 6 | 221, 217, 213, 209, 205, 201, 197, 193,  93,  89,  85,  81,  77,  73,  69,  65
 7 | 220, 216, 212, 208, 204, 200, 196, 192,  92,  88,  84,  80,  76,  72,  68,  64
 8 | 255, 251, 247, 243, 239, 235, 231, 227, 127, 123, 119, 115, 111, 107, 103,  99
 9 | 254, 250, 246, 242, 238, 234, 230, 226, 126, 122, 118, 114, 110, 106, 102,  98
10 | 253, 249, 245, 241, 237, 233, 229, 225, 125, 121, 117, 113, 109, 105, 101,  97
11 | 252, 248, 244, 240, 236, 232, 228, 224, 124, 120, 116, 112, 108, 104, 100,  96
12 | 159, 155, 151, 147, 143, 139, 135, 131,  31,  27,  23,  19,  15,  11,   7,   3
13 | 158, 154, 150, 146, 142, 138, 134, 130,  30,  26,  22,  18,  14,  10,   6,   2
14 | 157, 153, 149, 145, 141, 137, 133, 129,  29,  25,  21,  17,  13,   9,   5,   1
15 | 156, 152, 148, 144, 140, 136, 132, 128,  28,  24,  20,  16,  12,   8,   4,   0
```

## Registers

When starting the machine, if the inputs have not been hooked up yet (joysticks, buttons, coin slots), the IN0 and IN1 registers should be set to sane values.

- IN0: 0xbf
- IN1: 0xff

Leaving these values as zero will boot to the testing screen instead of the game. If set improperly, the demo game in attract mode will crash and end up showing one of the level transition animations.

Attract mode does not show up when free play is enabled.

## Ms. Pac-Man
Load the additional code ROM at address 0x8000 and it should be good to go!

## References
- Lomont, Chris, "Pac-Man Emulation Guide v0.1, Oct 2008", https://www.lomont.org/Software/Games/PacMan/PacmanEmulation.pdf
- Longstaff-Tyrrell, Mark, "Pac-Man Machine Emulator", http://www.frisnit.com/pac-man-machine-emulator
- Salmoria, Nicola, et al, "Namco PuckMan", https://github.com/mamedev/mame/blob/master/src/mame/drivers/pacman.cpp
