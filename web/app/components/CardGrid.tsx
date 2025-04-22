import { Link } from '@remix-run/react';
import type { Card } from '../types/card.js';

interface CardGridProps {
  cards: Card[];
  collectionCardIds: Set<string>;
  onCardClick?: (card: Card) => void;
  lastCardRef?: (node: HTMLDivElement | null) => void;
  isLoading?: boolean;
}

export function CardGrid({ cards, collectionCardIds, onCardClick, lastCardRef, isLoading = false }: CardGridProps) {
  console.log('CardGrid received cards:', cards.length);

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 p-4 bg-white">
      {cards.map((card, index) => {
        const isLastCard = index === cards.length - 1;
        return (
          <div
            key={card.id}
            ref={isLastCard ? lastCardRef : undefined}
            className={`group relative aspect-[2.5/3.5] rounded-lg overflow-hidden shadow-lg transition-all duration-300 hover:shadow-xl ${
              collectionCardIds.has(card.id) ? 'ring-2 ring-blue-500' : ''
            }`}
            onClick={() => onCardClick?.(card)}
          >
            <img
              src={card.imageUrl}
              alt={card.name}
              className="w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
              <div className="absolute bottom-0 left-0 right-0 p-2 text-white">
                <p className="text-sm font-medium truncate">{card.name}</p>
                <p className="text-xs">{card.setCode} {card.number}</p>
              </div>
            </div>
            {collectionCardIds.has(card.id) && (
              <div className="absolute top-2 right-2 bg-blue-500 text-white rounded-full w-6 h-6 flex items-center justify-center">
                ✓
              </div>
            )}
          </div>
        );
      })}
      {isLoading && (
        <div className="flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent"></div>
        </div>
      )}
    </div>
  );
} 