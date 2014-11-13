package storage

type btreePageHeaderSection int

const (
	node_type btreePageHeaderSection = iota
	first_freeblock
	number_of_cells
	cell_content_area
	cell_pointer_array
	number_of_fragmented_bytes
	right_most_pointer
)

type btreeType int

const (
	tableLeaf     btreeType = 5
	tableInterior btreeType = 13
)

const (
	page_header_length = 18
)

type headerData struct {
	offset int
	size   int
}

var btreePageHeaderConfig = map[btreePageHeaderSection]*headerData{
	node_type:                  &headerData{0, 1},
	first_freeblock:            &headerData{1, 2},
	number_of_cells:            &headerData{3, 2},
	cell_content_area:          &headerData{5, 2},
	cell_pointer_array:         &headerData{7, 2},
	number_of_fragmented_bytes: &headerData{9, 1},
	right_most_pointer:         &headerData{10, 8},
}

type pageHeader struct {
	length                      int
	cellPointerArray            uint16
	cellContentArea             uint16
	nodeType                    NodeType
	firstFreeBlock              uint16
	numberOfCells               uint16
	numberOfFragmentedFreeBytes uint16
	overflowPage                uint32
	rightMostPointer            uint64
}

func NewPageHeader() *pageHeader {
	return &pageHeader{
		length:           page_header_length,
		cellPointerArray: page_header_length + 1,
	}
}
