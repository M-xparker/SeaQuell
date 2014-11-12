package btree

import (
	"bytes"
	"github.com/MattParker89/seaquell/storage"
	"testing"
)

type MockPager struct {
}

func (m *MockPager) WriteLeaf(keys []uint64, values [][]byte, rightPtr uint64) {

}
func (m *MockPager) WriteInterior(keys []uint64, children []storage.Pager) {

}
func (m *MockPager) Create() storage.Pager {
	return &MockPager{}
}
func (m *MockPager) FetchInterior() ([]uint64, []storage.Pager) {
	return []uint64{}, []storage.Pager{}
}
func (m *MockPager) FetchLeaf() ([]uint64, [][]byte) {
	return []uint64{}, [][]byte{}
}
func (m *MockPager) Free() {

}
func (m *MockPager) Offset() uint64 {
	return 0
}
func (m *MockPager) Type() storage.NodeType {
	return storage.LEAF_NODE
}

func Test_Leaf_findIndexOfKey(t *testing.T) {
	l := &leafNode{
		keys: []uint64{0, 1, 2, 3, 4},
	}
	i := l.findIndexOfKey(3)
	if i != 3 {
		t.Error("wrong index returned", i)
	}
}

func Test_Leaf_get_no_fetch(t *testing.T) {
	selectedVal := []byte{2}
	l := &leafNode{
		keys:   []uint64{0, 1, 2, 3, 4},
		values: [][]byte{[]byte{0}, []byte{1}, selectedVal, []byte{3}},
	}
	val := l.get(2)
	if !bytes.Equal(selectedVal, val) {
		t.Errorf("Expected %v; got %v", selectedVal, val)
	}

}

func Test_split(t *testing.T) {
	keys := []uint64{0, 1, 2, 3}
	vals := [][]byte{[]byte{0}, []byte{1}, []byte{2}, []byte{3}}

	l := &leafNode{
		keys:   keys,
		values: vals,
		page:   &MockPager{},
	}
	parent := l.split()
	if len(parent.children) != 2 {
		t.Error("not split in two")
	}
	for i := 0; i < len(keys)/2; i++ {
		if parent.children[0].Keys()[i] != keys[i] {
			t.Error("split incorrectly")
		}
	}
	for i := 0; i < len(keys)/2; i++ {
		if parent.children[1].Keys()[i] != keys[i+(len(keys)/2)] {
			t.Error("split incorrectly")
		}
	}

	if parent.keys[0] != parent.children[1].Keys()[0] {
		t.Error("split on wrong key")
	}

}
