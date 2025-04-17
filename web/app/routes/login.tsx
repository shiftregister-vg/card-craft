import { Form, useActionData, useNavigation, useSearchParams, redirect } from '@remix-run/react';
import { json, ActionFunctionArgs } from '@remix-run/node';
import { createClient } from 'urql';
import { cacheExchange, fetchExchange } from '@urql/core';
import { useEffect } from 'react';
import { useAuth } from '../context/auth.js';

interface LoginResponse {
  error?: string;
  token?: string;
  user?: {
    id: string;
    username: string;
    email: string;
  };
}

const LOGIN_MUTATION = `
  mutation Login($identifier: String!, $password: String!) {
    login(identifier: $identifier, password: $password) {
      token
      user {
        id
        username
        email
      }
    }
  }
`;

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const identifier = formData.get('identifier') as string;
  const password = formData.get('password') as string;

  // Get the redirect URL from the form data or default to dashboard
  const redirectTo = formData.get('redirectTo') as string || '/dashboard';

  const client = createClient({
    url: 'http://localhost:8080/query',
    exchanges: [cacheExchange, fetchExchange],
  });

  try {
    const result = await client.mutation(LOGIN_MUTATION, {
      identifier,
      password,
    }).toPromise();

    if (result.error) {
      return json<LoginResponse>(
        { error: result.error.message },
        { status: 400 }
      );
    }

    if (!result.data?.login?.token) {
      return json<LoginResponse>(
        { error: 'Invalid credentials' },
        { status: 401 }
      );
    }

    // Set the token in a cookie and include user data in the response
    const headers = new Headers();
    headers.append('Set-Cookie', `token=${result.data.login.token}; Path=/; HttpOnly`);

    return json(
      { user: result.data.login.user },
      {
        headers,
        status: 200,
      }
    );
  } catch (err) {
    return json<LoginResponse>(
      { error: 'Login failed. Please check your credentials.' },
      { status: 400 }
    );
  }
}

export default function Login() {
  const actionData = useActionData<LoginResponse>();
  const navigation = useNavigation();
  const [searchParams] = useSearchParams();
  const { setUser } = useAuth();
  const isLoading = navigation.state === 'submitting';

  useEffect(() => {
    if (actionData?.user) {
      setUser(actionData.user);
      const redirectTo = searchParams.get('redirectTo') || '/dashboard';
      window.location.href = redirectTo;
    }
  }, [actionData, setUser, searchParams]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Use your username or email to sign in
          </p>
        </div>
        <Form method="post" className="mt-8 space-y-6">
          <input
            type="hidden"
            name="redirectTo"
            value={searchParams.get('redirectTo') || '/dashboard'}
          />
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="identifier" className="sr-only">
                Username or Email
              </label>
              <input
                id="identifier"
                name="identifier"
                type="text"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 bg-white rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Username or Email"
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
              {isLoading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>
        </Form>
      </div>
    </div>
  );
} 