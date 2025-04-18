import { Client, cacheExchange, createClient, fetchExchange } from '@urql/core';
import { devtoolsExchange } from '@urql/devtools';

const GRAPHQL_ENDPOINT = 'http://localhost:8080/query';

// Create a client-side client instance
export const client = typeof document !== 'undefined'
  ? createClient({
      url: GRAPHQL_ENDPOINT,
      exchanges: [devtoolsExchange, cacheExchange, fetchExchange],
      fetchOptions: () => {
        const token = document.cookie
          .split('; ')
          .find(row => row.startsWith('token='))
          ?.split('=')[1];
        return {
          headers: { Authorization: token ? `Bearer ${token}` : '' },
        };
      },
    })
  : null;

// Create a server-side client instance
export function createServerClient(request: Request): Client {
  const token = request.headers.get('cookie')
    ?.split('; ')
    .find(row => row.startsWith('token='))
    ?.split('=')[1];

  return createClient({
    url: GRAPHQL_ENDPOINT,
    exchanges: [cacheExchange, fetchExchange],
    fetchOptions: {
      headers: { Authorization: token ? `Bearer ${token}` : '' },
    },
  });
}

// Export the createUrqlClient function for SSR
export function createUrqlClient() {
  return {
    url: GRAPHQL_ENDPOINT,
    exchanges: [cacheExchange, fetchExchange],
  };
} 