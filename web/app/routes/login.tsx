import { Form, useActionData, useNavigation, useSearchParams, redirect } from '@remix-run/react';
import { json, ActionFunctionArgs } from '@remix-run/node';
import { gql } from '@urql/core';
import { createServerClient } from '../lib/urql.js';

interface LoginResponse {
  error?: string;
  user?: {
    id: string;
    email: string;
  };
}

const LOGIN_MUTATION = gql`
  mutation Login($identifier: String!, $password: String!) {
    login(identifier: $identifier, password: $password) {
      token
      user {
        id
        email
      }
    }
  }
`;

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const identifier = formData.get("identifier") as string;
  const password = formData.get("password") as string;

  try {
    const client = createServerClient(request);
    const result = await client.mutation(LOGIN_MUTATION, {
      identifier,
      password,
    }).toPromise();

    if (result.error) {
      console.error('Login error:', result.error);
      return json<LoginResponse>(
        { error: result.error.message },
        { status: 400 }
      );
    }

    if (!result.data?.login?.token) {
      console.error('No token in response:', result);
      return json<LoginResponse>(
        { error: "Invalid credentials" },
        { status: 401 }
      );
    }

    const headers = new Headers();
    headers.append("Set-Cookie", `token=${result.data.login.token}; Path=/; HttpOnly; SameSite=Lax`);

    return redirect("/", {
      headers,
    });
  } catch (error) {
    console.error('Login error:', error);
    return json<LoginResponse>(
      { error: "Invalid credentials" },
      { status: 401 }
    );
  }
}

export default function Login() {
  const actionData = useActionData<LoginResponse>();
  const navigation = useNavigation();
  const [searchParams] = useSearchParams();
  const isLoading = navigation.state === "submitting";

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
                Email or Username
              </label>
              <input
                id="identifier"
                name="identifier"
                type="text"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm bg-white"
                placeholder="Email or Username"
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
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm bg-white"
                placeholder="Password"
                disabled={isLoading}
              />
            </div>
          </div>

          {actionData?.error && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="flex">
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">
                    {actionData.error}
                  </h3>
                </div>
              </div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              {isLoading ? "Signing in..." : "Sign in"}
            </button>
          </div>
        </Form>
      </div>
    </div>
  );
} 