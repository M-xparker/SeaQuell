package machine

import (
	"github.com/MattParker89/seaquell/btree"
)

type table struct {
	key     int64
	name    string
	page    int64
	columns []*column
	tree    *btree.BTree
}

func (t *table) getColumnByName(name string) *column {
	for _, c := range t.columns {
		if c.name == name {
			return c
		}
	}
	return nil
}

func (t *table) findRecordWithID(id int64) map[string]interface{} {
	b := t.tree.Get(int(id))
	m := t.recordFromBytes(b)
	m["key"] = int64(t.tree.CursorKey())
	return m

}

func (t *table) findRecordsByFieldValue(field string, value interface{}) []map[string]interface{} {
	records := []map[string]interface{}{}
	for t.tree.CursorFront(); t.tree.CursorAvailable(); t.tree.CursorNext() {
		data := t.tree.CursorData()
		record := t.recordFromBytes(data)
		record["key"] = int64(t.tree.CursorKey())
		if record[field] == value {
			records = append(records, record)
		}
	}
	return records
}

func (t *table) recordFromBytes(b []byte) map[string]interface{} {
	var columnCount int
	record := map[string]interface{}{}
	for i := 0; i < len(b); {
		value, nextLocation := fromRecord(b[i:])
		record[t.columns[columnCount].name] = value
		columnCount++
		i += nextLocation
	}
	return record
}

type column struct {
	index int
	name  string
	typ   string
}
