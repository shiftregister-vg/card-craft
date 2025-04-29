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
  const cookie = request.headers.get('cookie');
  const token = cookie
    ?.split('; ')
    .find(row => row.startsWith('token='))
    ?.split('=')[1];

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return createClient({
    url: GRAPHQL_ENDPOINT,
    exchanges: [cacheExchange, fetchExchange],
    fetchOptions: {
      headers,
      credentials: 'include',
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