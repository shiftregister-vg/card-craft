package models

// CollectionInput represents the input for creating or updating a collection
type CollectionInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Game        string  `json:"game"`
}

// CardInput represents the input for creating or updating a card
type CardInput struct {
	Name     string  `json:"name"`
	Game     string  `json:"game"`
	SetCode  string  `json:"setCode"`
	SetName  string  `json:"setName"`
	Number   string  `json:"number"`
	Rarity   string  `json:"rarity"`
	ImageUrl *string `json:"imageUrl"`
}

// CollectionCardInput represents the input for adding a card to a collection
type CollectionCardInput struct {
	CardID    string  `json:"cardId"`
	Quantity  int     `json:"quantity"`
	Condition *string `json:"condition"`
	IsFoil    *bool   `json:"isFoil"`
	Notes     *string `json:"notes"`
}
