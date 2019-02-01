package app

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
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

func TestMemoryFirstLine(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(f.out.String(), "\n")
	have := lines[0]
	want := "$0000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", lines[0], want)
	}
}

func TestMemoryLastLine(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := lines[len(lines)-1]
	want := "$00f0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestMemoryPage(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m 0100 \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := lines[len(lines)-1]
	want := "$01f0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestMemoryNextPage(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m 0100 \n m \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := lines[len(lines)-1]
	want := "$02f0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestMemoryNextPageRepeat(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m 0100 \n m \n \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := lines[len(lines)-1]
	want := "$03f0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestMemoryRange(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("m 0100 018f \n q")
	testMonitorRun(f.mon)
	lines := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := lines[len(lines)-1]
	want := "$0180  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestRegisters(t *testing.T) {
	f := newMonitorFixture()
	f.mon.in = testMonitorInput("r \n q")
	testMonitorRun(f.mon)
	// := strings.Split(strings.TrimSpace(f.out.String()), "\n")
	have := f.out.String()
	want := "[pause]\ncpu status registers\n"
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v \n", have, want)
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
			have := dump(m, test.showFrom, test.showTo, AsciiDecoder)
			have = strings.TrimSpace(have)
			if have != test.want {
				t.Errorf("\n have: \n%v \n want: \n%v \n", have, test.want)
			}
		})
	}
}
