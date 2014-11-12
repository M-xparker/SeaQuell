package btree

import (
	"fmt"
	"github.com/MattParker89/seaquell/storage"
)

type leafNode struct {
	keys      []uint64
	values    [][]byte
	right     *leafNode
	page      storage.Pager
	isFetched bool
}

func (l *leafNode) insert(key uint64, value []byte) *interiorNode {
	l.Fetch(key)
	if len(l.keys) == 0 {
		l.keys = append(l.keys, key)
		l.values = append(l.values, value)
		l.write()
		return nil
	}
	var index int = len(l.keys)
	for x, k := range l.keys {
		if key <= (k) {
			index = x
			break
		}
	}

	oldKeys := l.keys
	l.keys = make([]uint64, index)
	copy(l.keys[:index], oldKeys[:index])
	l.keys = append(l.keys, key)
	l.keys = append(l.keys, oldKeys[index:]...)

	oldValues := l.values
	l.values = make([][]byte, index)
	copy(l.values[:index], oldValues[:index])
	l.values = append(l.values, value)
	l.values = append(l.values, oldValues[index:]...)

	keyLength := len(l.keys)

	//if the key length >= NODE_ORDER it means we've passed our limit on values
	//if m = len(values) then NODE_ORDER/2 <= m <= NODE_ORDER
	if keyLength >= NODE_ORDER {
		fmt.Println("split leaf")
		return l.split()
	}
	l.write()
	return nil

}

func (l *leafNode) write() {
	var rightPtr uint64
	if l.right != nil {
		rightPtr = l.right.page.Offset()
	}
	l.isFetched = true
	l.page.WriteLeaf(l.keys, l.values, rightPtr)
}

func (l *leafNode) Keys() []uint64 {
	return l.keys
}

func (l *leafNode) split() *interiorNode {
	keyLength := len(l.keys)
	i := int(keyLength / 2)
	newKeys := l.keys[i:]
	newValues := l.values[i:]
	l.keys = l.keys[:i]
	l.values = l.values[:i]

	newLeaf := &leafNode{
		keys:   newKeys,
		values: newValues,
		right:  l.right,
		page:   storage.NewPage(),
	}
	l.right = newLeaf

	parent := &interiorNode{
		keys:     []uint64{newLeaf.keys[0]},
		children: []noder{l, newLeaf},
		page:     l.page, //the promoted node keeps the current page since it might be root
	}
	l.page = storage.NewPage()
	l.write()
	newLeaf.write()

	return parent
}

func (l *leafNode) _print() {
	fmt.Print("Leaf Keys: ", l.keys, " Leaf Values: ", l.values)
	fmt.Println(" ")
}

func (l *leafNode) Page() storage.Pager {
	return l.page
}

func (l *leafNode) Fetch(key uint64) {
	if l.isFetched {
		return
	}
	keys, vals := l.page.FetchLeaf()
	if len(keys) == 0 {
		return
	}
	l.keys = keys
	l.values = vals
	l.isFetched = true
}

func (l *leafNode) delete(key uint64) {
	index := l.findIndexOfKey(key)
	if len(l.keys) == 1 {
		l.keys = []uint64{}
		l.values = [][]byte{}
		return
	}
	updatedKeys := append(l.keys[:index], l.keys[index+1:]...)
	l.keys = updatedKeys
	updatedValues := append(l.values[:index], l.values[index+1:]...)
	l.values = updatedValues
}

func (l *leafNode) get(key uint64) []byte {
	l.Fetch(key)

	index := l.findIndexOfKey(key)
	return l.values[index]
}

func (l *leafNode) getLeft() *leafNode {
	return l
}
func (l *leafNode) getRight() *leafNode {
	return l
}

func (l *leafNode) getValueAtIndex(index int) []byte {
	l.Fetch(0)
	return l.values[index]
}

func (l *leafNode) getKeyAtIndex(index int) uint64 {
	l.Fetch(0)
	return l.keys[index]
}

func (l *leafNode) findIndexOfKey(key uint64) int {
	var index int = len(l.keys) - 1
	for x, k := range l.keys {
		if key == k {
			index = x
			break
		}
	}
	return index
}
