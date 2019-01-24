# zex

The Z80 Instruction Exerciser written by Frank D. Cringle.

Source code used to run the tests are not found in this repository. Download
and place in the following locations, relative to the repository root:

- ext/zex/zexdoc.com

The original location of the exerciser seems to be here:

- http://mdfs.net/Software/Z80/Exerciser/

The sources found in the resource pack were downloaded from:

- https://github.com/anotherlin/z80emu/blob/master/testfiles

Helpful references:

- https://floooh.github.io/2016/07/12/z80-rust-ms1.html
- http://jeffavery.ca/computers/macintosh_z80exerciser.html

Run the functional test with:

```bash
go test -v -tags=long -timeout 60m
```

The following tests are currently failing:

- TestZexdoc/cpd1: `cpd<r>`
- TestZexdoc/rot8080: `<rlca,rrca,rla,rra>`

Run the benchmarks with:

```bash
go test -run=X -tags=long -bench=.
```