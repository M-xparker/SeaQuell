package machine

import (
	"github.com/MattParker89/seaquell/btree"
	"github.com/MattParker89/seaquell/storage"
	"github.com/MattParker89/seaquell/vm"
)

type Machine struct {
	R1     int64       //arbitrary integer. Usually cursor number
	R2     int64       //arbitrary integer. Usually a jump destination
	R3     int64       //arbitrary integer
	R4     interface{} //can be anything
	R5     rune        //used for flags
	stack  *stack
	rom    *vm.ROM
	master *table
	store  storage.Storer
}

func New(filename string) *Machine {
	m := &Machine{
		stack: new(stack),
	}
	m.open(filename)
	return m
}

func (m *Machine) Close() {
	m.store.Close()
}

func (v *Machine) open(filename string) {
	var err error
	v.store, err = storage.Open(filename)
	if err != nil {
		v.store, _ = storage.Create(filename)
	}
	tree := btree.Fetch(0)
	v.master = &table{
		key:     0,
		name:    "master",
		page:    0,
		columns: parseSchema("CREATE TABLE master(name text, page int, sql text);"),
		tree:    tree,
	}
}

func (v *Machine) Exec(rom *vm.ROM) []vm.ResultRow {
	// var err error
	v.rom, _ = v.Preprocess(rom)
	result := make(chan vm.ResultRow)
	go v.run(result)
	return gatherResults(result)
}

func gatherResults(result chan vm.ResultRow) (results []vm.ResultRow) {
	for r := range result {
		results = append(results, r)
	}
	return results
}

func (v *Machine) run(result chan vm.ResultRow) {
	var currentTable *table
	rowResult := vm.ResultRow{}
Loop:
	for i := 0; i < len(v.rom.Frames); i++ {
		frame := v.rom.Frames[i]
		switch frame.Op {
		case vm.OP_LOAD_R1, vm.OP_LOAD_R2, vm.OP_LOAD_R3:
			var register *int64
			switch frame.Op {
			case vm.OP_LOAD_R1:
				register = &v.R1
			case vm.OP_LOAD_R2:
				register = &v.R2
			case vm.OP_LOAD_R3:
				register = &v.R3
			}
			loadRegisterInt(register, frame.Value.(int64))
		case vm.OP_LOAD_R4:
			loadRegisterR4(&v.R4, frame.Value)
		case vm.OP_LOAD_R5:
			loadRegisterR5(&v.R5, frame.Value.(rune))
		case vm.OP_PUSH_R3:
			v.stack.Push(v.R3)
		case vm.OP_PUSH_R4:
			v.stack.Push(v.R4)
		case vm.OP_OPEN_READ:
			currentTable = openRead(v.master, frame.Value.(int64))
		case vm.OP_OPEN_WRITE:
			currentTable = openWrite(v.master, frame.Value.(int64))
		case vm.OP_NEW_ROW_ID:
			v.R3 = newRowID(currentTable)
		case vm.OP_STRING:
			v.R4 = frame.Value
		case vm.OP_INTEGER:
			v.R3 = frame.Value.(int64)
		case vm.OP_MAKE_RECORD:
			b := []byte{}
			for i := 0; i < int(frame.Value.(int64)); i++ {
				val := v.stack.Pop()
				r := toRecord(val)
				b = append(r, b...)
			}
			v.stack.Push(b)
		case vm.OP_WRITE_ROW:
			value := v.stack.Pop()
			key := v.stack.Pop()
			currentTable.tree.Insert(int(key.(int64)), value.([]byte))
		case vm.OP_CLOSE:
			currentTable = nil
		case vm.OP_HALT:
			close(result)
			break Loop
		case vm.OP_COLUMN:
			b := currentTable.tree.CursorData()
			column := frame.Value.(int)
			var columnCount int
			for i := 0; i < len(b); {
				if columnCount == column {
					value, _ := fromRecord(b[i:])
					rowResult.Data = append(rowResult.Data, value)
					break
				}
				i += 8 + dataLength(b[i:i+8])
				columnCount++
			}
			rowResult.Columns = append(rowResult.Columns, currentTable.columns[column].name)
		case vm.OP_RESULT_ROW:
			result <- rowResult
		case vm.OP_NEXT:
			rowResult = vm.ResultRow{}
			currentTable.tree.CursorNext()
			if currentTable.tree.CursorAvailable() {
				i = int(frame.Value.(int64)) - 1
			}
		case vm.OP_CREATE_TABLE:
			pageNumber := int(v.store.GetFreePage())
			page := storage.GetPageNumber(pageNumber)
			page.WriteLeaf([]uint64{}, [][]byte{}, nil)
			v.R3 = int64(pageNumber)
		}
	}
}

func (m *Machine) findTableByName(name string) *table {
	if name == "master" {
		return m.master
	}
	records := m.master.findRecordsByFieldValue("name", name)
	if len(records) == 0 {
		return nil
	}
	tableData := records[0]
	return &table{
		key:     tableData["key"].(int64),
		name:    tableData["name"].(string),
		columns: parseSchema(tableData["sql"].(string)),
		page:    tableData["page"].(int64),
	}
}
