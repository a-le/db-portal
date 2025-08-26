package copydata

// an data EndPoint represents a data origin or destination
type EndPoint struct {
	Type       string `json:"type"`               // "table", "query", or "file"
	DSName     string `json:"dsName,omitempty"`   // Data source name (for "table" and "query")
	DBVendor   string `json:"dbVendor,omitempty"` // Database vendor (for "table" and "query")
	Schema     string `json:"schema,omitempty"`   // Schema name (for "table" and "query")
	Table      string `json:"table,omitempty"`    // Table name (for "table")
	IsNewTable string `json:"newTable,omitempty"` // Whether to create the table (for "table")
	Query      string `json:"query,omitempty"`    // SQL query (for "query")
	Format     string `json:"format,omitempty"`   // File format: "csv", "xlsx", "json", "jsontabular" (for "file")
}

type CopyRequest struct {
	OriginEP EndPoint `json:"origin"`      // Source endpoint
	DestEP   EndPoint `json:"destination"` // Destination endpoint
}
