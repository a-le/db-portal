package copydata

import (
	"fmt"
	"io"
)

// CopyData reads from RowReader and writes to RowWriter, returns number of reads and writes.
func CopyData(r RowReader, w RowWriter) (reads int, writes int, err error) {

	fields := r.Fields()
	types := r.Types()

	// write fields header when applicable (file destination)
	if err = w.WriteFields(fields, types); err != nil {
		err = fmt.Errorf("error writing fields: %w", err)
		return
	}

	// read 1 row from origin, write it to dest
	// if dest is a table, writes are batched in multi-values INSERTs
	for {
		var row Row
		row, err = r.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		reads++

		var rowsWritten int
		if rowsWritten, err = w.WriteRow(row); err != nil {
			err = fmt.Errorf("error with row %v during WriteRow: %w", row, err)
			return
		} else {
			writes += rowsWritten
		}
	}

	// Flush (write last batch)
	var rowsWritten int
	if rowsWritten, err = w.Flush(); err != nil {
		err = fmt.Errorf("error during flush: %w", err)
		return
	}
	writes += rowsWritten

	return
}
