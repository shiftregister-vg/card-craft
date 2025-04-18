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
  const cookie = request.headers.get("Cookie");
  if (!cookie || !cookie.includes("token=")) {
    throw redirect("/login");
  }

  const client = createServerClient(request);
  const { data, error } = await client.query(ME_QUERY, {}).toPromise();

  if (error || !data?.me) {
    throw redirect("/login");
  }

  return data.me;
} 