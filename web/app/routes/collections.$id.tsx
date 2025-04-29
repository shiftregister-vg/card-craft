import { json, LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, useActionData, useLoaderData, useNavigation, useParams, useFetcher, Link, Outlet, useLocation, useSearchParams } from '@remix-run/react';
import { useState, useEffect, useRef, useCallback } from 'react';
import { createServerClient } from '../lib/urql.js';
import type { Card } from '../types/card.js';
import type { Collection } from '../types/collection.js';
import { COLLECTION_QUERY, CARDS_BY_GAME_QUERY, ADD_CARD_TO_COLLECTION_MUTATION, REMOVE_CARD_FROM_COLLECTION_MUTATION } from '../graphql/collections.js';
import { requireUser } from '../utils/auth.server.js';
import { CardGrid } from '../components/CardGrid.js';
import { CollectionCardGrid } from '../components/CollectionCardGrid.js';

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
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

interface ActionData {
  error?: string;
  success?: boolean;
}

type FetcherData = {
  initialCards: Card[];
  hasNextPage: boolean;
  endCursor: string | null;
};

interface CardEdge {
  node: Card;
}

export const loader = async ({ request, params }: LoaderFunctionArgs) => {
  await requireUser(request);
  const url = new URL(request.url);
  const cursor = url.searchParams.get("cursor") || null;
  const pageSize = 50;

  const serverClient = createServerClient(request);
  
  if (!params.id) {
    throw new Response('Collection ID is required', { status: 400 });
  }

  // Fetch collection details
  const { data: collectionData, error: collectionError } = await serverClient.query(COLLECTION_QUERY, {
    id: params.id,
  }).toPromise();

  if (collectionError) {
    throw new Response('Error fetching collection', { status: 500 });
  }

  if (!collectionData?.collection) {
    throw new Response('Collection not found', { status: 404 });
  }

  // Fetch cards for the collection's game
  const { data: cardsData, error: cardsError } = await serverClient.query(CARDS_BY_GAME_QUERY, {
    game: collectionData.collection.game.toLowerCase(),
    first: pageSize,
    after: cursor,
  }).toPromise();

  if (cardsError) {
    throw new Response('Error fetching cards', { status: 500 });
  }

  return json({
    collection: collectionData.collection,
    initialCards: cardsData.cardsByGame.edges.map((edge: CardEdge) => edge.node),
    hasNextPage: cardsData.cardsByGame.pageInfo.hasNextPage,
    endCursor: cardsData.cardsByGame.pageInfo.endCursor,
  });
};

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
  const { collection, initialCards, hasNextPage, endCursor } = useLoaderData<typeof loader>();
  const actionData = useActionData<ActionData>();
  const navigation = useNavigation();
  const params = useParams();
  const location = useLocation();
  const fetcher = useFetcher<FetcherData>();
  const [showAddCard, setShowAddCard] = useState(false);
  const [activeTab, setActiveTab] = useState("cards");
  const [cards, setCards] = useState<Card[]>(initialCards);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(hasNextPage);
  const [cursor, setCursor] = useState<string | null>(endCursor);
  const [showBackToTop, setShowBackToTop] = useState(false);
  const loadingRef = useRef(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const lastProcessedCursorRef = useRef<string | null>(null);

  // Handle scroll events for back to top button
  useEffect(() => {
    const handleScroll = () => {
      if (!headerRef.current) return;
      
      const headerRect = headerRef.current.getBoundingClientRect();
      setShowBackToTop(headerRect.bottom < 0);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  };

  const loadMoreCards = useCallback(() => {
    if (!cursor || loadingRef.current) {
      console.log('LoadMoreCards blocked:', { cursor, loadingRef: loadingRef.current });
      return;
    }
    
    console.log('Loading more cards with cursor:', cursor);
    loadingRef.current = true;
    setLoading(true);
    lastProcessedCursorRef.current = cursor;
    fetcher.load(`/collections/${params.id}?cursor=${encodeURIComponent(cursor)}`);
  }, [cursor, params.id, fetcher]);

  // Handle scroll events for infinite loading
  useEffect(() => {
    const handleScroll = () => {
      if (!containerRef.current || loadingRef.current || !hasMore || !cursor) {
        return;
      }

      const { scrollTop, scrollHeight, clientHeight } = document.documentElement;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 200;

      if (isNearBottom) {
        loadMoreCards();
      }
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [loadMoreCards, hasMore, cursor]);

  // Update cards when fetcher data changes
  useEffect(() => {
    if (fetcher.data) {
      const newCards = fetcher.data.initialCards || [];
      if (newCards.length > 0) {
        // Create a Set of existing card IDs for quick lookup
        const existingCardIds = new Set(cards.map(card => card.id));
        // Filter out any cards that already exist in our list
        const uniqueNewCards = newCards.filter(card => !existingCardIds.has(card.id));
        
        if (uniqueNewCards.length > 0) {
          setCards(prevCards => [...prevCards, ...uniqueNewCards]);
        }
      }
      
      setHasMore(fetcher.data.hasNextPage);
      setCursor(fetcher.data.endCursor);
      setLoading(false);
      loadingRef.current = false;
    }
  }, [fetcher.data, cards]);

  // Handle fetcher state changes
  useEffect(() => {
    if (fetcher.state === 'idle' && loadingRef.current) {
      setLoading(false);
      loadingRef.current = false;
    } else if (fetcher.state === 'submitting') {
      loadingRef.current = true;
      setLoading(true);
    }
  }, [fetcher.state]);

  // Reset cards when collection changes
  useEffect(() => {
    setCards(initialCards);
    setHasMore(hasNextPage);
    setCursor(endCursor);
    loadingRef.current = false;
    setLoading(false);
  }, [initialCards, hasNextPage, endCursor]);

  // Create a set of card IDs that are in the collection for quick lookup
  const collectionCardIds: Set<string> = new Set(collection.cards.map((card: CollectionCard) => card.card.id));

  const handleAddCard = (cardId: string) => {
    fetcher.submit(
      { cardId, action: 'add' },
      { method: 'post' }
    );
  };

  const handleRemoveCard = (cardId: string) => {
    fetcher.submit(
      { cardId, action: 'remove' },
      { method: 'post' }
    );
  };

  const handleCardClick = (card: Card) => {
    if (isCardInCollection(card, collectionCardIds)) {
      handleRemoveCard(card.id);
    } else {
      handleAddCard(card.id);
    }
  };

  // Check if we're on the import route
  const isImportRoute = location.pathname.endsWith('/import');

  // Add type for card parameter
  const isCardInCollection = (card: Card, collectionCardIds: Set<string>) => {
    return collectionCardIds.has(card.id);
  };

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-gray-900">{collection.name}</h1>
          {!isImportRoute && (
            <div className="flex space-x-4">
              {collection.game.toLowerCase() === "pokemon" && (
                <Link
                  to={`/collections/${collection.id}/import`}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Import from TCG Collector
                </Link>
              )}
              <button
                onClick={() => setShowAddCard(true)}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
              >
                Add Card
              </button>
            </div>
          )}
        </div>

        {isImportRoute ? (
          <Outlet />
        ) : (
          <>
            <div className="flex space-x-4 mb-4">
              <button
                onClick={() => setActiveTab("cards")}
                className={`px-4 py-2 rounded-lg ${
                  activeTab === "cards"
                    ? "bg-blue-500 text-white"
                    : "bg-gray-200 text-gray-700 hover:bg-gray-300"
                }`}
              >
                All Cards
              </button>
              <button
                onClick={() => setActiveTab("collection")}
                className={`px-4 py-2 rounded-lg ${
                  activeTab === "collection"
                    ? "bg-blue-500 text-white"
                    : "bg-gray-200 text-gray-700 hover:bg-gray-300"
                }`}
              >
                Collection Cards
              </button>
            </div>

            <div className="mt-8" ref={containerRef}>
              {activeTab === "cards" ? (
                <CardGrid
                  cards={cards}
                  collectionCardIds={collectionCardIds}
                  onCardClick={handleCardClick}
                  isLoading={loading}
                />
              ) : (
                <CollectionCardGrid
                  cards={collection.cards}
                  onCardClick={(collectionCard) => handleCardClick(collectionCard.card)}
                />
              )}
            </div>

            {showBackToTop && (
              <button
                onClick={scrollToTop}
                className="fixed bottom-8 right-8 bg-blue-500 text-white p-3 rounded-full shadow-lg hover:bg-blue-600 transition-colors z-50"
                aria-label="Back to top"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 10l7-7m0 0l7 7m-7-7v18"
                  />
                </svg>
              </button>
            )}

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
          </>
        )}
      </div>
    </div>
  );
} 