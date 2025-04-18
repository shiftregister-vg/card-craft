import { gql } from '@urql/core';

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
        condition
        isFoil
        notes
        card {
          id
          name
          game
          setCode
          setName
          number
          rarity
          imageUrl
        }
      }
    }
  }
`;

export const CREATE_COLLECTION_MUTATION = gql`
  mutation CreateCollection($input: CollectionInput!) {
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
      quantity
      condition
      isFoil
      notes
      card {
        id
        name
        game
        setCode
        setName
        number
        rarity
        imageUrl
      }
    }
  }
`;

export const UPDATE_COLLECTION_CARD_MUTATION = gql`
  mutation UpdateCollectionCard($id: ID!, $input: CollectionCardInput!) {
    updateCollectionCard(id: $id, input: $input) {
      id
      quantity
      condition
      isFoil
      notes
    }
  }
`;

export const REMOVE_CARD_FROM_COLLECTION_MUTATION = gql`
  mutation RemoveCardFromCollection($id: ID!) {
    removeCardFromCollection(id: $id)
  }
`; 