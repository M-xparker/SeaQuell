package vm

type ROM struct {
	Frames []Frame
}

type Frame struct {
	Op    OpCode
	Value interface{}
	Type  DataType
}

func NewFrame() Frame {
	return Frame{}
}

func NewROM() *ROM {
	return &ROM{}
}

func (r *ROM) Add(f Frame) {
	r.Frames = append(r.Frames, f)
}
