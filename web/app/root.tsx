import { Provider as UrqlProvider } from 'urql';
import { json, LoaderFunctionArgs } from '@remix-run/node';
import {
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useLoaderData,
} from '@remix-run/react';
import type { LinksFunction } from "@remix-run/node";
import { createClient, ssrExchange, cacheExchange, fetchExchange } from '@urql/core';
import { AuthProvider, useAuth } from './context/auth.js';
import Navigation from './components/Navigation.js';
import React from 'react';

import "./tailwind.css";

export const links: LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export async function loader({ request }: LoaderFunctionArgs) {
  const ssr = ssrExchange();

  // Get the token from cookies
  const cookie = request.headers.get('Cookie');
  const token = cookie
    ?.split('; ')
    .find(row => row.startsWith('token='))
    ?.split('=')[1];

  const client = createClient({
    url: 'http://localhost:8080/query',
    exchanges: [cacheExchange, ssr, fetchExchange],
    fetchOptions: {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
      },
    },
  });

  let user = null;
  if (token) {
    try {
      const result = await client.query(`
        query Me {
          me {
            id
            username
            email
          }
        }
      `, {}).toPromise();

      if (result.data?.me) {
        user = result.data.me;
      }
    } catch (err) {
      console.error('Error fetching user data:', err);
      // Don't throw here, just return null user
    }
  }

  return json({
    urqlState: ssr.extractData(),
    user,
    isAuthenticated: !!user,
  });
}

function ClientUrqlProvider({ children }: { children: React.ReactNode }) {
  const { urqlState } = useLoaderData<typeof loader>();
  const ssr = ssrExchange({ initialState: urqlState });
  
  const client = createClient({
    url: 'http://localhost:8080/query',
    exchanges: [cacheExchange, ssr, fetchExchange],
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

  return <UrqlProvider value={client}>{children}</UrqlProvider>;
}

function AppWithAuth() {
  const { user, isAuthenticated } = useLoaderData<typeof loader>();
  const { setUser } = useAuth();

  // Set the user in the auth context when it changes
  React.useEffect(() => {
    setUser(user);
  }, [user, setUser]);

  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        <ClientUrqlProvider>
          {isAuthenticated && <Navigation />}
          <Outlet />
        </ClientUrqlProvider>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppWithAuth />
    </AuthProvider>
  );
}
