import { ActionFunctionArgs, redirect } from "@remix-run/node";

export async function action({ request }: ActionFunctionArgs) {
  return redirect("/login", {
    headers: {
      "Set-Cookie": "token=; Path=/; HttpOnly; Max-Age=0"
    },
  });
}

export async function loader() {
  return redirect("/login");
}

export default function Logout() {
  return null;
} 