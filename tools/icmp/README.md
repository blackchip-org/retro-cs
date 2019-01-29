# icmp

Instruction comparision tool.

This tool was useful to find the source of errors in the z80 instruction code. It uses the z80emu code written by Lin Ke-Fong:

https://github.com/anotherlin/z80emu

There are two code wrappers, one for z80emu and one for RCS, to print a text file of all inputs and outputs of an instruction. Both programs are run and the text files are compared for differences.

The code is hard-wired to a specific instruction and is modified as needed. To run the test:

```
make
```

The z80emu code must be available to compile. By default the Makefile uses `$(HOME)/z80emu` but can be overriden by setting the Z80EMU_HOME variable.
