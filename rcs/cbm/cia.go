package cbm

type CIA struct {
	DataA *uint8
	DataB *uint8
}

func NewCIA() *CIA {
	return &CIA{}
}
