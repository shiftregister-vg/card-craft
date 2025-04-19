import { json, LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, Link, useActionData, useLoaderData, useNavigation } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { client, createServerClient } from '../lib/urql.js';
import { requireUser } from '../utils/auth.server.js';
import { MY_COLLECTIONS_QUERY, CREATE_COLLECTION_MUTATION } from '../graphql/collections.js';

interface Collection {
  id: string;
  name: string;
  description: string | null;
  game: string;
  cards: Array<{ id: string }>;
}

interface LoaderData {
  collections: Collection[];
}

interface ActionData {
  error?: string;
  collection?: Collection;
}

export async function loader({ request }: LoaderFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  const { data, error } = await serverClient.query(MY_COLLECTIONS_QUERY, {}).toPromise();
  
  if (error) {
    console.error('GraphQL error:', error);
    return json<LoaderData>({ collections: [] });
  }
  
  if (!data || !data.myCollections) {
    return json<LoaderData>({ collections: [] });
  }
  
  return json<LoaderData>({ collections: data.myCollections });
}

export async function action({ request }: ActionFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  const formData = await request.formData();
  const name = formData.get('name') as string;
  const description = formData.get('description') as string;
  const game = formData.get('game') as string;

  const { data, error } = await serverClient.mutation(CREATE_COLLECTION_MUTATION, {
    input: {
      name,
      description,
      game,
    },
  }).toPromise();

  if (error) {
    return json<ActionData>({ error: error.message });
  }

  return json<ActionData>({ collection: data.createCollection });
}

export default function Collections() {
  const { collections } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const [showCreateForm, setShowCreateForm] = useState(false);
  
  const isSubmitting = navigation.state === 'submitting';
  
  useEffect(() => {
    if (actionData?.collection) {
      setShowCreateForm(false);
    }
  }, [actionData]);
  
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-3xl font-bold text-gray-900">My Collections</h1>
            <button
              onClick={() => setShowCreateForm(!showCreateForm)}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              {showCreateForm ? 'Cancel' : 'Create Collection'}
            </button>
          </div>
          
          {showCreateForm && (
            <div className="bg-white shadow sm:rounded-lg mb-6">
              <div className="px-4 py-5 sm:p-6">
                <Form method="post" className="space-y-4">
                  {actionData?.error && (
                    <div className="text-red-600">{actionData.error}</div>
                  )}
                  <div>
                    <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                      Name
                    </label>
                    <input
                      type="text"
                      name="name"
                      id="name"
                      required
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    />
                  </div>
                  <div>
                    <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <textarea
                      name="description"
                      id="description"
                      rows={3}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    />
                  </div>
                  <div>
                    <label htmlFor="game" className="block text-sm font-medium text-gray-700">
                      Game
                    </label>
                    <select
                      name="game"
                      id="game"
                      required
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    >
                      <option value="">Select a game</option>
                      <option value="POKEMON">Pokémon</option>
                      <option value="LORCANA">Disney Lorcana</option>
                      <option value="STARWARS">Star Wars: Unlimited</option>
                    </select>
                  </div>
                  <div>
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                    >
                      {isSubmitting ? 'Creating...' : 'Create Collection'}
                    </button>
                  </div>
                </Form>
              </div>
            </div>
          )}
          
          {collections.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-lg shadow">
              <h3 className="mt-2 text-sm font-medium text-gray-900">No collections</h3>
              <p className="mt-1 text-sm text-gray-500">Get started by creating a new collection.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {collections.map((collection) => (
                <div
                  key={collection.id}
                  className="relative rounded-lg border border-gray-300 bg-white px-6 py-5 shadow-sm hover:border-gray-400"
                >
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{collection.name}</h3>
                      <p className="mt-1 text-sm text-gray-500">{collection.description}</p>
                    </div>
                    <span className="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
                      {collection.game}
                    </span>
                  </div>
                  <div className="mt-4">
                    <p className="text-sm text-gray-500">
                      {collection.cards.length} cards
                    </p>
                  </div>
                  <Link
                    to={`/collections/${collection.id}`}
                    className="absolute inset-0 focus:outline-none"
                  >
                    <span className="sr-only">View collection</span>
                  </Link>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
} 