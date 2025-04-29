import { LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, useActionData, useLoaderData, useNavigation, useParams, useFetcher, Link, Outlet, useLocation, useSearchParams } from '@remix-run/react';
import { useState, useEffect, useRef } from 'react';
import { createServerClient } from '../lib/urql.js';
import type { Card } from '../types/card.js';
import { COLLECTION_QUERY, CARDS_BY_GAME_QUERY, ADD_CARD_TO_COLLECTION_MUTATION, REMOVE_CARD_FROM_COLLECTION_MUTATION, SEARCH_CARDS_QUERY } from '../graphql/collections.js';
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

interface ActionData {
  error?: string;
  success?: boolean;
}

type FetcherData = {
  initialCards: Card[];
  hasNextPage: boolean;
  endCursor: string | null;
  searchQuery?: string;
};

interface CardEdge {
  node: Card;
}

export const loader = async ({ request, params }: LoaderFunctionArgs) => {
  try {
    await requireUser(request);
    const url = new URL(request.url);
    const cursor = url.searchParams.get("cursor") || null;
    const searchQuery = url.searchParams.get("search") || "";
    const pageSize = 50;

    console.log('Loader called with search query:', searchQuery);

    const serverClient = createServerClient(request);
    
    if (!params.id) {
      throw new Response('Collection ID is required', { status: 400 });
    }

    // First fetch collection details to get the game
    const { data: collectionData, error: collectionError } = await serverClient.query(COLLECTION_QUERY, {
      id: params.id,
    }).toPromise();

    if (collectionError) {
      console.error('Collection error:', collectionError);
      throw new Response('Error fetching collection', { status: 500 });
    }

    if (!collectionData?.collection) {
      throw new Response('Collection not found', { status: 404 });
    }

    const game = collectionData.collection.game.toLowerCase();
    console.log('Fetching cards for game:', game, 'with search query:', searchQuery);

    // Then fetch cards for the collection's game
    const { data: cardsData, error: cardsError } = await serverClient.query(
      searchQuery ? SEARCH_CARDS_QUERY : CARDS_BY_GAME_QUERY,
      searchQuery
        ? {
            game,
            name: searchQuery,
            page: 1,
            pageSize: 20,
          }
        : {
            game,
            first: pageSize,
            after: cursor,
          }
    ).toPromise();

    if (cardsError) {
      console.error('Cards error:', cardsError);
      throw new Response('Error fetching cards', { status: 500 });
    }

    console.log('Received cards data:', cardsData);

    const cards = searchQuery
      ? cardsData.searchCards.cards
      : cardsData.cardsByGame.edges.map((edge: CardEdge) => edge.node);

    console.log('Processed cards:', cards.length);

    return {
      collection: collectionData.collection,
      initialCards: cards,
      hasNextPage: searchQuery ? false : cardsData.cardsByGame.pageInfo.hasNextPage,
      endCursor: searchQuery ? null : cardsData.cardsByGame.pageInfo.endCursor,
      searchQuery,
    };
  } catch (error) {
    console.error('Loader error:', error);
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Internal Server Error', { status: 500 });
  }
};

export async function action({ request, params }: ActionFunctionArgs) {
  try {
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
        console.error('Add card error:', error);
        return { error: error.message };
      }

      return { success: true };
    }

    if (action === 'remove-card') {
      const cardId = formData.get('cardId') as string;

      const { error } = await serverClient.mutation(REMOVE_CARD_FROM_COLLECTION_MUTATION, {
        id: cardId,
      }).toPromise();

      if (error) {
        console.error('Remove card error:', error);
        return { error: error.message };
      }

      return { success: true };
    }

    return { error: 'Invalid action' };
  } catch (error) {
    console.error('Action error:', error);
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Internal Server Error', { status: 500 });
  }
}

export default function CollectionDetail() {
  const { collection, initialCards, hasNextPage, endCursor, searchQuery: initialSearchQuery } = useLoaderData<typeof loader>();
  const actionData = useActionData<ActionData>();
  const params = useParams();
  const location = useLocation();
  const fetcher = useFetcher<FetcherData>();
  const [activeTab, setActiveTab] = useState("cards");
  const [cards, setCards] = useState<Card[]>(initialCards);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(hasNextPage);
  const [cursor, setCursor] = useState<string | null>(endCursor);
  const [showBackToTop, setShowBackToTop] = useState(false);
  const [searchQuery, setSearchQuery] = useState(initialSearchQuery);
  const [previousSearchQuery, setPreviousSearchQuery] = useState(initialSearchQuery);
  const [selectedCard, setSelectedCard] = useState<Card | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const loadingRef = useRef(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const searchTimeoutRef = useRef<NodeJS.Timeout>();
  const isSearchingRef = useRef(false);

  // Handle scroll events for back to top button and infinite loading
  useEffect(() => {
    const handleScroll = () => {
      // Handle back to top button visibility
      if (headerRef.current) {
        const headerRect = headerRef.current.getBoundingClientRect();
        setShowBackToTop(headerRect.bottom < 0);
      }

      // Handle infinite loading
      if (!containerRef.current || loadingRef.current || !hasMore || !cursor || isSearchingRef.current) {
        return;
      }

      const { scrollTop, scrollHeight, clientHeight } = document.documentElement;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 200;

      if (isNearBottom) {
        loadingRef.current = true;
        setLoading(true);
        fetcher.load(`/collections/${params.id}?cursor=${encodeURIComponent(cursor)}`);
      }
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [hasMore, cursor, params.id, fetcher]);

  // Handle search input changes
  useEffect(() => {
    // Skip initial render
    if (searchQuery === initialSearchQuery) {
      return;
    }

    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    if (isSearchingRef.current) {
      return;
    }

    if (searchQuery.trim() === "") {
      // Cancel any pending fetcher requests
      fetcher.data = undefined;
      isSearchingRef.current = false;
      
      console.log('Clearing search, initialCards:', initialCards);
      setCards(initialCards);
      setHasMore(hasNextPage);
      setCursor(endCursor);
      setPreviousSearchQuery("");
      setLoading(false);
      console.log('State after clearing search:', {
        cards: initialCards,
        hasMore: hasNextPage,
        cursor: endCursor
      });
      return;
    }

    if (searchQuery === previousSearchQuery) {
      return;
    }

    setPreviousSearchQuery(searchQuery);
    setLoading(true);
    setCards([]);
    setHasMore(false);
    setCursor(null);

    searchTimeoutRef.current = setTimeout(() => {
      isSearchingRef.current = true;
      fetcher.load(`/collections/${params.id}?search=${encodeURIComponent(searchQuery)}`);
    }, 300);

    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [searchQuery, params.id, fetcher, initialSearchQuery]);

  // Handle clearing the search
  const handleClearSearch = () => {
    setSearchQuery("");
    setCards(initialCards);
    setHasMore(hasNextPage);
    setCursor(endCursor);
    isSearchingRef.current = false;
  };

  // Update cards when fetcher data changes
  useEffect(() => {
    if (fetcher.data) {
      // Skip if we're not searching and the data contains a search query
      if (!isSearchingRef.current && fetcher.data.searchQuery) {
        return;
      }

      console.log('Fetcher data received:', fetcher.data);
      if (isSearchingRef.current) {
        console.log('Updating cards from search results');
        setCards(fetcher.data.initialCards);
        setHasMore(false);
        setCursor(null);
        isSearchingRef.current = false;
      } else if (cursor && fetcher.data) {
        console.log('Appending more cards');
        const data = fetcher.data;
        setCards(prevCards => [...prevCards, ...data.initialCards]);
        setHasMore(data.hasNextPage);
        setCursor(data.endCursor);
      } else {
        console.log('Setting initial cards from fetcher');
        setCards(fetcher.data.initialCards);
        setHasMore(fetcher.data.hasNextPage);
        setCursor(fetcher.data.endCursor);
      }
      
      setLoading(false);
      loadingRef.current = false;
    }
  }, [fetcher.data, cursor]);

  // Reset cards when collection changes
  useEffect(() => {
    setCards(initialCards);
    setHasMore(hasNextPage);
    setCursor(endCursor);
    loadingRef.current = false;
    setLoading(false);
    isSearchingRef.current = false;
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

            {activeTab === "cards" && (
              <div className="mb-6">
                <div className="relative">
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => {
                      const newValue = e.target.value;
                      if (newValue.trim() === "") {
                        handleClearSearch();
                      } else {
                        setSearchQuery(newValue);
                      }
                    }}
                    placeholder="Search cards..."
                    className="w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                  />
                  {searchQuery && (
                    <button
                      onClick={handleClearSearch}
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                    >
                      <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  )}
                </div>
              </div>
            )}

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
                onClick={() => {
                  window.scrollTo({
                    top: 0,
                    behavior: 'smooth'
                  });
                }}
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

            {showAddModal && selectedCard && (
              <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                <div className="bg-white p-6 rounded-lg shadow-xl max-w-lg w-full">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h2 className="text-2xl font-bold">Add Card to Collection</h2>
                      <p className="text-gray-600">{selectedCard.name}</p>
                    </div>
                    <button
                      onClick={() => {
                        setShowAddModal(false);
                        setSelectedCard(null);
                      }}
                      className="text-gray-400 hover:text-gray-600"
                    >
                      <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                  <Form method="post" className="space-y-4">
                    <input type="hidden" name="action" value="add-card" />
                    <input type="hidden" name="cardId" value={selectedCard.id} />
                    {actionData?.error && (
                      <div className="text-red-600">{actionData.error}</div>
                    )}
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
                        onClick={() => {
                          setShowAddModal(false);
                          setSelectedCard(null);
                        }}
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