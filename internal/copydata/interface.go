package copydata

// Row represents a single row of data, with each column as an any type.
type Row []any

type RowReader interface {
	ReadRow() (Row, error)
	Fields() []string
	Types() []string
}

type RowWriter interface {
	WriteFields(fields []string, types []string) error
	WriteRow(row Row) (rowsWritten int, err error)
	Flush() (rowsWritten int, err error)
}
