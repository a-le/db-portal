package internaldb

type Vendor struct {
	Name string `json:"name"`
}

func (s *Store) GetAllVendors() ([]Vendor, error) {
	query := `
    SELECT name
    FROM vendor
    ORDER BY 1
    `

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendors []Vendor
	for rows.Next() {
		var vendor Vendor
		if err := rows.Scan(&vendor.Name); err != nil {
			return nil, err
		}
		vendors = append(vendors, vendor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vendors, nil
}
