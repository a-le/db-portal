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

type jasontabularData struct {
	Fields []string `json:"fields"`
	Types  []string `json:"types"`
	Rows   [][]any  `json:"rows"`
}

type jasontabularRowReader struct {
	data   jasontabularData
	index  int
	closer io.Closer
}

func NewJSONTabularRowReader(r io.Reader) (RowReader, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	var data jasontabularData
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	var closer io.Closer
	if c, ok := r.(io.Closer); ok {
		closer = c
	}
	return &jasontabularRowReader{
		data:   data,
		index:  0,
		closer: closer,
	}, nil
}

func (j *jasontabularRowReader) ReadRow() (Row, error) {
	if j.index >= len(j.data.Rows) {
		return nil, io.EOF
	}
	row := j.data.Rows[j.index]
	j.index++
	return row, nil
}

func (j *jasontabularRowReader) Fields() []string { return j.data.Fields }
func (j *jasontabularRowReader) Types() []string  { return j.data.Types }
func (j *jasontabularRowReader) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}

type jasontabularRowWriter struct {
	enc    *json.Encoder
	fields []string
	types  []string
	rows   [][]any
	closer io.Closer
	//written bool
}

func NewJSONTabularRowWriter(w io.Writer) (RowWriter, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}
	var closer io.Closer
	if c, ok := w.(io.Closer); ok {
		closer = c
	}
	return &jasontabularRowWriter{
		enc:    json.NewEncoder(w),
		closer: closer,
	}, nil
}

func (j *jasontabularRowWriter) WriteFields(fields []string, types []string) error {
	// if j.written {
	// 	return nil
	// }
	j.fields = append([]string{}, fields...)
	j.types = append([]string{}, types...)
	j.rows = [][]any{}
	//j.written = true
	return nil
}

func (j *jasontabularRowWriter) WriteRow(row Row) error {
	// if !j.written {
	// 	return errors.New("fields not set")
	// }
	j.rows = append(j.rows, append([]any{}, row...))
	return nil
}

func (j *jasontabularRowWriter) Close() error {
	//if j.written {
	data := jasontabularData{
		Fields: j.fields,
		Types:  j.types,
		Rows:   j.rows,
	}
	if err := j.enc.Encode(data); err != nil {
		return err
	}
	//}
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}
