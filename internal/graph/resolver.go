package graph

import (
	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

// Resolver is the root resolver that contains all the services and stores needed by the resolvers
type Resolver struct {
	authService   *auth.Service
	cardStore     *models.CardStore
	searchService *cards.SearchService
}

// NewResolver creates a new resolver with the given services and stores
func NewResolver(authService *auth.Service, cardStore *models.CardStore, searchService *cards.SearchService) *Resolver {
	return &Resolver{
		authService:   authService,
		cardStore:     cardStore,
		searchService: searchService,
	}
}

// Card returns CardResolver implementation.
func (r *Resolver) Card() generated.CardResolver {
	return &cardResolver{r}
}

// CardSearchResult returns CardSearchResultResolver implementation.
func (r *Resolver) CardSearchResult() generated.CardSearchResultResolver {
	return &cardSearchResultResolver{r}
}

// User returns UserResolver implementation.
func (r *Resolver) User() generated.UserResolver {
	return &userResolver{r}
}

// cardResolver handles field-level resolvers for the Card type
type cardResolver struct{ *Resolver }

// cardSearchResultResolver handles field-level resolvers for the CardSearchResult type
type cardSearchResultResolver struct{ *Resolver }

// userResolver handles field-level resolvers for the User type
type userResolver struct{ *Resolver }
