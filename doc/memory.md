# memory

If it was only as easy as...

```go
var mem [65536]uint8
```

Sometime around the year 2009, I embarked on writing my first emulator. I did not implement memory as an array because I knew that the Commodore 64 used banked memory. The 6510 processor has 16 address lines and can, therefore, access up to 64K of different memory locations. The computer has 64K of RAM available, but all of that RAM might not be visible depending on the bank that is selected.

Powering on the Commodore 64 with no cartridges plugged in lands the user at a "READY" prompt where it patiently waits for BASIC commands. The ROM that holds the code for the BASIC interpreter needs to be accessible to the CPU somewhere. In this case it is mapped to the address range starting with 0xa000 and ending with 0xbfff. While the BASIC ROM is banked in, the RAM at these addresses cannot be read.

In that first emulator, I came up with an abstraction to memory that was overly complicated and horribly inefficient. The goal of that emulator was to hack up an implementation of a 6502 series processor and to put some fun fluff on top. Emulating an actual Commodore was never in the plan. I'm not sure why I ever thought I would need banked memory. I guess I was in more fear of the 64K boundary than I was of the cringe-worthy code that I was writing. I lost interest in that project while there was still wide expanses of empty memory. I should have used an array.

Early in the year 2018, I started work on this emulator which was planned from the start to emulate the Commodore 64 and only the Commodore 64. This time banked memory would be necessary. And once again, I came up with an abstraction that was overly complicated. Memory was divided into seven different regions and various "chunks" could be mapped to a region depending on the bank that was selected.

It ignored the fact that a lot of the addresses in the IO region are mapped to registers on various chips. When the IO region is mapped in, the memory address 0xd000 does not point to RAM or ROM at all. It reads or writes to a register on the Video Interface Chip (VIC-II) and this register controls the X coordinate for sprite #0. In this emulator the IO region was created as a RAM chunk and external chips had to use this RAM for their registers. Not great but I at least got the emulator to boot to a "READY" prompt with this technique.

The Pac-Man emulator came next. I couldn't use the memory with the seven regions from the Commodore emulator because that didn't apply here. It was too Commodore 64 specific. I decided on blocks of memory that could be mapped at page boundaries. It seemed like a good idea at the time. Everything seemed to line up at page boundaries.

I should have read the memory map instead of skimming it.

When I finally had the parts of the Pac-Man emulator ready for assembly I realized this page boundary scheme might be a problem. There were certain addresses that would write to one value but read from another. For example, a write to 0x5000 enables or disables interrupts but a read returns the value of port IN0--inputs for joysticks, coin slots, etc. By the way, IN0 can also be read at 0x5001, 0x5002, and every address up to 0x503f. So much for being exact.

At this point, I had a common interface that memory types implemented:

```go
type Memory interface {
    Read(addr int) uint8
    Write(addr int, val uint8)
}
```
There was RAM which was backed by a byte array, ROM which was also backed by a byte array that ignored writes, null memory that did nothing, page mapped memory, and memory that would spy on reads and writes. I introduced another memory, IO, that had two values per address, one for reading and one for writing. And instead of being actual values, they were pointers to values.

Galaga was next and the CPU now had to interface with two different chips,
the N51XX and N54XX, but accesses to these chips were through another chip, the N06XX. A write to an address could go to either the N51XX or the N54XX depending on a value set in the N06XX. What was really needed was neither a value, nor a pointer to a value, but a function. I wanted to introduce another memory type that had two functions per address, one for reading and one for writing. This was starting to get complicated. Again.

How about making all memory addresses have a read and write function and scrap all the other memory types? Refactoring memory started the grand refactor by combining the Commodore 64 and Pac-Man/Galaga code together to make the Retro-CS.

Will this be the [final memory scheme](https://godoc.org/github.com/blackchip-org/retro-cs/rcs#Memory)? Who knows. But I'm good at making things overly complicated.

## Retrospection

Factors to consider when emulating memory:

- An address might point to different values depending on a selected bank
- An address might be mapped to a register on an external chip
- An address value might map to two values--one for reading, one for writing.
- A value might map to multiple addresses.
- A read or write to an address value might "do something"

Use functions.





