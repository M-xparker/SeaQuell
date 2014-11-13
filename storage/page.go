package storage

import (
	"encoding/binary"
)

/*
Basics of a page:
- Each page has a header
- The header will tell us where the cell pointer array (CPA) is
- CPA tells us where each cell is

Leaf:
- The first 2 bytes tell you the key
- The remaining bytes have the following format:
	- 2 bytes = length of the following value
	- The value
Interior:
- 8 bytes pointer
- 2 bytes key
*/

type Pager interface {
	WriteLeaf(keys []uint64, values [][]byte, rightPtr Pager)
	WriteInterior(keys []uint64, children []Pager)
	FetchInterior() ([]uint64, []Pager)
	FetchLeaf() ([]uint64, [][]byte, Pager)
	Free()
	Offset() uint64
	Type() NodeType
	NumberOfKeys() uint16
}

type NodeType uint8

func (n NodeType) Byte() byte {
	return byte(n)
}

const (
	LEAF_NODE NodeType = iota
	INTERIOR_NODE
)

const (
	key_length = 8
)

type page struct {
	buffer       [page_length]byte
	offset       uint64
	header       *pageHeader
	cellPointers []byte
}

func NewPage() *page {
	return &page{
		header: NewPageHeader(),
		offset: store.GetFreePage()*page_length + db_header_length + 1,
	}
}

func (p *page) WriteLeaf(keys []uint64, values [][]byte, rightPtr Pager) {
	p.header.nodeType = LEAF_NODE
	p.buffer = [page_length]byte{}

	//start writing from the back of the page and move inwards
	p.header.cellContentArea = uint16(page_length)

	//make sure we're starting with a blank page
	cellPointer := p.header.cellPointerArray
	p.header.numberOfCells = 0

	//we loop through the keys because with a leaf node there is a 1 to 1 match on keys to values
	for i, k := range keys {
		//write the payload
		val := values[i]
		copy(p.buffer[p.header.cellContentArea-uint16(len(val)):p.header.cellContentArea], val)
		p.header.cellContentArea -= uint16(len(val))

		//write the length of the payload so we know where the cell ends
		valueLength := uint16(len(val))
		binary.LittleEndian.PutUint16(p.buffer[p.header.cellContentArea-2:p.header.cellContentArea], valueLength)
		p.header.cellContentArea -= 2

		//write the key
		binary.LittleEndian.PutUint64(p.buffer[p.header.cellContentArea-8:p.header.cellContentArea], k)
		p.header.cellContentArea -= 8

		//add a pointer to the cell pointer array
		binary.LittleEndian.PutUint16(p.buffer[cellPointer:cellPointer+2], uint16(p.header.cellContentArea))
		cellPointer += 2
		p.header.numberOfCells += 1
	}
	if rightPtr != nil {
		p.header.rightMostPointer = rightPtr.Offset()
	}
	p.writeHeader()
	store.writeHeader()
	store.WritePage(p)
}

func (p *page) WriteInterior(keys []uint64, children []Pager) {
	p.header.nodeType = INTERIOR_NODE

	//clearing the buffer because I had a weird bug
	//due to it being used multiple times when a node gets promoted
	p.buffer = [page_length]byte{}

	//start writing from the back of the page and move inwards
	p.header.cellContentArea = uint16(page_length)

	//make sure we're starting with a blank page
	cellPointer := p.header.cellPointerArray
	p.header.numberOfCells = 0

	//we loop over the children because the format for an interior node is
	//child -> key -> child
	//so there are more children than keys
	for i, c := range children {
		//write the pointer to the child's page
		binary.LittleEndian.PutUint64(p.buffer[p.header.cellContentArea-8:p.header.cellContentArea], c.Offset())
		p.header.cellContentArea -= 8

		//there are more children than keys
		//so we need to make sure we don't go over
		if i < len(keys) {
			//write the key
			binary.LittleEndian.PutUint64(p.buffer[p.header.cellContentArea-8:p.header.cellContentArea], keys[i])
			p.header.cellContentArea -= 8

			//totally arbitrary. We can choose to keep track
			//of the keys or the pointers
			p.header.numberOfCells += 1
		}

		//add the pointer to the cell pointer array
		binary.LittleEndian.PutUint16(p.buffer[cellPointer:cellPointer+2], uint16(p.header.cellContentArea))

		cellPointer += 2
	}
	p.writeHeader()
	store.writeHeader()
	store.WritePage(p)
}

func (p *page) FetchInterior() ([]uint64, []Pager) {
	if !p.isFetched() {
		p.fetch()
	}
	p.parseHeader()

	keys := []uint64{}
	pages := []Pager{}

	for i := 0; i < int(p.header.numberOfCells); i++ {
		pointerBytes := p.cellPointers[i*2 : (i*2)+2]
		cellOffset := binary.LittleEndian.Uint16(pointerBytes)

		keyBytes := p.buffer[cellOffset : cellOffset+8]
		key := binary.LittleEndian.Uint64(keyBytes)

		childPointerBytes := p.buffer[cellOffset+8 : cellOffset+8+8]
		childPointer := binary.LittleEndian.Uint64(childPointerBytes)
		childPage := &page{
			offset: childPointer,
			header: NewPageHeader(),
		}

		keys = append(keys, uint64(key))
		pages = append(pages, childPage)

		if i+1 == int(p.header.numberOfCells) {
			childPointerBytes = p.buffer[cellOffset-8 : cellOffset]
			childPointer = binary.LittleEndian.Uint64(childPointerBytes)
			childPage = &page{
				offset: childPointer,
				header: NewPageHeader(),
			}

			pages = append(pages, childPage)

		}
	}

	return keys, pages

}
func (p *page) NumberOfKeys() uint16 {
	if !p.isFetched() {
		p.fetch()
		p.parseHeader()
	}
	return p.header.numberOfCells
}
func (p *page) fetch() {
	copy(p.buffer[:], store.Get(p.offset, page_length))
}
func (p *page) FetchLeaf() (keys []uint64, values [][]byte, rightPtr Pager) {
	if !p.isFetched() {
		p.fetch()
	}
	p.parseHeader()

	for i := 0; i < int(p.header.numberOfCells); i++ {
		//go through the cell pointer array and find the offset
		pointerBytes := p.cellPointers[i*2 : (i*2)+2]
		cellOffset := binary.LittleEndian.Uint16(pointerBytes)

		//we have the offset. Let's first read the key
		keyBytes := p.buffer[cellOffset : cellOffset+8]
		key := binary.LittleEndian.Uint64(keyBytes)

		//Now let's read the length of the body
		cellLengthBytes := p.buffer[cellOffset+8 : cellOffset+8+2]
		cellLength := binary.LittleEndian.Uint16(cellLengthBytes)

		//Get the payload
		val := p.buffer[cellOffset+8+2 : cellOffset+8+2+cellLength]

		/*
			We need to make a copy of the byte array. If not then the values in the btree get tied to the page's buffer because their pointers are the same. If you clear the buffer, you clear the btree node's values.
		*/
		v := make([]byte, len(val))
		copy(v, val)

		keys = append(keys, uint64(key))
		values = append(values, v)
	}

	var rightPage Pager = nil
	if p.header.rightMostPointer != 0 {
		rightPage = &page{
			offset: p.header.rightMostPointer,
			header: NewPageHeader(),
		}
	}

	return keys, values, rightPage
}

func (p *page) parseHeader() {
	cpaOffsetBytes := p.buffer[btreePageHeaderConfig[cell_pointer_array].offset : btreePageHeaderConfig[cell_pointer_array].offset+btreePageHeaderConfig[cell_pointer_array].size]
	p.header.cellPointerArray = binary.LittleEndian.Uint16(cpaOffsetBytes)
	if p.header.cellPointerArray == 0 {
		p.header = NewPageHeader()
		return
	}

	numberCellsBytes := p.buffer[btreePageHeaderConfig[number_of_cells].offset : btreePageHeaderConfig[number_of_cells].offset+btreePageHeaderConfig[number_of_cells].size]
	p.header.numberOfCells = binary.LittleEndian.Uint16(numberCellsBytes)

	rightPtrBytes := p.buffer[btreePageHeaderConfig[right_most_pointer].offset : btreePageHeaderConfig[right_most_pointer].offset+btreePageHeaderConfig[right_most_pointer].size]
	p.header.rightMostPointer = binary.LittleEndian.Uint64(rightPtrBytes)
	p.cellPointers = p.buffer[p.header.cellPointerArray : p.header.cellPointerArray+p.header.numberOfCells*2]

	nodeTypeBytes := p.buffer[btreePageHeaderConfig[node_type].offset]
	p.header.nodeType = NodeType(nodeTypeBytes)
}

func (p *page) writeHeader() {
	binary.LittleEndian.PutUint16(p.buffer[btreePageHeaderConfig[cell_pointer_array].offset:btreePageHeaderConfig[cell_pointer_array].offset+btreePageHeaderConfig[cell_pointer_array].size], p.header.cellPointerArray)
	binary.LittleEndian.PutUint16(p.buffer[btreePageHeaderConfig[number_of_cells].offset:btreePageHeaderConfig[number_of_cells].offset+btreePageHeaderConfig[number_of_cells].size], p.header.numberOfCells)
	binary.LittleEndian.PutUint64(p.buffer[btreePageHeaderConfig[right_most_pointer].offset:btreePageHeaderConfig[right_most_pointer].offset+btreePageHeaderConfig[right_most_pointer].size], p.header.rightMostPointer)
	kb := p.header.nodeType.Byte()
	p.buffer[btreePageHeaderConfig[node_type].offset] = kb
}

func (p *page) Free() {

}

func (p *page) Type() NodeType {
	if !p.isFetched() {
		p.fetch()
		p.parseHeader()
	}
	return p.header.nodeType
}

func (p *page) isFetched() bool {
	return p.header.numberOfCells > 0
}

func (p *page) Offset() uint64 {
	return p.offset
}

func (p *page) Buffer() [page_length]byte {
	return p.buffer
}

func createTableInteriorPage() *page {
	p := &page{}
	writeIntToHeader(btreePageHeaderConfig[node_type].offset, btreePageHeaderConfig[node_type].size, int(tableInterior), p.buffer)
	return p
}

func writeIntToHeader(offset, size, value int, header [page_length]byte) {
	switch size {
	case 1:
		header[offset+1] = uint8(value)
	case 2:
		binary.LittleEndian.PutUint16(header[offset:offset+size], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(header[offset:offset+size], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(header[offset:offset+size], uint64(value))
	}
}
