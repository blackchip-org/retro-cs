package monitor

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

func newCompleter(m *Monitor) *readline.PrefixCompleter {
	cmds := []readline.PrefixCompleterInterface{
		readline.PcItem("b",
			readline.PcItem("clear"),
			readline.PcItem("clear-all"),
			readline.PcItem("list"),
			readline.PcItem("set"),
		),
		readline.PcItem("break",
			readline.PcItem("clear"),
			readline.PcItem("clear-all"),
			readline.PcItem("list"),
			readline.PcItem("set"),
		),
		readline.PcItem("cpu",
			readline.PcItem("flag",
				readline.PcItemDynamic(acFlags(m)),
			),
			readline.PcItem("reg",
				readline.PcItemDynamic(acRegisters(m)),
			),
			readline.PcItem("select"),
		),
		readline.PcItem("d"),
		readline.PcItem("dasm",
			readline.PcItem("lines"),
			readline.PcItem("list"),
		),
		readline.PcItem("export"),
		readline.PcItem("g"),
		readline.PcItem("go"),
		readline.PcItem("import"),
		readline.PcItem("m"),
		readline.PcItem("mem",
			readline.PcItem("dump"),
			readline.PcItem("encoding",
				readline.PcItemDynamic(acEncodings(m)),
			),
			readline.PcItem("fill"),
			readline.PcItem("lines"),
		),
		readline.PcItem("n"),
		readline.PcItem("next"),
		readline.PcItem("p"),
		readline.PcItem("pause"),
		readline.PcItem("poke"),
		readline.PcItem("peek"),
		readline.PcItem("q"),
		readline.PcItem("quit"),
		readline.PcItem("r"),
		readline.PcItem("s"),
		readline.PcItem("step"),
		readline.PcItem("t"),
		readline.PcItem("trace"),
		readline.PcItem("w",
			readline.PcItem("clear"),
			readline.PcItem("clear-all"),
			readline.PcItem("list"),
			readline.PcItem("set"),
		),
		readline.PcItem("watch",
			readline.PcItem("clear"),
			readline.PcItem("clear-all"),
			readline.PcItem("list"),
			readline.PcItem("set"),
		),
	}
	switch config.System {
	case "c64":
		cmds = append(cmds, []readline.PrefixCompleterInterface{
			readline.PcItem("load-prg",
				readline.PcItemDynamic(acDataFiles(m, ".prg")),
			),
		}...)
	}
	return readline.NewPrefixCompleter(cmds...)
}

func acDataFiles(m *Monitor, suffix string) func(string) []string {
	return func(line string) []string {
		results := make([]string, 0, 0)
		arg := ""
		i := strings.LastIndex(line, " ")
		if i >= 0 {
			arg = line[i:]
		}
		var dir, prefix string
		if filepath.IsAbs(arg) {
			dir = filepath.Dir(arg)
			prefix = filepath.Base(arg)
		} else {
			dir = config.DataDir
			prefix = arg
		}
		prefix = strings.TrimSpace(prefix)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			m.out.Println(err)
			return []string{}
		}
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
				results = append(results, name)
			}
		}
		return results
	}
}

func acRegisters(m *Monitor) func(string) []string {
	return func(line string) []string {
		cpu, ok := m.sc.cpu.(rcs.CPUEditor)
		if !ok {
			return []string{}
		}
		names := make([]string, 0)
		for k := range cpu.Registers() {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}

func acFlags(m *Monitor) func(string) []string {
	return func(line string) []string {
		cpu, ok := m.sc.cpu.(rcs.CPUEditor)
		if !ok {
			return []string{}
		}
		names := make([]string, 0)
		for k := range cpu.Flags() {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}

func acEncodings(m *Monitor) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		for k := range m.mach.CharDecoders {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}
