package vm

type VM interface {
	Exec(*ROM) *ResultRow
}

type resultCode int

const (
	RESULT_OK resultCode = iota
	RESULT_ERROR
)

type ResultRow struct {
	Columns    []string
	Data       []interface{}
	ResultCode resultCode
}
