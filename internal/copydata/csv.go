package copydata

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"time"
)

// csvRowReader implements RowReader for CSV files.
type csvRowReader struct {
	r      *csv.Reader
	fields []string
	types  []string
}

func NewCSVRowReader(file io.Reader) (RowReader, error) {
	if file == nil {
		return nil, errors.New("file reader is nil")
	}
	r := csv.NewReader(file)

	// Read header (fields)
	fields, err := r.Read()
	if err != nil {
		return nil, err
	}

	return &csvRowReader{
		r:      r,
		fields: fields,
	}, nil
}

func (c *csvRowReader) ReadRow() (Row, error) {
	rec, err := c.r.Read()
	if err != nil {
		return nil, err
	}
	row := make(Row, len(rec))
	for i, v := range rec {
		row[i] = v
	}
	return row, nil
}

func (c *csvRowReader) Fields() []string { return c.fields }
func (c *csvRowReader) Types() []string  { return c.types }

// csvRowWriter implements RowWriter for CSV files.
type csvRowWriter struct {
	w           *csv.Writer
	wroteHeader bool
}

func NewCSVRowWriter(file io.Writer) (RowWriter, error) {
	if file == nil {
		return nil, errors.New("file writer is nil")
	}
	w := csv.NewWriter(file)

	return &csvRowWriter{
		w: w,
	}, nil
}

func (c *csvRowWriter) WriteFields(fields []string, types []string) error {
	if c.wroteHeader {
		return nil
	}
	if err := c.w.Write(fields); err != nil {
		return err
	}
	c.wroteHeader = true
	return nil
}

func (c *csvRowWriter) WriteRow(row Row) (rowsWritten int, err error) {
	rec := make([]string, len(row))
	for i, v := range row {
		if v == nil {
			rec[i] = ""
		} else {
			rec[i] = toString(v)
		}
	}
	if err = c.w.Write(rec); err != nil {
		return
	}
	rowsWritten = 1
	return
}

func (c *csvRowWriter) Flush() (rowsWritten int, err error) {
	c.w.Flush()
	return
}

// toString converts any value to its string representation.
// If the value is a string, it returns it as-is.
// If the value is a []byte, it converts the bytes to a string.
// For simples types like int, float, bool, it uses fmt.Sprint to convert to string.
// For all other types, it marshals the value to JSON and returns the resulting JSON string.
func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool,
		complex64, complex128:
		return fmt.Sprint(t)
	case time.Time:
		return t.Format(time.RFC3339)
	case time.Duration:
		return t.String()
	case time.Month:
		return t.String()
	case time.Weekday:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
