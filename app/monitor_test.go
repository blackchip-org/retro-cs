package app

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs"
)

type monitorFixture struct {
	out bytes.Buffer
	mon *Monitor
}

func newMonitorFixture() *monitorFixture {
	mach := mock.NewMach()
	f := &monitorFixture{mon: NewMonitor(mach)}
	f.mon.out.SetOutput(&f.out)
	return f
}

func testMonitorInput(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func testMonitorRun(mon *Monitor) {
	go mon.Run()
	mon.mach.Run()
}

func TestMonitor(t *testing.T) {
	for _, test := range monitorTests {
		t.Run(test.name, func(t *testing.T) {
			f := newMonitorFixture()
			f.mon.in = testMonitorInput(strings.Join(test.in, "\n"))
			testMonitorRun(f.mon)
			have := strings.TrimSpace(f.out.String())
			want := strings.TrimSpace(test.want)
			if have != want {
				t.Errorf("\n have: \n%v \n want: \n%v", have, want)
			}
		})
	}
}

var monitorTests = []struct {
	name string
	in   []string
	want string
}{
	{
		"dasm",
		[]string{
			"dasm lines 4",
			"poke 10 09",
			"poke 11 19 ab",
			"poke 13 29 cd ab",
			"poke 16 27 cd ab",
			"d 10",
			"q",
		},
		`
$0010:  09        i09
$0011:  19 ab     i19 $ab
$0013:  29 cd ab  i29 $abcd
$0016:  27 cd ab  i27 $abcd
`,
	}, {
		"dasm continue",
		[]string{
			"dasm lines 1",
			"poke 0 09",
			"poke 1 19 ab",
			"poke 3 29 cd ab",
			"poke 6 27 cd ab",
			"d",
			"d",
			"",
			"q",
		},
		`
$0000:  09        i09
$0001:  19 ab     i19 $ab
$0003:  29 cd ab  i29 $abcd
		`,
	}, {
		"dasm range",
		[]string{
			"dasm lines 4",
			"poke 10 09",
			"poke 11 19 ab",
			"poke 13 29 cd ab",
			"poke 16 27 cd ab",
			"d 11 14",
			"q",
		},
		`
$0011:  19 ab     i19 $ab
$0013:  29 cd ab  i29 $abcd
		`,
	}, {
		"memory",
		[]string{"mem lines 2", "m", "q"},
		`
$0000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
	`,
	}, {
		"memory page 1",
		[]string{"mem lines 2", "m $100", "q"},
		`
$0100  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0110  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
	`,
	}, {
		"memory next page",
		[]string{"mem lines 2", "m $100", "m", "q"},
		`
$0100  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0110  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0120  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0130  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory next page repeat",
		[]string{"mem lines 2", "m $100", "", "q"},
		`
$0100  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0110  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0120  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0130  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory range",
		[]string{"m $140 $15f", "q"},
		`
$0140  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0150  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory lines",
		[]string{"mem lines 3", "mem lines", "m", "q"},
		`
3
$0000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory encoding",
		[]string{
			"mem encoding",
			"poke 0 1 2 3 41 42 43",
			"m 00 0f",
			"mem encoding az26",
			"m 00 0f",
			"mem encoding ebcdic",
			"q",
		},
		`
* ascii
  az26
$0000  01 02 03 41 42 43 00 00  00 00 00 00 00 00 00 00  ...ABC..........
$0000  01 02 03 41 42 43 00 00  00 00 00 00 00 00 00 00  ABC.............
no such encoding: ebcdic
		`,
	}, {
		"registers and flags set",
		[]string{
			"r",
			"cpu reg pc 1234",
			"cpu reg a 56",
			"cpu reg c ff",
			"cpu flag q on",
			"cpu flag a on",
			"r",
			"q",
		},
		`
[pause]
pc:0000 a:00 b:00 q:false z:false
no such register: c
no such flag: a
[pause]
pc:1234 a:56 b:00 q:true z:false
		`,
	}, {
		"registers and flags list",
		[]string{"cpu reg", "cpu flag", "q"},
		`
a
b
pc
q
z
		`,
	}, {
		"registers and flags set",
		[]string{
			"cpu reg a 56",
			"cpu flag q on",
			"cpu reg a",
			"cpu reg c",
			"cpu flag q",
			"cpu flag a",
			"q"},
		`
$56 +86 %01010110
no such register: c
true
no such flag: a
		`,
	},
}

func TestDump(t *testing.T) {
	var dumpTests = []struct {
		name     string
		start    int
		data     func() []int
		showFrom int
		showTo   int
		want     string
	}{
		{
			"one line", 0x10,
			func() []int { return []int{} },
			0x10, 0x20, "" +
				"$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................",
		}, {
			"two lines", 0x10,
			func() []int { return []int{} },
			0x10, 0x30, "" +
				"$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................\n" +
				"$0020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................",
		}, {
			"jagged top", 0x10,
			func() []int { return []int{} },
			0x14, 0x30, "" +
				"$0010              00 00 00 00  00 00 00 00 00 00 00 00      ............\n" +
				"$0020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................",
		}, {
			"jagged bottom", 0x10,
			func() []int { return []int{} },
			0x10, 0x2b, "" +
				"$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................\n" +
				"$0020  00 00 00 00 00 00 00 00  00 00 00 00              ............",
		},
		{
			"single value", 0x10,
			func() []int { return []int{0, 0x41} },
			0x11, 0x11, "" +
				"$0010     41                                              A",
		},
		{
			"$40-$5f", 0x10,
			func() []int {
				data := make([]int, 0)
				for i := 0x40; i < 0x60; i++ {
					data = append(data, i)
				}
				return data
			},
			0x10, 0x30, "" +
				"$0010  40 41 42 43 44 45 46 47  48 49 4a 4b 4c 4d 4e 4f  @ABCDEFGHIJKLMNO\n" +
				"$0020  50 51 52 53 54 55 56 57  58 59 5a 5b 5c 5d 5e 5f  PQRSTUVWXYZ[\\]^_",
		},
	}

	mock.ResetMemory()
	m := mock.TestMemory
	for _, test := range dumpTests {
		t.Run(test.name, func(t *testing.T) {
			for i, value := range test.data() {
				m.Write(test.start+i, uint8(value))
			}
			have := dump(m, test.showFrom, test.showTo, rcs.AsciiDecoder)
			have = strings.TrimSpace(have)
			if have != test.want {
				t.Errorf("\n have: \n%v \n want: \n%v \n", have, test.want)
			}
		})
	}
}
