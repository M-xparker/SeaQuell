package storage

import (
	"encoding/binary"
	"os"
)

type storage struct {
	file          *os.File
	firstFreePage uint64
}

type Storer interface {
	Close()
	WritePage(p *page)
	Get(offset uint64, length int) []byte
	GetFreePage() uint64
	writeHeader()
}

//UGLY!!!!!!!
var store Storer

func init() {
	//bad hack in order to get the btree tests to pass
	store = &storage{
		firstFreePage: db_header_length + 1,
	}
}

func Create(name string) (*storage, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	s := &storage{
		file:          f,
		firstFreePage: db_header_length + 1 + page_length,
	}

	h := writeHeader()
	s.file.WriteAt(h[:], 0)
	// p := createTableInteriorPage()
	// s.addPage(p)
	store = s
	return s, nil
}

func Open(name string) (*storage, error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	s := &storage{
		file:          f,
		firstFreePage: db_header_length + 1 + page_length,
	}
	s.parseHeader()

	store = s
	return s, nil
}

func (s *storage) GetFreePage() uint64 {
	defer func() {
		s.firstFreePage += page_length

	}()
	return (s.firstFreePage - 1 - db_header_length) / page_length
}

func GetPageNumber(number int) *page {
	return &page{
		offset: uint64(number*page_length + db_header_length + 1),
		header: NewPageHeader(),
	}
}

func (s *storage) writeHeader() {
	var h dbHeader
	copy(h[header_string_offset:header_string_offset+header_string_size], header_string)
	binary.LittleEndian.PutUint16(h[page_size_offset:page_size_offset+page_size_length], uint16(page_length))
	binary.LittleEndian.PutUint32(h[free_page_offset:free_page_offset+free_page_size], uint32(s.firstFreePage))
	s.file.WriteAt(h[:], 0)
}

func (s *storage) parseHeader() {
	h := s.Get(0, db_header_length)
	s.firstFreePage = uint64(binary.LittleEndian.Uint32(h[free_page_offset : free_page_offset+free_page_size]))
}

func (s *storage) Close() {
	defer s.file.Close()
}
func (s *storage) WritePage(p *page) {
	// fmt.Println("WRITE PAGE", p.offset)
	// return
	s.file.WriteAt(p.buffer[:], int64(p.offset))
	// fi, _ := s.file.Stat()
	// fmt.Println("File size:", fi.Size())
}

func (s *storage) Get(offset uint64, length int) []byte {
	b := make([]byte, length)
	s.file.ReadAt(b, int64(offset))
	// fmt.Println("FETCH FROM DISK")
	// fi, _ := s.file.Stat()
	// fmt.Println("File size:", fi.Size())
	return b
}
