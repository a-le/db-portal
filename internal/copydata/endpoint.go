package copydata

type DataEndpoint struct {
	Type   string `json:"type"` // "table", "query", "file"
	DSName string `json:"dsName,omitempty"`
	Schema string `json:"schema,omitempty"`
	Table  string `json:"table,omitempty"`
	Query  string `json:"query,omitempty"`
	Format string `json:"format,omitempty"` // "csv", "xlsx", "json"
}

type DataTransferRequest struct {
	Origin      DataEndpoint `json:"origin"`
	Destination DataEndpoint `json:"destination"`
}
