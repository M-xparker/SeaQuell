package parse

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// unexported keeps Node implementations local to the package.
// All implementations embed Pos, so this takes care of it.
func (Pos) unexported() {
}

type FieldNode struct {
	name string
}

type statementer interface {
}

type NodeSelectFrom struct {
	fields    []*FieldNode
	table     string
	modifiers []statementer
	hasStar   bool
}

type Column struct {
	name string
	typ  string
}

type NodeCreateTable struct {
	columns []*Column
	table   string
}

type NodeInsertInto struct {
	values []interface{}
	table  string
}
