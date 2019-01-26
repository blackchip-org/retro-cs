package rcs

type Status int

const (
	Halt Status = iota
	Run
	Break
)

func (s Status) String() string {
	switch s {
	case Halt:
		return "halt"
	case Run:
		return "run"
	case Break:
		return "break"
	}
	return "???"
}

type Mach struct {
	Mem  []*Memory
	CPU  []CPU
	Proc []Proc
}
