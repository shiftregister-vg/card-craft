import gql from 'graphql-tag';

export const LOGIN_MUTATION = gql`
  mutation Login($email: String!, $password: String!) {
    login(input: { email: $email, password: $password }) {
      token
      refreshToken
      user {
        id
        email
      }
    }
  }
`;

export const REGISTER_MUTATION = gql`
  mutation Register($email: String!, $password: String!) {
    register(input: { email: $email, password: $password }) {
      token
      refreshToken
      user {
        id
        email
      }
    }
  }
`;

export const ME_QUERY = gql`
  query Me {
    me {
      id
      email
    }
  }
`; 