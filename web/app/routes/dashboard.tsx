import { Link, useLoaderData } from '@remix-run/react';
import { json, LoaderFunctionArgs, redirect } from '@remix-run/node';
import { createClient } from 'urql';
import { cacheExchange, fetchExchange } from '@urql/core';

interface DashboardData {
  deckCount: number;
  collectionCount: number;
}

export async function loader({ request }: LoaderFunctionArgs) {
  const cookie = request.headers.get('Cookie');
  if (!cookie) {
    return redirect('/login');
  }

  const token = cookie
    .split('; ')
    .find(row => row.startsWith('token='))
    ?.split('=')[1];

  if (!token) {
    return redirect('/login');
  }

  const client = createClient({
    url: 'http://localhost:8080/query',
    exchanges: [cacheExchange, fetchExchange],
    fetchOptions: {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    },
  });

  try {
    const result = await client.query(`
      query DashboardStats {
        myDecks {
          id
        }
        myCollections {
          id
        }
      }
    `, {}).toPromise();

    return json<DashboardData>({
      deckCount: result.data?.myDecks?.length || 0,
      collectionCount: result.data?.myCollections?.length || 0,
    });
  } catch (err) {
    return redirect('/login');
  }
}

export default function Dashboard() {
  const { deckCount, collectionCount } = useLoaderData<DashboardData>();

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Dashboard</h1>
        
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {/* Decks Card */}
          <Link
            to="/decks"
            className="bg-white overflow-hidden shadow rounded-lg hover:shadow-lg transition-shadow duration-200"
          >
            <div className="p-6">
              <div className="flex items-center">
                <div className="flex-shrink-0 bg-indigo-500 rounded-md p-3">
                  <svg
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      Total Decks
                    </dt>
                    <dd className="flex items-baseline">
                      <div className="text-2xl font-semibold text-gray-900">
                        {deckCount}
                      </div>
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </Link>

          {/* Collections Card */}
          <Link
            to="/collections"
            className="bg-white overflow-hidden shadow rounded-lg hover:shadow-lg transition-shadow duration-200"
          >
            <div className="p-6">
              <div className="flex items-center">
                <div className="flex-shrink-0 bg-green-500 rounded-md p-3">
                  <svg
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      Total Collections
                    </dt>
                    <dd className="flex items-baseline">
                      <div className="text-2xl font-semibold text-gray-900">
                        {collectionCount}
                      </div>
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </Link>

          {/* Create New Card */}
          <Link
            to="/decks/new"
            className="bg-white overflow-hidden shadow rounded-lg hover:shadow-lg transition-shadow duration-200"
          >
            <div className="p-6">
              <div className="flex items-center">
                <div className="flex-shrink-0 bg-blue-500 rounded-md p-3">
                  <svg
                    className="h-6 w-6 text-white"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                    />
                  </svg>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      Create New
                    </dt>
                    <dd className="flex items-baseline">
                      <div className="text-2xl font-semibold text-gray-900">
                        Deck
                      </div>
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </Link>
        </div>
      </div>
    </div>
  );
} 