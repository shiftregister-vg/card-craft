import { gql } from '@urql/core';

export const COLLECTION_QUERY = gql`
  query Collection($id: ID!) {
    collection(id: $id) {
      id
      userId
      name
      description
      game
      createdAt
      updatedAt
      cards {
        id
        collectionId
        cardId
        quantity
        condition
        isFoil
        notes
        createdAt
        updatedAt
        card {
          id
          name
          game
          setCode
          setName
          number
          rarity
          imageUrl
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export const MY_COLLECTIONS_QUERY = gql`
  query MyCollections {
    myCollections {
      id
      name
      description
      game
      createdAt
      updatedAt
      cards {
        id
        quantity
      }
    }
  }
`;

export const CARDS_BY_GAME_QUERY = gql`
  query CardsByGame($game: String!, $first: Int, $after: String) {
    cardsByGame(game: $game, first: $first, after: $after) {
      edges {
        node {
          id
          name
          game
          setCode
          setName
          number
          rarity
          imageUrl
          createdAt
          updatedAt
        }
        cursor
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export const CREATE_COLLECTION_MUTATION = gql`
  mutation CreateCollection($input: CreateCollectionInput!) {
    createCollection(input: $input) {
      id
      name
      description
      game
      createdAt
      updatedAt
    }
  }
`;

export const ADD_CARD_TO_COLLECTION_MUTATION = gql`
  mutation AddCardToCollection($collectionId: ID!, $input: CollectionCardInput!) {
    addCardToCollection(collectionId: $collectionId, input: $input) {
      id
      collectionId
      cardId
      quantity
      condition
      isFoil
      notes
      createdAt
      updatedAt
      card {
        id
        name
        game
        setCode
        setName
        number
        rarity
        imageUrl
        createdAt
        updatedAt
      }
    }
  }
`;

export const REMOVE_CARD_FROM_COLLECTION_MUTATION = gql`
  mutation RemoveCardFromCollection($id: ID!) {
    removeCardFromCollection(id: $id)
  }
`; 