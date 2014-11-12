package parse

import (
	"github.com/MattParker89/seaquell/vm"
)

func Generate(sql string) *vm.ROM {
	tree := New(sql)
	tree.Parse()
	c := &codeGen{}
	return c.Compile(tree.Root)
}

type codeGen struct {
}

func (c *codeGen) Compile(s statementer) *vm.ROM {
	var rom *vm.ROM
	switch s := s.(type) {
	case *NodeSelectFrom:
		rom = selectFrom(s)
	case *NodeCreateTable:
		rom = createTable(s)
	case *NodeInsertInto:
		rom = insertInto(s)
	}
	var f vm.Frame
	f = vm.NewFrame()
	f.Op = vm.OP_CLOSE
	rom.Add(f)

	f = vm.NewFrame()
	f.Op = vm.OP_HALT
	rom.Add(f)
	return rom
}

func createTable(n *NodeCreateTable) *vm.ROM {
	rom := vm.NewROM()
	var f vm.Frame
	f = vm.NewFrame()

	f.Op = vm.OP_OPEN_WRITE
	f.Value = "master"
	rom.Add(f)

	f.Op = vm.OP_NEW_ROW_ID
	rom.Add(f)

	f.Op = vm.OP_PUSH_R3
	rom.Add(f)

	f.Op = vm.OP_STRING
	f.Value = n.table
	rom.Add(f)

	f.Op = vm.OP_PUSH_R4
	rom.Add(f)

	f.Op = vm.OP_CREATE_TABLE
	rom.Add(f)

	f.Op = vm.OP_PUSH_R3
	rom.Add(f)

	f.Op = vm.OP_STRING
	sql := "CREATE TABLE " + n.table + "("
	for i, c := range n.columns {
		if i > 0 {
			sql += ", "
		}
		sql += c.name + " " + c.typ
	}
	sql += ");"
	f.Value = sql
	rom.Add(f)

	f.Op = vm.OP_PUSH_R4
	rom.Add(f)

	f.Op = vm.OP_MAKE_RECORD
	f.Value = int64(3)
	rom.Add(f)

	f.Op = vm.OP_WRITE_ROW
	rom.Add(f)
	return rom
}

func selectFrom(n *NodeSelectFrom) *vm.ROM {
	rom := vm.NewROM()
	var f vm.Frame

	f = vm.NewFrame()
	f.Op = vm.OP_OPEN_READ
	f.Value = n.table
	rom.Add(f)

	f = vm.NewFrame()
	f.Op = vm.OP_REWIND
	rom.Add(f)

	for _, field := range n.fields {
		f = vm.NewFrame()
		f.Op = vm.OP_COLUMN
		f.Value = field.name
		rom.Add(f)
	}

	f = vm.NewFrame()
	f.Op = vm.OP_RESULT_ROW
	f.Value = len(n.fields)
	rom.Add(f)

	f = vm.NewFrame()
	f.Op = vm.OP_NEXT
	rom.Add(f)

	return rom
}

func insertInto(n *NodeInsertInto) *vm.ROM {
	rom := vm.NewROM()
	var f vm.Frame
	f = vm.NewFrame()

	f.Op = vm.OP_OPEN_WRITE
	f.Value = n.table
	rom.Add(f)

	f.Op = vm.OP_NEW_ROW_ID
	rom.Add(f)

	f.Op = vm.OP_PUSH_R3
	rom.Add(f)

	for _, value := range n.values {
		switch value := value.(type) {
		case int64:
			f.Op = vm.OP_INTEGER
			f.Value = value
			rom.Add(f)

			f.Op = vm.OP_PUSH_R3
			rom.Add(f)
		case string:
			f.Op = vm.OP_STRING
			f.Value = value
			rom.Add(f)

			f.Op = vm.OP_PUSH_R4
			rom.Add(f)
		}
	}

	f.Op = vm.OP_MAKE_RECORD
	f.Value = int64(len(n.values))
	rom.Add(f)

	f.Op = vm.OP_WRITE_ROW
	rom.Add(f)
	return rom
}
