package copydata

import (
	"encoding/csv"
	"errors"
	"io"
)

// csvRowReader implements RowReader for CSV files.
type csvRowReader struct {
	r      *csv.Reader
	fields []string
	types  []string
	closer io.Closer // optional, for closing file if needed
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

	var closer io.Closer
	if c, ok := file.(io.Closer); ok {
		closer = c
	}

	return &csvRowReader{
		r:      r,
		fields: fields,
		closer: closer,
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
func (c *csvRowReader) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}

// csvRowWriter implements RowWriter for CSV files.
type csvRowWriter struct {
	w           *csv.Writer
	closer      io.Closer // optional, for closing file if needed
	wroteHeader bool
}

func NewCSVRowWriter(file io.Writer) (RowWriter, error) {
	if file == nil {
		return nil, errors.New("file writer is nil")
	}
	w := csv.NewWriter(file)

	var closer io.Closer
	if c, ok := file.(io.Closer); ok {
		closer = c
	}

	return &csvRowWriter{
		w:      w,
		closer: closer,
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

func (c *csvRowWriter) WriteRow(row Row) error {
	rec := make([]string, len(row))
	for i, v := range row {
		if v == nil {
			rec[i] = ""
		} else {
			rec[i] = toString(v)
		}
	}
	return c.w.Write(rec)
}

func (c *csvRowWriter) Close() error {
	c.w.Flush()
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
