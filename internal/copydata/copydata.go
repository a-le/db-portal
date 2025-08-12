package copydata

import (
	"fmt"
	"io"
)

func CopyData(r RowReader, w RowWriter) (reads, writes int, err error) {
	defer r.Close()
	defer w.Close()

	fields := r.Fields()
	types := r.Types()

	// write fields header when applicable (file destination)
	if w.WriteFields(fields, types) != nil {
		return reads, writes, fmt.Errorf("error writing fields: %w", err)
	}

	// read 1 row from origin, write it to dest
	for {
		row, err := r.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return reads, writes, fmt.Errorf("error during ReadRow: %w, row number: %d", err, reads)
		}
		reads++

		if err := w.WriteRow(row); err != nil {
			return reads, writes, fmt.Errorf("error with row %v during WriteRow: %w", row, err)
		}
		writes++
	}
	return reads, writes, nil
}
