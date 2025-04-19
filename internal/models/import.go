package models

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalCards    int      `json:"totalCards"`
	ImportedCards int      `json:"importedCards"`
	UpdatedCards  int      `json:"updatedCards"`
	Errors        []string `json:"errors"`
}

// ImportSource represents the source of an import operation
type ImportSource struct {
	Source string `json:"source"`
	Format string `json:"format"`
}
