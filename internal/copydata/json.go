package copydata

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// jsonRowReader
type jsonRowReader struct {
	scanner   *bufio.Scanner
	fields    []string
	types     []string
	firstRow  map[string]any // buffer for the first row
	firstRead bool           // has the first row been read?
}

func NewJSONRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	scanner := bufio.NewScanner(r)

	// Read the first line to get fields
	var firstRow map[string]any
	var fields []string

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return &jsonRowReader{
			scanner:   scanner,
			fields:    fields,
			firstRow:  nil,
			firstRead: false,
		}, nil
	}
	line := scanner.Bytes()

	// Use a Decoder to preserve field order
	dec := json.NewDecoder(bytes.NewReader(line))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return nil, errors.New("expected JSON object")
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if key, ok := t.(string); ok {
			fields = append(fields, key)
			// skip value
			if err := dec.Decode(&json.RawMessage{}); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("expected string key")
		}
	}
	// Unmarshal values as before
	if err := json.Unmarshal(line, &firstRow); err != nil {
		return nil, err
	}

	return &jsonRowReader{
		scanner:   scanner,
		fields:    fields,
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

// jsonRowWriter
type jsonRowWriter struct {
	enc     *json.Encoder
	fields  []string
	written bool
}

func NewJSONRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}
	return &jsonRowWriter{
		enc: json.NewEncoder(w),
	}, nil
}

func (j *jsonRowWriter) WriteFields(fields []string, types []string) error {
	j.fields = append([]string{}, fields...)
	j.written = true
	return nil
}

func (j *jsonRowWriter) WriteRow(row Row) (rowsWritten int, err error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range j.fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, _ := json.Marshal(f)
		buf.Write(key)
		buf.WriteByte(':')
		val := []byte("null")
		if i < len(row) {
			val, _ = json.Marshal(row[i])
		}
		buf.Write(val)
	}
	buf.WriteByte('}')
	err = j.enc.Encode(json.RawMessage(buf.Bytes()))
	if err != nil {
		return
	}
	rowsWritten = 1
	return
}

func (j *jsonRowWriter) Flush() (rowsWritten int, err error) { return }
