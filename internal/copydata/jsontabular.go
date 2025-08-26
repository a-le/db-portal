package copydata

/*
Example of supported json format:
{
 "fields": ["id", "name", "active", "height"],
 "types": ["integer", "string", "boolean", "number"],
 "rows": [
   [1, "Alice", true, 1.72],
   [2, "Bob", false, 1.80],
 ]
}
*/

import (
	"encoding/json"
	"errors"
	"io"
)

type jsontabularData struct {
	Fields []string `json:"fields"`
	Types  []string `json:"types"`
	Rows   [][]any  `json:"rows"`
}

// jsontabularRowReader
type jsontabularRowReader struct {
	data  jsontabularData
	index int
	//closer io.Closer
}

func NewJSONTabularRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	var data jsontabularData
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	return &jsontabularRowReader{
		data:  data,
		index: 0,
	}, nil
}

func (j *jsontabularRowReader) ReadRow() (Row, error) {
	if j.index >= len(j.data.Rows) {
		return nil, io.EOF
	}
	row := j.data.Rows[j.index]
	j.index++
	return row, nil
}

func (j *jsontabularRowReader) Fields() []string { return j.data.Fields }
func (j *jsontabularRowReader) Types() []string  { return j.data.Types }

// jasontabularRowWriter
type jasontabularRowWriter struct {
	enc    *json.Encoder
	fields []string
	types  []string
	rows   [][]any
}

func NewJSONTabularRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}

	return &jasontabularRowWriter{
		enc: json.NewEncoder(w),
	}, nil
}

func (j *jasontabularRowWriter) WriteFields(fields []string, types []string) error {
	j.fields = append([]string{}, fields...)
	j.types = append([]string{}, types...)
	j.rows = [][]any{}
	return nil
}

func (j *jasontabularRowWriter) WriteRow(row Row) (rowsWritten int, err error) {
	j.rows = append(j.rows, append([]any{}, row...))
	return
}

func (j *jasontabularRowWriter) Flush() (rowsWritten int, err error) {
	data := jsontabularData{
		Fields: j.fields,
		Types:  j.types,
		Rows:   j.rows,
	}
	if err = j.enc.Encode(data); err != nil {
		return
	}
	rowsWritten = len(j.rows)
	return
}
