package machine

import (
	"fmt"
	"github.com/MattParker89/seaquell/vm"
)

func (v *Machine) Preprocess(rom *vm.ROM) (*vm.ROM, error) {
	var rewindLocations []int64
	var nextLocations []int64

	var currentTable *table

	for x, frame := range rom.Frames {
		i := int64(x)
		switch frame.Op {
		case vm.OP_OPEN_READ, vm.OP_OPEN_WRITE:
			switch val := frame.Value.(type) {
			case string:
				currentTable = v.findTableByName(val)
				frame.Value = currentTable.key
				rom.Frames[i] = frame
			}
		case vm.OP_REWIND:
			rewindLocations = append(rewindLocations, i)
		case vm.OP_COLUMN:
			switch val := frame.Value.(type) {
			case string:
				column := currentTable.getColumnByName(val)
				frame.Value = column.index
				rom.Frames[i] = frame
			}
		case vm.OP_NEXT:
			nextLocations = append(nextLocations, i)
		}
	}
	if len(rewindLocations) < len(nextLocations) {
		return nil, fmt.Errorf("Infinite Loop")
	}
	if len(rewindLocations) > len(nextLocations) {
		return nil, fmt.Errorf("Unnecessary loop")
	}
	for i := range rewindLocations {
		rewind := rom.Frames[rewindLocations[i]]
		rewind.Value = nextLocations[i] + 1
		rom.Frames[rewindLocations[i]] = rewind

		next := rom.Frames[nextLocations[i]]
		next.Value = rewindLocations[i] + 1
		rom.Frames[nextLocations[i]] = next
	}
	return rom, nil
}
