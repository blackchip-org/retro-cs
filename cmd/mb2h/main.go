package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func mb2h(in string) (string, error) {
	in = strings.Replace(in, "-", "0", -1)
	strlo := strings.Replace(in, "x", "0", -1)
	strhi := strings.Replace(in, "x", "1", -1)

	lo, err := strconv.ParseUint(strlo, 2, 16)
	if err != nil {
		return "", err
	}
	hi, err := strconv.ParseUint(strhi, 2, 16)
	if err != nil {
		return "", err
	}
	if lo == hi {
		return fmt.Sprintf("%04x", lo), nil
	}
	return fmt.Sprintf("%04x - %04x", lo, hi), nil
}

func main() {
	flag.Parse()
	val, err := mb2h(flag.Arg(0))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(val)
}
