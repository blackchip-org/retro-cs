# fuse

Z80 tests designed for the Free Unix Spectrum Emulator (FUSE).

Source files used to generate the tests are not found in this repository. Download and place in the following locations, relative to the repository root:

- ext/fuse/tests.expected
- ext/fuse/tests.in

The original location of FUSE is here:

- http://fuse-emulator.sourceforge.net/

The files found in the resource pack were downloaded from:

- https://github.com/descarte1/fuse-emulator-fuse/tree/fuse-1-3-6/z80/tests

Generate `in.go` and `expected.go` with:

```bash
go generate
```