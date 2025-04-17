import { Form, useActionData, useNavigation, redirect } from '@remix-run/react';
import { json, ActionFunctionArgs, LoaderFunctionArgs } from '@remix-run/node';
import { createClient } from 'urql';
import { cacheExchange, fetchExchange } from '@urql/core';

interface RegisterResponse {
  error?: string;
  token?: string;
  user?: {
    id: string;
    username: string;
    email: string;
  };
}

const REGISTER_MUTATION = `
  mutation Register($username: String!, $email: String!, $password: String!) {
    register(username: $username, email: $email, password: $password) {
      token
      user {
        id
        username
        email
      }
    }
  }
`;

export async function loader({ request }: LoaderFunctionArgs) {
  const cookie = request.headers.get('Cookie');
  if (!cookie) {
    return null;
  }

  // Extract token from cookie
  const token = cookie
    .split('; ')
    .find(row => row.startsWith('token='))
    ?.split('=')[1];

  if (!token) {
    return null;
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
          email
        }
      }
    `, {}).toPromise();

    if (result.data?.me) {
      return redirect('/account');
    }
  } catch (err) {
    // If there's an error (e.g., invalid token), let them register
    return null;
  }

  return null;
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const username = formData.get('username') as string;
  const email = formData.get('email') as string;
  const password = formData.get('password') as string;

  const client = createClient({
    url: 'http://localhost:8080/query',
    exchanges: [cacheExchange, fetchExchange],
    fetchOptions: {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    },
  });

  try {
    const result = await client.mutation(REGISTER_MUTATION, {
      username,
      email,
      password,
    }).toPromise();

    if (result.error) {
      return json<RegisterResponse>(
        { error: result.error.message },
        { status: 400 }
      );
    }

    if (!result.data?.register?.token) {
      return json<RegisterResponse>(
        { error: 'Registration failed' },
        { status: 400 }
      );
    }

    // Set the token in a cookie
    const headers = new Headers();
    headers.append('Set-Cookie', `token=${result.data.register.token}; Path=/; HttpOnly; SameSite=Lax`);

    return redirect('/account', {
      headers,
    });
  } catch (err) {
    return json<RegisterResponse>(
      { error: 'Registration failed. Please try again.' },
      { status: 400 }
    );
  }
}

export default function Register() {
  const actionData = useActionData<RegisterResponse>();
  const navigation = useNavigation();
  const isLoading = navigation.state === 'submitting';

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Create your account
          </h2>
        </div>
        <Form method="post" className="mt-8 space-y-6">
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="username" className="sr-only">
                Username
              </label>
              <input
                id="username"
                name="username"
                type="text"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 bg-white rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Username"
                disabled={isLoading}
              />
            </div>
            <div>
              <label htmlFor="email" className="sr-only">
                Email address
              </label>
              <input
                id="email"
                name="email"
                type="email"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 bg-white focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Email address"
                disabled={isLoading}
              />
            </div>
            <div>
              <label htmlFor="password" className="sr-only">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 bg-white rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Password"
                disabled={isLoading}
              />
            </div>
          </div>

          {actionData?.error && (
            <div className="text-red-500 text-sm text-center">
              {actionData.error}
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className={`group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white ${
                isLoading
                  ? 'bg-indigo-400 cursor-not-allowed'
                  : 'bg-indigo-600 hover:bg-indigo-700'
              } focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500`}
            >
              {isLoading ? 'Creating account...' : 'Create account'}
            </button>
          </div>
        </Form>
      </div>
    </div>
  );
} 