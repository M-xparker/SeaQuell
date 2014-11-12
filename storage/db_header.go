package storage

import (
	"encoding/binary"
)

const (
	db_header_length     = 100
	page_length          = 32768
	header_string        = "SeaQuell Test\000\000\000"
	header_string_offset = 0
	header_string_size   = 16
	page_size_offset     = 16
	page_size_length     = 2
	free_page_offset     = 36
	free_page_size       = 4
)

type dbHeader [db_header_length]byte

func writeHeader() dbHeader {
	var h dbHeader
	copy(h[header_string_offset:header_string_offset+header_string_size], header_string)
	binary.LittleEndian.PutUint16(h[page_size_offset:page_size_offset+page_size_length], uint16(page_length))

	//kick off the first free page after the header
	binary.LittleEndian.PutUint32(h[free_page_offset:free_page_offset+free_page_size], uint32(db_header_length))
	return h
}
