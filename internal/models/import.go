package models

type ImportResult struct {
	TotalCards    int
	ImportedCards int
	UpdatedCards  int
	Errors        []string
}
