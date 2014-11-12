package machine

import (
	"encoding/binary"
	"github.com/MattParker89/seaquell/btree"
	"math"
	"strings"
)

type stack struct {
	pointer int
	values  []interface{}
}

func (s *stack) Push(value interface{}) {
	if s.pointer >= len(s.values) {
		s.values = append(s.values, value)
	} else {
		s.values[s.pointer] = value
	}
	s.pointer++
}

func (s *stack) Pop() interface{} {
	s.pointer--
	v := s.values[s.pointer]
	s.values[s.pointer] = nil
	return v
}

func toRecord(value interface{}) []byte {
	switch v := value.(type) {
	case int64:
		b := make([]byte, 8+8)
		binary.LittleEndian.PutUint64(b[:8], 6)
		binary.LittleEndian.PutUint64(b[8:], uint64(v))
		return b
	case float64:
		b := make([]byte, 8+8)
		bits := math.Float64bits(v)
		binary.LittleEndian.PutUint64(b[:8], 7)
		binary.LittleEndian.PutUint64(b[8:], bits)
		return b
	case string:
		b := make([]byte, len(v)+8)
		binary.LittleEndian.PutUint64(b[:8], uint64(len(v)*2+13))
		copy(b[8:], []byte(v))
		return b
	}
	return nil
}

//Returns the value that was extracted and the length of the entire value (code + body)
func fromRecord(b []byte) (interface{}, int) {
	code := int(binary.LittleEndian.Uint64(b[:8]))
	switch {
	case code == 6:
		return int64(binary.LittleEndian.Uint64(b[8 : 8+8])), 16
	case code == 7:
		return float64(binary.LittleEndian.Uint64(b[8 : 8+8])), 16
	case code >= 13 && code%2 != 0:
		return string(b[8 : 8+((code-13)/2)]), 8 + ((code - 13) / 2)
	}
	return nil, 0
}

func dataLength(b []byte) int {
	code := int(binary.LittleEndian.Uint64(b))
	var length int
	switch {
	case code == 6:
		length = 8
	case code == 7:
		length = 8
	case code >= 13 && code%2 != 0:
		length = ((code - 13) / 2)
	}
	return length
}

func loadRegisterInt(register *int64, value int64) {
	*register = value
}
func loadRegisterR4(register *interface{}, value interface{}) {
	*register = value
}
func loadRegisterR5(register *rune, value rune) {
	*register = value
}

func openRead(master *table, tableID int64) *table {
	t := open(master, tableID)
	t.tree.CursorFront()
	return t
}

func openWrite(master *table, tableID int64) *table {
	return open(master, tableID)
}

func open(master *table, tableID int64) *table {
	if tableID == 0 {
		return master
	}
	record := master.findRecordWithID(tableID)
	columns := parseSchema(record["sql"].(string))
	t := &table{
		key:     record["key"].(int64),
		page:    record["page"].(int64),
		columns: columns,
	}
	t.tree = btree.Fetch(int(t.page))
	return t
}

func newRowID(t *table) int64 {
	return int64(t.tree.LastKey() + 1)
}

/*
HELPER FUNCTIONS
*/
func (v *Machine) findTableByID(id int) *table {
	return &table{}
}

func parseSchema(sql string) []*column {
	var columns []*column
	beginning := strings.Index(sql, "(")
	end := strings.Index(sql, ")")
	sql = string(sql[beginning+1 : end])
	vars := strings.Split(sql, ", ")
	for i, vari := range vars {
		args := strings.Split(vari, " ")
		columns = append(columns, &column{
			index: i,
			name:  args[0],
		})
	}
	return columns
}
