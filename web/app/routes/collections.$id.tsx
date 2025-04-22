import { json, LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, useActionData, useLoaderData, useNavigation, useParams } from '@remix-run/react';
import { useState, useEffect, useRef, useCallback } from 'react';
import { createServerClient } from '../lib/urql.js';
import { requireUser } from '../utils/auth.server.js';
import { COLLECTION_QUERY, ADD_CARD_TO_COLLECTION_MUTATION, REMOVE_CARD_FROM_COLLECTION_MUTATION, CARDS_BY_GAME_QUERY } from '../graphql/collections.js';
import { CardGrid } from '../components/CardGrid.js';
import { CollectionCardGrid } from '../components/CollectionCardGrid.js';
import type { Card } from '../types/card.js';
import type { Collection } from '../types/collection.js';

interface CollectionCard {
  id: string;
  card: Card;
  quantity: number;
  condition: string | null;
  isFoil: boolean | null;
  notes: string | null;
}

interface LoaderData {
  collection: Collection;
  initialCards: Card[];
  hasNextPage: boolean;
  endCursor: string | null;
}

interface ActionData {
  error?: string;
  success?: boolean;
}

export async function loader({ request, params }: LoaderFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  
  if (!params.id) {
    throw new Response('Collection ID is required', { status: 400 });
  }

  console.log('Fetching collection with ID:', params.id);

  // Fetch collection details
  const { data, error } = await serverClient.query(COLLECTION_QUERY, {
    id: params.id,
  }).toPromise();
  
  if (error) {
    console.error('GraphQL error:', error);
    throw new Response('Error fetching collection', { status: 500 });
  }
  
  console.log('Collection data:', JSON.stringify(data, null, 2));
  
  if (!data?.collection) {
    throw new Response('Collection not found', { status: 404 });
  }

  // Fetch initial set of cards for the collection's game
  const { data: cardsData, error: cardsError } = await serverClient.query(CARDS_BY_GAME_QUERY, {
    game: data.collection.game.toLowerCase(),
    first: 200,
  }).toPromise();

  if (cardsError) {
    console.error('GraphQL error:', cardsError);
    throw new Response('Error fetching cards', { status: 500 });
  }

  const initialCards = cardsData?.cardsByGame?.edges?.map((edge: any) => edge.node) || [];
  const hasNextPage = cardsData?.cardsByGame?.pageInfo?.hasNextPage || false;
  const endCursor = cardsData?.cardsByGame?.pageInfo?.endCursor || null;

  return json<LoaderData>({ 
    collection: data.collection,
    initialCards,
    hasNextPage,
    endCursor,
  });
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
  const { collection, initialCards, hasNextPage, endCursor } = useLoaderData<LoaderData>();
  const actionData = useActionData<ActionData>();
  const navigation = useNavigation();
  const params = useParams();
  const [showAddCard, setShowAddCard] = useState(false);
  const [activeTab, setActiveTab] = useState("cards");
  const [cards, setCards] = useState<Card[]>(initialCards);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(hasNextPage);
  const [cursor, setCursor] = useState<string | null>(endCursor);
  const observer = useRef<IntersectionObserver | null>(null);
  const lastCardElementRef = useCallback((node: HTMLDivElement | null) => {
    if (loading) return;
    if (observer.current) observer.current.disconnect();
    observer.current = new IntersectionObserver(entries => {
      if (entries[0].isIntersecting && hasMore) {
        loadMoreCards();
      }
    });
    if (node) observer.current.observe(node);
  }, [loading, hasMore]);

  // Auto-load next page when current page is loaded
  useEffect(() => {
    if (!loading && hasMore && cards.length > 0) {
      const scrollPosition = window.innerHeight + window.scrollY;
      const documentHeight = document.documentElement.scrollHeight;
      if (documentHeight - scrollPosition < window.innerHeight * 2) {
        loadMoreCards();
      }
    }
  }, [loading, hasMore, cards.length]);

  const loadMoreCards = async () => {
    if (!cursor || loading) return;
    
    setLoading(true);
    const serverClient = createServerClient(new Request(window.location.href));
    
    try {
      const { data, error } = await serverClient.query(CARDS_BY_GAME_QUERY, {
        game: collection.game.toLowerCase(),
        first: 200,
        after: cursor,
      }).toPromise();

      if (error) {
        console.error('Error loading more cards:', error);
        return;
      }

      const newCards = data?.cardsByGame?.edges?.map((edge: any) => edge.node) || [];
      setCards(prevCards => [...prevCards, ...newCards]);
      setHasMore(data?.cardsByGame?.pageInfo?.hasNextPage || false);
      setCursor(data?.cardsByGame?.pageInfo?.endCursor || null);
    } catch (error) {
      console.error('Error loading more cards:', error);
    } finally {
      setLoading(false);
    }
  };

  // Create a set of card IDs that are in the collection for quick lookup
  const collectionCardIds: Set<string> = new Set(collection.cards.map(card => card.card.id));

  const handleCardClick = (card: Card) => {
    console.log('Card clicked:', card);
    if (collectionCardIds.has(card.id)) {
      // Card is in collection, show details or remove
      console.log('Card is in collection');
      setShowAddCard(true);
    } else {
      // Card is not in collection, show add form
      console.log('Card is not in collection');
      setShowAddCard(true);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold">{collection.name}</h1>
            {collection.description && (
              <p className="text-gray-600 mt-2">{collection.description}</p>
            )}
          </div>
          <button
            onClick={() => setShowAddCard(true)}
            className="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 transition-colors"
          >
            Add Card
          </button>
        </div>

        <div className="flex space-x-4 mb-4">
          <button
            onClick={() => setActiveTab("cards")}
            className={`px-4 py-2 rounded-lg ${
              activeTab === "cards"
                ? "bg-blue-500 text-white"
                : "bg-gray-200 text-gray-700 hover:bg-gray-300"
            }`}
          >
            All Cards ({cards.length})
          </button>
          <button
            onClick={() => setActiveTab("collection")}
            className={`px-4 py-2 rounded-lg ${
              activeTab === "collection"
                ? "bg-blue-500 text-white"
                : "bg-gray-200 text-gray-700 hover:bg-gray-300"
            }`}
          >
            Collection Cards ({collection.cards.length})
          </button>
        </div>

        <div className="mt-8">
          {activeTab === "cards" ? (
            <>
              <CardGrid
                cards={cards}
                collectionCardIds={collectionCardIds}
                onCardClick={handleCardClick}
                lastCardRef={lastCardElementRef}
              />
              {loading && (
                <div className="text-center py-4">
                  <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent"></div>
                </div>
              )}
            </>
          ) : (
            <CollectionCardGrid
              cards={collection.cards}
              onCardClick={(collectionCard) => handleCardClick(collectionCard.card)}
            />
          )}
        </div>

        {showAddCard && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="bg-white p-6 rounded-lg shadow-xl max-w-lg w-full">
              <h2 className="text-2xl font-bold mb-4">Add Card to Collection</h2>
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
                    <option value="mint">Mint</option>
                    <option value="near_mint">Near Mint</option>
                    <option value="excellent">Excellent</option>
                    <option value="good">Good</option>
                    <option value="light_played">Light Played</option>
                    <option value="played">Played</option>
                    <option value="poor">Poor</option>
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
                <div className="flex justify-end space-x-3">
                  <button
                    type="button"
                    onClick={() => setShowAddCard(false)}
                    className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                  >
                    Add Card
                  </button>
                </div>
              </Form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
} 