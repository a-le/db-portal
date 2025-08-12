package copydata

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

type jsonRowReader struct {
	scanner   *bufio.Scanner
	fields    []string
	types     []string
	closer    io.Closer
	firstRow  map[string]any // buffer for the first row
	firstRead bool           // has the first row been read?
}

func NewJSONRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	scanner := bufio.NewScanner(r)
	var closer io.Closer
	if c, ok := r.(io.Closer); ok {
		closer = c
	}

	// Read the first line to get fields
	var firstRow map[string]any
	var fields []string
	// for scanner.Scan() {
	// 	line := scanner.Bytes()
	// 	if len(line) == 0 {
	// 		continue // skip empty lines
	// 	}
	// 	if err := json.Unmarshal(line, &firstRow); err != nil {
	// 		return nil, err
	// 	}
	// 	for k := range firstRow {
	// 		fields = append(fields, k)
	// 	}
	// 	break
	// }
	scanner.Scan()
	line := scanner.Bytes()
	if err := json.Unmarshal(line, &firstRow); err != nil {
		return nil, err
	}
	for k := range firstRow {
		fields = append(fields, k)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// If no data, fields will be nil, that's ok

	return &jsonRowReader{
		scanner:   scanner,
		fields:    fields,
		closer:    closer,
		firstRow:  firstRow,
		firstRead: false,
	}, nil
}

func (j *jsonRowReader) ReadRow() (Row, error) {
	// Serve the buffered first row first
	if !j.firstRead && j.firstRow != nil {
		j.firstRead = true
		row := make(Row, len(j.fields))
		for i, f := range j.fields {
			row[i] = j.firstRow[f]
		}
		return row, nil
	}
	for j.scanner.Scan() {
		line := j.scanner.Bytes()
		if len(line) == 0 {
			continue // skip empty lines
		}
		var obj map[string]any
		if err := json.Unmarshal(line, &obj); err != nil {
			return nil, err
		}
		row := make(Row, len(j.fields))
		for i, f := range j.fields {
			row[i] = obj[f]
		}
		return row, nil
	}
	if err := j.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func (j *jsonRowReader) Fields() []string { return j.fields }
func (j *jsonRowReader) Types() []string  { return j.types }
func (j *jsonRowReader) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}

type jsonRowWriter struct {
	enc     *json.Encoder
	fields  []string
	closer  io.Closer
	written bool
}

func NewJSONRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}
	var closer io.Closer
	if c, ok := w.(io.Closer); ok {
		closer = c
	}
	return &jsonRowWriter{
		enc:    json.NewEncoder(w),
		closer: closer,
	}, nil
}

func (j *jsonRowWriter) WriteFields(fields []string, types []string) error {
	// if j.written {
	// 	return nil
	// }
	j.fields = append([]string{}, fields...)
	j.written = true
	return nil
}

func (j *jsonRowWriter) WriteRow(row Row) error {
	// if j.fields == nil {
	// 	return errors.New("fields not set")
	// }
	obj := make(map[string]any, len(j.fields))
	for i, f := range j.fields {
		if i < len(row) {
			obj[f] = row[i]
		} else {
			obj[f] = nil
		}
	}
	return j.enc.Encode(obj)
}

func (j *jsonRowWriter) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}
