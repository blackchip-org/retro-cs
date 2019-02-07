package main

//go:generate go run .
//go:generate go fmt ../../../rcs/cbm/petscii_table.go

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/blackchip-org/retro-cs/config"
)

var (
	root      = filepath.Join("..", "..", "..")
	targetDir = filepath.Join(root, "rcs", "cbm")
	sourceDir = filepath.Join(config.ResourceDir(), "ext", "petscii")
)

func main() {
	log.SetFlags(0)
	infile := filepath.Join(sourceDir, "table.txt")
	in, err := os.Open(infile)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	var table [2][0x100]rune

	s := bufio.NewScanner(in)
	s.Split(bufio.ScanLines)
	for i := 0; i < 6; i++ {
		s.Scan()
	}
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, "CHARACTER") {
			break
		}
		fields := strings.Split(line, "|")
		cp := strings.TrimSpace(fields[1])
		charCode, err := strconv.ParseUint(cp[1:], 16, 8)
		if err != nil {
			log.Fatalln(err)
		}
		ucp := strings.TrimSpace(fields[2])
		ucps := regexp.MustCompile("\\s+").Split(ucp, -1)
		if len(ucps) == 1 {
			ucps = append(ucps, ucps[0])
		}

		for tableN, uni := range ucps {
			if uni == "-" {
				table[tableN][charCode] = 0xfffd
			} else {
				value, err := strconv.ParseUint(uni[2:], 16, 16)
				if err != nil {
					log.Fatalln(err)
				}
				table[tableN][int(charCode)] = rune(value)
			}
		}
	}
	if s.Err() != nil {
		log.Fatalln(s.Err())
	}
	var out bytes.Buffer
	out.WriteString("package cbm\n")
	out.WriteString("var petsciiUnshifted = map[uint8]rune {\n")
	for i, val := range table[0] {
		ch := string(val)
		if val == '\'' {
			ch = "\\" + ch
		}
		if printable(i) {
			out.WriteString(fmt.Sprintf("0x%02x: '%s',\n", i, ch))
		}
	}
	out.WriteString("}\n")
	out.WriteString("var petsciiShifted = map[uint8]rune {\n")
	for i, val := range table[1] {
		ch := string(val)
		if val == '\'' {
			ch = "\\" + ch
		}
		if printable(i) {
			out.WriteString(fmt.Sprintf("0x%02x: '%s',\n", i, ch))
		}
	}
	out.WriteString("}\n")
	outfile := filepath.Join(targetDir, "petscii_table.go")
	err = ioutil.WriteFile(outfile, out.Bytes(), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func printable(i int) bool {
	if i >= 0x20 && i <= 0x7f {
		return true
	}
	if i >= 0xa0 && i <= 0xbf {
		return true
	}
	return false
}
