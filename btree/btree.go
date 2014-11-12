package btree

import (
	"github.com/MattParker89/seaquell/storage"
)

/*
A key's right child is inclusive of the key
The left child is < the key
*/

const (
	NODE_ORDER = 4
)

type cursor struct {
	node  *leafNode
	index int
}

type BTree struct {
	root   noder
	cursor cursor
}

func New() *BTree {
	return &BTree{
		root: &leafNode{
			page: storage.NewPage(),
		},
	}
}

func Fetch(number int) *BTree {
	rootPage := storage.GetPageNumber(number)
	switch rootPage.Type() {
	case storage.LEAF_NODE:
		l := &leafNode{
			page: rootPage,
		}
		keys, vals := l.page.FetchLeaf()
		l.keys = keys
		l.values = vals
		return &BTree{
			root: l,
		}
	case storage.INTERIOR_NODE:
		i := &interiorNode{
			page: rootPage,
		}
		keys, _ := i.page.FetchInterior()
		i.keys = keys
		return &BTree{
			root: i,
		}
	}
	return nil
}
func (t *BTree) Insert(key int, value []byte) {
	newNode := t.root.insert(uint64(key), value)
	if newNode != nil {
		t.root = newNode
		t.root.write()
	}
}

func (t *BTree) Delete(key int) {
	t.root.delete(uint64(key))
}

func (t *BTree) Get(key int) []byte {
	return t.root.get(uint64(key))
}

func (t *BTree) Print() {
	t.root._print()
}

func (b *BTree) CursorData() []byte {
	if b.cursor.node == nil {
		return nil
	}
	data := b.cursor.node.getValueAtIndex(b.cursor.index)
	return data
}

func (b *BTree) CursorFront() {
	b.cursor.node = b.root.getLeft()
}

func (b *BTree) CursorAvailable() bool {
	return b.cursor.node != nil
}

func (b *BTree) CursorNext() {
	b.cursor.index++
	if b.cursor.index >= len(b.cursor.node.values) {
		b.cursor.node = b.cursor.node.right
		b.cursor.index = 0
	}

}

func (b *BTree) CursorKey() uint64 {
	if b.cursor.node == nil {
		return 0
	}
	return b.cursor.node.getKeyAtIndex(b.cursor.index)
}

func (b *BTree) LastKey() int {
	return int(b.root.getRight().page.NumberOfKeys())
}

type noder interface {
	insert(key uint64, value []byte) *interiorNode
	delete(uint64)
	get(uint64) []byte
	Page() storage.Pager
	Keys() []uint64 //only capitalized because nodes have a field keys
	write()
	getLeft() *leafNode  //get left most child
	getRight() *leafNode //get right most child
	_print()             //debugging only
}
