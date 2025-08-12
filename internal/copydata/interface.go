package copydata

type Row []any

type RowReader interface {
	ReadRow() (Row, error)
	Fields() []string
	Types() []string
	Close() error
}

type RowWriter interface {
	WriteFields(fields []string, types []string) error
	WriteRow(row Row) error
	Close() error
}
