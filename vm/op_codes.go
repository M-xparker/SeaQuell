package vm

type OpCode int

const (
	OP_LOAD_R1 OpCode = iota
	OP_LOAD_R2
	OP_LOAD_R3
	OP_LOAD_R4
	OP_LOAD_R5
	OP_PUSH_R1
	OP_PUSH_R2
	OP_PUSH_R3
	OP_PUSH_R4
	OP_PUSH_R5
	OP_STRING     //loads string onto stack
	OP_INTEGER    //loads integer onto stack
	OP_RESULT_ROW //invokes callback
	OP_COLUMN     //takes the column number at R3 and returns that data
	OP_JUMP       //Unconditional jump to address at R2
	OP_HALT       //stop execution
	OP_OPEN_READ  //Open table
	OP_OPEN_WRITE
	OP_WRITE_ROW
	OP_MAKE_RECORD
	OP_CLOSE
	OP_NEW_ROW_ID
	OP_CREATE_TABLE
	OP_COLUMN_NAME //names of the columns in the return data
	OP_REWIND
	OP_NEXT
)

type DataType int

const (
	TEXT DataType = iota
	INTEGER
)

func (d DataType) String() string {
	switch d {
	case TEXT:
		return "TEXT"
	case INTEGER:
		return "INTEGER"
	}
	return "NONE"
}

func (o OpCode) String() string {
	switch o {
	case OP_LOAD_R1:
		return "OP_LOAD_R1"
	case OP_LOAD_R2:
		return "OP_LOAD_R2"
	case OP_LOAD_R3:
		return "OP_LOAD_R3"
	case OP_LOAD_R4:
		return "OP_LOAD_R4"
	case OP_LOAD_R5:
		return "OP_LOAD_R5"
	case OP_PUSH_R1:
		return "OP_PUSH_R1"
	case OP_PUSH_R2:
		return "OP_PUSH_R2"
	case OP_PUSH_R3:
		return "OP_PUSH_R3"
	case OP_PUSH_R4:
		return "OP_PUSH_R4"
	case OP_PUSH_R5:
		return "OP_PUSH_R5"
	case OP_STRING:
		return "OP_STRING"
	case OP_INTEGER:
		return "OP_INTEGER"
	case OP_RESULT_ROW:
		return "OP_RESULT_ROW"
	case OP_COLUMN:
		return "OP_COLUMN"
	case OP_JUMP:
		return "OP_JUMP"
	case OP_HALT:
		return "OP_HALT"
	case OP_OPEN_READ:
		return "OP_OPEN_READ"
	case OP_OPEN_WRITE:
		return "OP_OPEN_WRITE"
	case OP_WRITE_ROW:
		return "OP_WRITE_ROW"
	case OP_MAKE_RECORD:
		return "OP_MAKE_RECORD"
	case OP_CLOSE:
		return "OP_CLOSE"
	case OP_NEW_ROW_ID:
		return "OP_NEW_ROW_ID"
	case OP_CREATE_TABLE:
		return "OP_CREATE_TABLE"
	case OP_COLUMN_NAME:
		return "OP_COLUMN_NAME"
	case OP_REWIND:
		return "OP_REWIND"
	case OP_NEXT:
		return "OP_NEXT"
	}
	return "NONE"
}
