import { Form, useLoaderData, useNavigation, redirect } from '@remix-run/react';
import { json, LoaderFunctionArgs } from '@remix-run/node';
import { createClient } from 'urql';
import { cacheExchange, fetchExchange } from '@urql/core';

interface User {
  id: string;
  username: string;
  email: string;
  createdAt: string;
  updatedAt: string;
}

export async function loader({ request }: LoaderFunctionArgs) {
  const cookie = request.headers.get('Cookie');
  if (!cookie) {
    return redirect('/login');
  }

  // Extract token from cookie
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
      query Me {
        me {
          id
          username
          email
          createdAt
          updatedAt
        }
      }
    `, {}).toPromise();

    if (!result.data?.me) {
      return redirect('/login');
    }

    return json<User>(result.data.me);
  } catch (err) {
    return redirect('/login');
  }
}

export default function Account() {
  const user = useLoaderData<User>();
  const navigation = useNavigation();
  const isLoading = navigation.state === 'loading';

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <div className="bg-white shadow sm:rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              Account Information
            </h3>
            <div className="mt-5 border-t border-gray-200">
              <dl className="divide-y divide-gray-200">
                <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4">
                  <dt className="text-sm font-medium text-gray-500">
                    Username
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {user.username}
                  </dd>
                </div>
                <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4">
                  <dt className="text-sm font-medium text-gray-500">
                    Email
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {user.email}
                  </dd>
                </div>
                <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4">
                  <dt className="text-sm font-medium text-gray-500">
                    Member since
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(user.createdAt).toLocaleDateString()}
                  </dd>
                </div>
                <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4">
                  <dt className="text-sm font-medium text-gray-500">
                    Last updated
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(user.updatedAt).toLocaleDateString()}
                  </dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 