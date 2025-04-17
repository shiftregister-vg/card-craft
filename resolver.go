package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

type Resolver struct{}

// ID is the resolver for the id field.
func (r *cardResolver) ID(ctx context.Context, obj *models.Card) (string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *cardResolver) CreatedAt(ctx context.Context, obj *models.Card) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *cardResolver) UpdatedAt(ctx context.Context, obj *models.Card) (string, error) {
	panic("not implemented")
}

// TotalCount is the resolver for the totalCount field.
func (r *cardSearchResultResolver) TotalCount(ctx context.Context, obj *types.CardSearchResult) (int, error) {
	panic("not implemented")
}

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, email string, password string) (*models.AuthPayload, error) {
	panic("not implemented")
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*models.AuthPayload, error) {
	panic("not implemented")
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context) (*models.AuthPayload, error) {
	panic("not implemented")
}

// CreateCard is the resolver for the createCard field.
func (r *mutationResolver) CreateCard(ctx context.Context, input types.CardInput) (*models.Card, error) {
	panic("not implemented")
}

// UpdateCard is the resolver for the updateCard field.
func (r *mutationResolver) UpdateCard(ctx context.Context, id string, input types.CardInput) (*models.Card, error) {
	panic("not implemented")
}

// DeleteCard is the resolver for the deleteCard field.
func (r *mutationResolver) DeleteCard(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// Card is the resolver for the card field.
func (r *queryResolver) Card(ctx context.Context, id string) (*models.Card, error) {
	panic("not implemented")
}

// CardsByGame is the resolver for the cardsByGame field.
func (r *queryResolver) CardsByGame(ctx context.Context, game string) ([]*models.Card, error) {
	panic("not implemented")
}

// CardsBySet is the resolver for the cardsBySet field.
func (r *queryResolver) CardsBySet(ctx context.Context, game string, setCode string) ([]*models.Card, error) {
	panic("not implemented")
}

// SearchCards is the resolver for the searchCards field.
func (r *queryResolver) SearchCards(ctx context.Context, game *string, setCode *string, rarity *string, name *string, page *int, pageSize *int, sortBy *string, sortOrder *string) (*types.CardSearchResult, error) {
	panic("not implemented")
}

// CardFilters is the resolver for the cardFilters field.
func (r *queryResolver) CardFilters(ctx context.Context, game string) (*types.CardFilters, error) {
	panic("not implemented")
}

// ID is the resolver for the id field.
func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// Card returns generated.CardResolver implementation.
func (r *Resolver) Card() generated.CardResolver { return &cardResolver{r} }

// CardSearchResult returns generated.CardSearchResultResolver implementation.
func (r *Resolver) CardSearchResult() generated.CardSearchResultResolver {
	return &cardSearchResultResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type cardResolver struct{ *Resolver }
type cardSearchResultResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
