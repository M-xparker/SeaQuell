package storage

import (
	"bytes"
	"testing"
)

type MockStorer struct {
}

func (m *MockStorer) Close() {

}
func (m *MockStorer) WritePage(p *page) {

}
func (m *MockStorer) Get(offset uint64, length int) []byte {
	return []byte{}
}
func (m *MockStorer) GetFreePage() uint64 {
	return 0
}

func Test_Leaf(t *testing.T) {
	store = &MockStorer{}
	p := &page{header: NewPageHeader()}

	keys := []uint64{0, 1, 2, 3}
	values := [][]byte{[]byte{0}, []byte{1}, []byte{2}, []byte{3}}
	p.WriteLeaf(keys, values, 0)

	//The key thing here is that we keep the same buffer.
	newPage := &page{
		header: NewPageHeader(),
		buffer: p.buffer,
	}
	fetchedKeys, fetchedValues := newPage.FetchLeaf()
	for i, k := range keys {
		if k != fetchedKeys[i] {
			t.Errorf("Keys: Expected %d; got %d", k, fetchedKeys[i])
		}
		if !bytes.Equal(values[i], fetchedValues[i]) {
			t.Errorf("Values: Expected %v; got %v", values[i], fetchedValues[i])
		}
	}
}

func Test_Interior(t *testing.T) {
	store = &MockStorer{}
	mainPage := &page{
		header: NewPageHeader(),
	}

	keys := []uint64{1}
	children := []Pager{&page{offset: 10}, &page{offset: 20}}
	mainPage.WriteInterior(keys, children)

	//The key thing here is that we keep the same buffer.
	newPage := &page{
		header: NewPageHeader(),
		buffer: mainPage.buffer,
	}
	fetchedKeys, fetchedChildren := newPage.FetchInterior()
	for i, k := range keys {
		if k != fetchedKeys[i] {
			t.Errorf("Keys: Expected %d; got %d", k, fetchedKeys[i])
		}
	}
	for i, c := range children {
		if c.Offset() != fetchedChildren[i].Offset() {
			t.Errorf("Children: Expected %v; got %v", c.Offset(), fetchedChildren[i].Offset())
		}
	}
}
