import { ApolloClient, InMemoryCache } from "@apollo/client/core/core.cjs";
import { createHttpLink } from "@apollo/client/link/http/http.cjs";

const isServer = typeof window === 'undefined';

function getClient() {
  const httpLink = createHttpLink({
    uri: 'http://localhost:8080/query',
    credentials: 'include',
  });

  return new ApolloClient({
    ssrMode: isServer,
    link: httpLink,
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: {
        fetchPolicy: 'cache-and-network',
      },
    },
  });
}

let clientSingleton: ReturnType<typeof getClient> | null = null;
export function getApolloClient() {
  if (!clientSingleton) {
    clientSingleton = getClient();
  }
  return clientSingleton;
} 