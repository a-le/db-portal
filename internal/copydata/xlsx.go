package copydata

import (
	"errors"
	"io"
	"os"

	"github.com/tealeg/xlsx"
)

// xlsxRowReader implements RowReader for XLSX files.
type xlsxRowReader struct {
	file   *xlsx.File
	sheet  *xlsx.Sheet
	rowIdx int
	fields []string
	types  []string
	closer io.Closer // optional, for closing file if needed
}

// Helper: writes io.Reader to temp file, returns *os.File and size
func readerToTempFile(r io.Reader) (*os.File, int64, error) {
	tmpfile, err := os.CreateTemp("", "user-upload-*.xlsx")
	if err != nil {
		return nil, 0, err
	}
	written, err := io.Copy(tmpfile, r)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, 0, err
	}
	_, err = tmpfile.Seek(0, io.SeekStart)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, 0, err
	}
	return tmpfile, written, nil
}

// NewXLSXRowReader accepts io.Reader, writes to a temp file, and opens as io.ReaderAt.
func NewXLSXRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	tmpfile, size, err := readerToTempFile(r)
	if err != nil {
		return nil, err
	}

	file, err := xlsx.OpenReaderAt(tmpfile, size)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, err
	}
	if len(file.Sheets) == 0 {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, errors.New("xlsx: no sheets found")
	}
	sheet := file.Sheets[0]
	if len(sheet.Rows) == 0 {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, errors.New("xlsx: no rows found")
	}
	headerRow := sheet.Rows[0]
	fields := make([]string, len(headerRow.Cells))
	for i, cell := range headerRow.Cells {
		fields[i] = cell.String()
	}
	xr := &xlsxRowReader{
		file:   file,
		sheet:  sheet,
		rowIdx: 1,
		fields: fields,
		closer: &tempFileCloser{File: tmpfile},
	}
	return xr, nil
}

// tempFileCloser closes and removes the temp file.
type tempFileCloser struct {
	*os.File
}

func (t *tempFileCloser) Close() error {
	name := t.Name()
	if err := t.File.Close(); err != nil {
		return err
	}
	if err := os.Remove(name); err != nil {
		return err
	}
	return nil
}

func (x *xlsxRowReader) ReadRow() (Row, error) {
	if x.rowIdx >= len(x.sheet.Rows) {
		return nil, io.EOF
	}
	rowCells := x.sheet.Rows[x.rowIdx].Cells
	row := make(Row, len(x.fields))
	for i := range x.fields {
		if i < len(rowCells) {
			row[i] = rowCells[i].Value
		} else {
			row[i] = nil
		}
	}
	x.rowIdx++
	return row, nil
}

func (x *xlsxRowReader) Fields() []string { return x.fields }
func (x *xlsxRowReader) Types() []string  { return x.types }
func (x *xlsxRowReader) Close() error {
	if x.closer != nil {
		return x.closer.Close()
	}
	return nil
}

// xlsxRowWriter implements RowWriter for XLSX files.
type xlsxRowWriter struct {
	file    *xlsx.File
	sheet   *xlsx.Sheet
	fields  []string
	types   []string
	written bool
	closer  io.Closer
	writer  io.Writer
}

// NewXLSXRowWriter creates a RowWriter for XLSX files.
// w must be an io.Writer (e.g., *os.File, bytes.Buffer).
func NewXLSXRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("1")
	if err != nil {
		return nil, err
	}
	var closer io.Closer
	if c, ok := w.(io.Closer); ok {
		closer = c
	}
	return &xlsxRowWriter{
		file:   file,
		sheet:  sheet,
		writer: w,
		closer: closer,
	}, nil
}

func (x *xlsxRowWriter) WriteFields(fields []string, types []string) error {
	if x.written {
		return nil
	}
	x.fields = append([]string{}, fields...)
	x.types = append([]string{}, types...)
	header := x.sheet.AddRow()
	for _, f := range x.fields {
		cell := header.AddCell()
		cell.SetString(f)
	}
	x.written = true
	return nil
}

func (x *xlsxRowWriter) WriteRow(row Row) error {
	if !x.written {
		return errors.New("fields not set")
	}
	r := x.sheet.AddRow()
	for i := range x.fields {
		cell := r.AddCell()
		if i < len(row) && row[i] != nil {
			cell.SetValue(row[i])
		} else {
			cell.SetString("")
		}
	}
	return nil
}

func (x *xlsxRowWriter) Close() error {
	if x.writer != nil {
		if err := x.file.Write(x.writer); err != nil {
			return err
		}
	}
	if x.closer != nil {
		return x.closer.Close()
	}
	return nil
}
