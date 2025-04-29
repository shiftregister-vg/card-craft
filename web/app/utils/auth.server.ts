import { redirect } from "@remix-run/node";
import { gql } from "@urql/core";
import { createServerClient } from "../lib/urql.js";

const ME_QUERY = gql`
  query Me {
    me {
      id
      username
      email
    }
  }
`;

export async function requireUser(request: Request) {
  try {
    const cookie = request.headers.get("Cookie");
    if (!cookie || !cookie.includes("token=")) {
      throw redirect("/login");
    }

    const client = createServerClient(request);
    const { data, error } = await client.query(ME_QUERY, {}).toPromise();

    if (error) {
      console.error('Auth error:', error);
      if (error.networkError) {
        throw new Error('Network error during authentication');
      }
      throw redirect("/login");
    }

    if (!data?.me) {
      console.error('No user data found');
      throw redirect("/login");
    }

    return data.me;
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    if (error instanceof Error && error.message === 'Network error during authentication') {
      throw error;
    }
    console.error('Unexpected auth error:', error);
    throw redirect("/login");
  }
} 