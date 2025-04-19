package model

import (
	"github.com/shiftregister-vg/card-craft/internal/models"
)

// These types are used by gqlgen to generate the GraphQL schema
type (
	Card                = models.Card
	CardConnection      = models.CardConnection
	CardEdge            = models.CardEdge
	PageInfo            = models.PageInfo
	Collection          = models.Collection
	CollectionCard      = models.CollectionCard
	ImportResult        = models.ImportResult
	ImportSource        = models.ImportSource
	CardInput           = models.CardInput
	CollectionInput     = models.CollectionInput
	CollectionCardInput = models.CollectionCardInput
)
