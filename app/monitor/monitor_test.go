package monitor

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs"
)

type monitorFixture struct {
	cpu *mock.CPU
	out bytes.Buffer
	mon *Monitor
}

func newMonitorFixture() *monitorFixture {
	mach := mock.NewMach()
	mach.Init()
	cpu := mach.CPU["cpu"].(*mock.CPU)
	mon, err := New(mach)
	if err != nil {
		panic(err)
	}
	f := &monitorFixture{
		mon: mon,
		cpu: cpu,
	}
	f.mon.out = log.New(&f.out, "", 0)
	return f
}

func TestMonitor(t *testing.T) {
	for _, test := range monitorTests {
		t.Run(test.name, func(t *testing.T) {
			f := newMonitorFixture()
			go f.mon.mach.Run()
			defer func() {
				f.mon.mach.Command(rcs.MachQuit)
			}()
			f.mon.Eval(strings.Join(test.in, "\n"))
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
		"conversions",
		[]string{"42", "$2a", "%101010"},
		`
+ 42
42 $2a %101010
+ $2a
42 $2a %101010
+ %101010
42 $2a %101010
		`,
	}, {
		"break",
		[]string{
			"bps $3456",
			"bps $2345",
			"bps $1234",
			"bp",
			"bpc $2345",
			"bp",
			"bpn",
			"bp",
			"bps $123456",
		},
		`
+ bps $3456
+ bps $2345
+ bps $1234
+ bp
$1234
$2345
$3456
+ bpc $2345
+ bp
$1234
$3456
+ bpn
+ bp
+ bps $123456
invalid address: $123456
		`,
	}, {
		"dasm",
		[]string{
			"config lines-disassembly 4",
			"poke $10 $09",
			"poke $11 $19 $ab",
			"poke $13 $29 $cd $ab",
			"poke $16 $27 $cd $ab",
			"d $10",
		},
		`
+ config lines-disassembly 4
+ poke $10 $09
+ poke $11 $19 $ab
+ poke $13 $29 $cd $ab
+ poke $16 $27 $cd $ab
+ d $10
$0010:  09        i09
$0011:  19 ab     i19 $ab
$0013:  29 cd ab  i29 $abcd
$0016:  27 cd ab  i27 $abcd
`,
	}, {
		"dasm continue",
		[]string{
			"config lines-disassembly 1",
			"poke $0 $09",
			"poke $1 $19 $ab",
			"poke $3 $29 $cd $ab",
			"poke $6 $27 $cd $ab",
			"d",
			"d",
		},
		`
+ config lines-disassembly 1
+ poke $0 $09
+ poke $1 $19 $ab
+ poke $3 $29 $cd $ab
+ poke $6 $27 $cd $ab
+ d
$0000:  09        i09
+ d
$0001:  19 ab     i19 $ab
		`,
	}, {
		"dasm range",
		[]string{
			"config lines-disassembly 4",
			"poke $10 $09",
			"poke $11 $19 $ab",
			"poke $13 $29 $cd $ab",
			"poke $16 $27 $cd $ab",
			"d $11 $14",
		},
		`
+ config lines-disassembly 4
+ poke $10 $09
+ poke $11 $19 $ab
+ poke $13 $29 $cd $ab
+ poke $16 $27 $cd $ab
+ d $11 $14
$0011:  19 ab     i19 $ab
$0013:  29 cd ab  i29 $abcd
		`,
	}, {
		"go",
		[]string{"bps $10", "g", "sleep 100"},
		`
+ bps $10
+ g
+ sleep 100

[break]
pc:0010 a:00 b:00 q:false z:false
		`,
	}, {
		"memory",
		[]string{"config lines-memory 2", "m"},
		`
+ config lines-memory 2
+ m
$0000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
	`,
	}, {
		"memory page 1",
		[]string{"config lines-memory 2", "m $100"},
		`
+ config lines-memory 2
+ m $100
$0100  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0110  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
	`,
	}, {
		"memory next page",
		[]string{"config lines-memory 2", "m $100", "m"},
		`
+ config lines-memory 2
+ m $100
$0100  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0110  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
+ m
$0120  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0130  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory range",
		[]string{"m $140 $15f"},
		`
+ m $140 $15f
$0140  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0150  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory lines",
		[]string{"config lines-memory 3", "config lines-memory", "m"},
		`
+ config lines-memory 3
+ config lines-memory
3 $3 %11
+ m
$0000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
$0020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
`,
	}, {
		"memory encoding",
		[]string{
			"encoding",
			"poke $0 1 2 3 $41 $42 $43",
			"m 00 $0f",
			"encoding az26",
			"m 00 $0f",
			"encoding ebcdic",
		},
		`
+ encoding
ascii
+ poke $0 1 2 3 $41 $42 $43
+ m 00 $0f
$0000  01 02 03 41 42 43 00 00  00 00 00 00 00 00 00 00  ...ABC..........
+ encoding az26
+ m 00 $0f
$0000  01 02 03 41 42 43 00 00  00 00 00 00 00 00 00 00  ABC.............
+ encoding ebcdic
invalid value: ebcdic
		`,
	}, {
		"memory fill",
		[]string{
			"config lines-memory 1",
			"mem fill 0004 $000b $ff",
			"m 0",
			"mem fill $000b 0004 $aa",
			"m 0",
		},
		`
+ config lines-memory 1
+ mem fill 0004 $000b $ff
+ m 0
$0000  00 00 00 00 ff ff ff ff  ff ff ff ff 00 00 00 00  ................
+ mem fill $000b 0004 $aa
+ m 0
$0000  00 00 00 00 ff ff ff ff  ff ff ff ff 00 00 00 00  ................
		`,
	}, {
		"next",
		[]string{"n"},
		`
+ n
$0000:  00        i00
		`,
	}, {
		"peek",
		[]string{"poke $1234 $ab", "peek $1234"},
		`
+ poke $1234 $ab
+ peek $1234
171 $ab %10101011
		`,
	}, {
		"step",
		[]string{"s", "i", "s", "s", "i"},
		`
+ s
$0001:  00        i00
+ i
[pause]
pc:0001 a:00 b:00 q:false z:false
+ s
$0002:  00        i00
+ s
$0003:  00        i00
+ i
[pause]
pc:0003 a:00 b:00 q:false z:false
		`,
	}, {
		"trace",
		[]string{
			"poke 0 $a $b $c",
			"trace",
			"bps 1",
			"bps 2",
			"go",
			"sleep 100",
			"trace",
			"go",
		},
		`
+ poke 0 $a $b $c
+ trace
+ bps 1
+ bps 2
+ go
+ sleep 100
$0000:  0a        i0a

[break]
pc:0001 a:00 b:00 q:false z:false
+ trace
+ go
		`,
	}, {
		"watch",
		[]string{
			"ws $10 rw",
			"w",
			"poke $10 $22",
			"peek $10",
			"wc $10",
			"ws $10 r",
			"w",
			"poke $10 $22",
			"peek $10",
			"wc $10",
			"ws $10 w",
			"w",
			"poke $10 $22",
			"peek $10",
			"wn",
			"w",
		},
		`
+ ws $10 rw
+ w
$0010 rw
+ poke $10 $22
write($0010) => $22
+ peek $10
$22 <= read($0010)
34 $22 %100010
+ wc $10
+ ws $10 r
+ w
$0010 r
+ poke $10 $22
+ peek $10
$22 <= read($0010)
34 $22 %100010
+ wc $10
+ ws $10 w
+ w
$0010 w
+ poke $10 $22
write($0010) => $22
+ peek $10
34 $22 %100010
+ wn
+ w
		`,
	},
}

func TestBreakpointOffset(t *testing.T) {
	f := newMonitorFixture()
	f.cpu.OffsetPC = 1
	cmds := `
bps $10
g
sleep 100
	`
	go f.mon.mach.Run()
	defer func() {
		f.mon.mach.Command(rcs.MachQuit)
	}()
	f.mon.Eval(cmds)
	have := strings.TrimSpace(f.out.String())
	want := strings.TrimSpace(`
+ bps $10
+ g
+ sleep 100

[break]
pc:000f a:00 b:00 q:false z:false
`)
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v", have, want)
	}
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
			have := dump(m, test.showFrom, test.showTo, rcs.ASCIIDecoder, "")
			have = strings.TrimSpace(have)
			if have != test.want {
				t.Errorf("\n have: \n%v \n want: \n%v \n", have, test.want)
			}
		})
	}
}

func TestRepeatWriter(t *testing.T) {
	tests := []struct {
		in  []string
		out []string
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		{
			[]string{"a", "b", "b", "c"},
			[]string{"a", "b", "... repeats 1 time", "c"},
		},
		{
			[]string{"a", "b", "b", "b", "b", "b", "c"},
			[]string{"a", "b", "... repeats 4 times", "c"},
		},
		{
			[]string{"a", "b", "b", "b", "b", "b", "c", "c", "d"},
			[]string{"a", "b", "... repeats 4 times", "c", "... repeats 1 time", "d"},
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		rw := newRepeatWriter(&buf)
		rw.ansi = false
		str := strings.Join(test.in, "\n")
		rw.Write([]byte(str))
		rw.Close()
		have := buf.String()
		want := strings.Join(test.out, "\n") + "\n"
		if have != want {
			t.Errorf("\n have: \n[%v] \n want: \n[%v]", have, want)
		}
	}
}
