import { json, LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, Link, useActionData, useLoaderData, useNavigation, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { createServerClient } from '../lib/urql.js';
import { requireUser } from '../utils/auth.server.js';
import { COLLECTION_QUERY, ADD_CARD_TO_COLLECTION_MUTATION, REMOVE_CARD_FROM_COLLECTION_MUTATION } from '../graphql/collections.js';

interface CollectionCard {
  id: string;
  card: {
    id: string;
    name: string;
    game: string;
    setCode: string;
    setName: string;
    number: string;
    rarity: string;
    imageUrl: string;
  };
  quantity: number;
  condition: string | null;
  isFoil: boolean | null;
  notes: string | null;
}

interface Collection {
  id: string;
  name: string;
  description: string | null;
  game: string;
  cards: CollectionCard[];
}

interface LoaderData {
  collection: Collection;
}

interface ActionData {
  error?: string;
  success?: boolean;
}

export async function loader({ request, params }: LoaderFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  const { data, error } = await serverClient.query(COLLECTION_QUERY, {
    id: params.id,
  }).toPromise();
  
  if (error) {
    console.error('GraphQL error:', error);
    throw new Response('Collection not found', { status: 404 });
  }
  
  if (!data || !data.collection) {
    throw new Response('Collection not found', { status: 404 });
  }
  
  return json<LoaderData>({ collection: data.collection });
}

export async function action({ request, params }: ActionFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  const formData = await request.formData();
  const action = formData.get('action') as string;

  if (action === 'add-card') {
    const cardId = formData.get('cardId') as string;
    const quantity = parseInt(formData.get('quantity') as string);
    const condition = formData.get('condition') as string;
    const isFoil = formData.get('isFoil') === 'true';
    const notes = formData.get('notes') as string;

    const { error } = await serverClient.mutation(ADD_CARD_TO_COLLECTION_MUTATION, {
      collectionId: params.id,
      input: {
        cardId,
        quantity,
        condition,
        isFoil,
        notes,
      },
    }).toPromise();

    if (error) {
      return json<ActionData>({ error: error.message });
    }

    return json<ActionData>({ success: true });
  }

  if (action === 'remove-card') {
    const cardId = formData.get('cardId') as string;

    const { error } = await serverClient.mutation(REMOVE_CARD_FROM_COLLECTION_MUTATION, {
      id: cardId,
    }).toPromise();

    if (error) {
      return json<ActionData>({ error: error.message });
    }

    return json<ActionData>({ success: true });
  }

  return json<ActionData>({ error: 'Invalid action' });
}

export default function CollectionDetail() {
  const { collection } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const [showAddCardForm, setShowAddCardForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  
  const isSubmitting = navigation.state === 'submitting';
  
  useEffect(() => {
    if (actionData?.success) {
      setShowAddCardForm(false);
    }
  }, [actionData]);

  const filteredCards = collection.cards.filter(card => 
    card.card.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    card.card.setName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    card.card.setCode.toLowerCase().includes(searchQuery.toLowerCase())
  );
  
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <Link
              to="/collections"
              className="text-sm text-indigo-600 hover:text-indigo-900 mb-4 inline-block"
            >
              ← Back to Collections
            </Link>
            <div className="flex justify-between items-center">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">{collection.name}</h1>
                {collection.description && (
                  <p className="mt-1 text-sm text-gray-500">{collection.description}</p>
                )}
              </div>
              <div className="flex items-center space-x-4">
                <span className="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
                  {collection.game}
                </span>
                <button
                  onClick={() => setShowAddCardForm(!showAddCardForm)}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  {showAddCardForm ? 'Cancel' : 'Add Card'}
                </button>
              </div>
            </div>
          </div>
          
          {showAddCardForm && (
            <div className="bg-white shadow sm:rounded-lg mb-6">
              <div className="px-4 py-5 sm:p-6">
                <Form method="post" className="space-y-4">
                  <input type="hidden" name="action" value="add-card" />
                  {actionData?.error && (
                    <div className="text-red-600">{actionData.error}</div>
                  )}
                  <div>
                    <label htmlFor="cardId" className="block text-sm font-medium text-gray-700">
                      Card ID
                    </label>
                    <input
                      type="text"
                      name="cardId"
                      id="cardId"
                      required
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    />
                  </div>
                  <div>
                    <label htmlFor="quantity" className="block text-sm font-medium text-gray-700">
                      Quantity
                    </label>
                    <input
                      type="number"
                      name="quantity"
                      id="quantity"
                      required
                      min="1"
                      defaultValue="1"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    />
                  </div>
                  <div>
                    <label htmlFor="condition" className="block text-sm font-medium text-gray-700">
                      Condition
                    </label>
                    <select
                      name="condition"
                      id="condition"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    >
                      <option value="">Select condition</option>
                      <option value="MINT">Mint</option>
                      <option value="NEAR_MINT">Near Mint</option>
                      <option value="LIGHTLY_PLAYED">Lightly Played</option>
                      <option value="MODERATELY_PLAYED">Moderately Played</option>
                      <option value="HEAVILY_PLAYED">Heavily Played</option>
                      <option value="DAMAGED">Damaged</option>
                    </select>
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      name="isFoil"
                      id="isFoil"
                      className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                    />
                    <label htmlFor="isFoil" className="ml-2 block text-sm text-gray-900">
                      Foil
                    </label>
                  </div>
                  <div>
                    <label htmlFor="notes" className="block text-sm font-medium text-gray-700">
                      Notes
                    </label>
                    <textarea
                      name="notes"
                      id="notes"
                      rows={3}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm bg-white text-gray-900"
                    />
                  </div>
                  <div>
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                    >
                      {isSubmitting ? 'Adding...' : 'Add Card'}
                    </button>
                  </div>
                </Form>
              </div>
            </div>
          )}

          <div className="mb-4">
            <input
              type="text"
              placeholder="Search cards..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            />
          </div>
          
          {filteredCards.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-lg shadow">
              <h3 className="mt-2 text-sm font-medium text-gray-900">
                {searchQuery ? 'No cards match your search' : 'No cards in collection'}
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                {searchQuery ? 'Try a different search term' : 'Add cards to your collection to get started.'}
              </p>
            </div>
          ) : (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Card
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Set
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Quantity
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Condition
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Notes
                    </th>
                    <th scope="col" className="relative px-6 py-3">
                      <span className="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredCards.map((card) => (
                    <tr key={card.id}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10">
                            {card.card.imageUrl && (
                              <img className="h-10 w-10 rounded" src={card.card.imageUrl} alt={card.card.name} />
                            )}
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium text-gray-900">{card.card.name}</div>
                            <div className="text-sm text-gray-500">{card.card.number}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">{card.card.setName}</div>
                        <div className="text-sm text-gray-500">{card.card.setCode}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {card.quantity}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                          {card.condition || 'Not specified'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {card.notes}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <Form method="post">
                          <input type="hidden" name="action" value="remove-card" />
                          <input type="hidden" name="cardId" value={card.id} />
                          <button
                            type="submit"
                            className="text-indigo-600 hover:text-indigo-900"
                          >
                            Remove
                          </button>
                        </Form>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
} 