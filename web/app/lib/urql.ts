import { createClient, fetchExchange } from '@urql/core';
import { cacheExchange, Data } from '@urql/exchange-graphcache';

export function createUrqlClient(ssrExchange: any) {
  return createClient({
    url: 'http://localhost:8080/query',
    exchanges: [
      cacheExchange({
        keys: {
          User: (data: Data) => data.id?.toString() || null,
          Card: (data: Data) => data.id?.toString() || null,
          Deck: (data: Data) => data.id?.toString() || null,
          DeckCard: (data: Data) => data.id?.toString() || null,
        },
      }),
      ssrExchange,
      fetchExchange,
    ],
    fetchOptions: () => {
      const token = document.cookie
        .split('; ')
        .find(row => row.startsWith('token='))
        ?.split('=')[1];

      return {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        },
      };
    },
  });
} 