package parse

import (
	"testing"
)

func Test_Compile_Select_Create(t *testing.T) {
	c := &codeGen{}
	r := c.Compile(&NodeSelectFrom{})
	if r == nil {
		t.Error("Expected a ROM; got nil")
	}
}

// func Test_Compile_Select_Instructions(t *testing.T) {
// 	c := &codeGen{}
// 	r := c.Compile(&NodeSelectFrom{})
// 	t.Error(r)
// }
