package copydata

import (
	"errors"
	"fmt"
	"io"

	"github.com/tealeg/xlsx"
)

// xlsxRowReader
type xlsxRowReader struct {
	file     *xlsx.File
	sheet    *xlsx.Sheet
	fields   []string
	types    []string
	rowIndex int
}

func NewXLSXRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}

	// Read all data from reader into a byte slice
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Open XLSX from byte slice
	file, err := xlsx.OpenBinary(data)
	if err != nil {
		return nil, err
	}

	if len(file.Sheets) == 0 {
		return nil, errors.New("no sheets found in XLSX file")
	}

	// Use the first sheet
	sheet := file.Sheets[0]
	if len(sheet.Rows) == 0 {
		return &xlsxRowReader{
			file:     file,
			sheet:    sheet,
			fields:   nil,
			types:    nil,
			rowIndex: 0,
		}, nil
	}

	// Use first row as field names
	firstRow := sheet.Rows[0]
	fields := make([]string, len(firstRow.Cells))
	for i, cell := range firstRow.Cells {
		if cell != nil {
			fields[i] = cell.String()
		} else {
			fields[i] = ""
		}
	}

	return &xlsxRowReader{
		file:     file,
		sheet:    sheet,
		fields:   fields,
		types:    nil, // XLSX doesn't provide type information
		rowIndex: 1,   // Start from second row (skip header)
	}, nil
}

func (x *xlsxRowReader) ReadRow() (Row, error) {
	if x.sheet == nil || x.rowIndex >= len(x.sheet.Rows) {
		return nil, io.EOF
	}

	xlsxRow := x.sheet.Rows[x.rowIndex]
	x.rowIndex++

	row := make(Row, len(x.fields))
	for i := range x.fields {
		if i < len(xlsxRow.Cells) && xlsxRow.Cells[i] != nil {
			cell := xlsxRow.Cells[i]
			// Try to get the value in the most appropriate type
			if cell.Type() == xlsx.CellTypeNumeric {
				if val, err := cell.Float(); err == nil {
					row[i] = val
				} else {
					row[i] = cell.String()
				}
			} else if cell.Type() == xlsx.CellTypeBool {
				row[i] = cell.Bool()
			} else {
				row[i] = cell.String()
			}
		} else {
			row[i] = nil
		}
	}

	return row, nil
}

func (x *xlsxRowReader) Fields() []string { return x.fields }
func (x *xlsxRowReader) Types() []string  { return x.types }

// xlsxRowWriter
type xlsxRowWriter struct {
	file    *xlsx.File
	sheet   *xlsx.Sheet
	fields  []string
	writer  io.Writer
	written bool
}

func NewXLSXRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return nil, err
	}

	return &xlsxRowWriter{
		file:   file,
		sheet:  sheet,
		writer: w,
	}, nil
}

func (x *xlsxRowWriter) WriteFields(fields []string, types []string) error {
	x.fields = append([]string{}, fields...)

	// Write header row
	if len(x.fields) > 0 {
		row := x.sheet.AddRow()
		for _, field := range x.fields {
			cell := row.AddCell()
			cell.Value = field
		}
		x.written = true
	}

	return nil
}

func (x *xlsxRowWriter) WriteRow(row Row) (rowsWritten int, err error) {
	xlsxRow := x.sheet.AddRow()
	for i := range x.fields {
		cell := xlsxRow.AddCell()
		if i < len(row) && row[i] != nil {
			switch v := row[i].(type) {
			case string:
				cell.Value = v
			case int:
				cell.SetInt64(int64(v))
			case int8:
				cell.SetInt64(int64(v))
			case int16:
				cell.SetInt64(int64(v))
			case int32:
				cell.SetInt64(int64(v))
			case int64:
				cell.SetInt64(v)
			case float32:
				cell.SetFloat(float64(v))
			case float64:
				cell.SetFloat(v)
			case bool:
				cell.SetBool(v)
			default:
				cell.Value = fmt.Sprintf("%v", v)
			}
		}
	}
	return 1, nil
}

func (x *xlsxRowWriter) Flush() (rowsWritten int, err error) {
	// Write the entire file to the writer
	err = x.file.Write(x.writer)
	return 0, err
}
