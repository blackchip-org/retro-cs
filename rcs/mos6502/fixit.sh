#!/bin/bash

r() {
    sed -e "$1" -i instructions_test.go
}

r s/StoreN/WriteN/g
r s/Store16/WriteLE/g
r s/Store/Write/g
r s/Load16/ReadLE/g
r s/Load/Read/g
r s/flagZ/FlagZ/g
r s/flagB/FlagB/g
r s/flagN/FlagN/g
r s/flagC/FlagC/g
r 's/| flag5//g'
r 's/c.SR()/c.SR/g'
r 's/testRunCPU.c./testRunCPU\(t, c\)/g'
