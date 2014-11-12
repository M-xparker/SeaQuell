package btree

import (
	"fmt"
	"github.com/MattParker89/seaquell/storage"
)

type interiorNode struct {
	keys     []uint64
	children []noder
	page     storage.Pager
}

func (i *interiorNode) insert(key uint64, value []byte) *interiorNode {
	keyLength := len(i.keys)
	var index int = keyLength
	for x := 0; x < keyLength; x++ {
		if key <= i.keys[x] {
			index = x
		}
	}
	i.Fetch(key)
	child := i.children[index]
	newNode := child.insert(key, value)
	if newNode != nil {
		newNode.page.Free()

		keys := newNode.Keys()
		k := keys[0]

		oldKeys := i.keys
		i.keys = make([]uint64, len(oldKeys)+1)
		copy(i.keys[:index], oldKeys[:index])
		i.keys[index] = k
		copy(i.keys[index+1:], oldKeys[index:])

		oldChildren := i.children
		oldChildren = append(oldChildren[:index], oldChildren[index+1:]...)
		i.children = make([]noder, len(oldChildren)+len(newNode.children))
		copy(i.children[:index], oldChildren[:index])
		copy(i.children[index+len(newNode.children):], oldChildren[index:])
		copy(i.children[index:], newNode.children)

		newNode = nil
	}
	if len(i.keys) >= NODE_ORDER {
		return i.split()
	}
	i.write()
	return nil
}

func (i *interiorNode) split() *interiorNode {
	n := int(len(i.keys) / 2)
	parentKey := i.keys[n]
	newKeys := i.keys[n+1:]
	newChildren := i.children[n+1:]
	i.keys = i.keys[:n]
	i.children = i.children[:n+1]

	right := &interiorNode{
		keys:     newKeys,
		children: newChildren,
		page:     storage.NewPage(),
	}

	left := &interiorNode{
		keys:     i.keys,
		children: i.children,
		page:     storage.NewPage(),
	}
	i.keys = []uint64{parentKey}
	i.children = []noder{left, right}

	right.write()
	left.write()
	return i
}

func (i *interiorNode) delete(key uint64) {
	index := i.findIndexOfKey(key)
	child := i.children[index]
	child.delete(key)
	if len(child.Keys()) < (NODE_ORDER/2)-1 {
		updatedKeys := append(i.keys[:index], i.keys[index+1:]...)
		i.keys = updatedKeys
		updatedChildren := append(i.children[:index], i.children[index+1:]...)
		i.children = updatedChildren
	}
}

func (i *interiorNode) get(key uint64) []byte {
	if !i.isFetched() {
		i.Fetch(key)
	}
	index := i.findIndexOfKey(key)
	child := i.children[index]
	if child == nil {
		//this is ugly.
		//Right now the first fetch gets the keys
		//and which ever pointer we need.
		//We avoid fetching all pointers because we would need
		//to know what type they are and that requires reading from disk
		//To circumvent we read one pointer at a time except we allocate an array of children big enough to hold all of the children. This means the fetch condition is never triggered.
		i.Fetch(key)
		child = i.children[index]
	}
	return child.get(key)
}

func (i *interiorNode) getLeft() *leafNode {
	i.fetchIndex(0)
	return i.children[0].getLeft()
}
func (i *interiorNode) getRight() *leafNode {
	i.keys, _ = i.page.FetchInterior()
	i.fetchIndex(len(i.keys))
	return i.children[len(i.children)-1].getLeft()
}

func (i *interiorNode) write() {
	var childPages []storage.Pager
	var pages []storage.Pager
	for j, c := range i.children {
		var p storage.Pager
		if c == nil {
			if childPages == nil {
				_, childPages = i.page.FetchInterior()
			}
			p = childPages[j]
		} else {
			p = c.Page()
		}
		pages = append(pages, p)
	}
	i.page.WriteInterior(i.keys, pages)
}
func (i *interiorNode) Page() storage.Pager {
	return i.page
}

func (i *interiorNode) findIndexOfKey(key uint64) int {
	var index int = len(i.keys)
	for x, k := range i.keys {
		if key < (k) {
			index = x
			break
		}
	}
	return index
}

func (i *interiorNode) Keys() []uint64 {
	return i.keys
}

func (i *interiorNode) _print() {
	fmt.Println("Interior Node: ", i.keys)
	for _, c := range i.children {
		if c == nil {
			fmt.Println(nil)
			continue
		}
		c._print()
	}

}

func (i *interiorNode) fetch(indexFunc func() int) {
	keys, childPages := i.page.FetchInterior()
	i.keys = keys
	if len(i.children) < len(childPages) {
		//if we don't have enough children
		//we at least want enough spots to hold them
		i.children = make([]noder, len(childPages))
	}
	index := indexFunc()
	if i.children[index] != nil {
		return
	}
	childPage := childPages[index]

	var child noder
	switch childPage.Type() {
	case storage.LEAF_NODE:
		child = &leafNode{
			page: childPage,
		}
	case storage.INTERIOR_NODE:
		child = &interiorNode{
			page: childPage,
		}
	}
	i.children[index] = child
}

func (i *interiorNode) Fetch(key uint64) {
	i.fetch(func() int { return i.findIndexOfKey(key) })
}

func (i *interiorNode) fetchIndex(index int) {
	i.fetch(func() int { return index })
}

func (i *interiorNode) isFetched() bool {
	return len(i.children) == len(i.keys)+1
}
