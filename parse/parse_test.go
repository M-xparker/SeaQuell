package parse

import (
	"testing"
)

func Test_Parse_SelectFrom(t *testing.T) {
	tree := New("SELECT age FROM person;")
	tree.Parse()
	node, ok := tree.Root.(*NodeSelectFrom)
	if !ok {
		t.Error("wrong node type returned")
	}
	if len(node.fields) != 1 {
		t.Error("fields weren't populated")
	}
	if node.table != "person" {
		t.Error("table name wasn't populated ", node.table)
	}
}

func Test_Parse_CreateTable(t *testing.T) {
	tree := New("CREATE TABLE test(one int, two text);")
	tree.Parse()
	node, ok := tree.Root.(*NodeCreateTable)
	if !ok {
		t.Error("wrong node type returned")
	}
	if node.table != "test" {
		t.Error("table name wasn't populated ", node.table)
	}
	if len(node.columns) != 2 {
		t.Error("wrong number of columns", len(node.columns), node.columns[2])
	}
}
