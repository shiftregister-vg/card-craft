import { Link } from '@remix-run/react';
import type { Card } from '../types/card.js';

interface CardGridProps {
  cards: Card[];
  collectionCardIds: Set<string>;
  onCardClick?: (card: Card) => void;
}

export function CardGrid({ cards, collectionCardIds, onCardClick }: CardGridProps) {
  console.log('CardGrid received cards:', cards.length);
  console.log('CardGrid received collectionCardIds:', Array.from(collectionCardIds));

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 p-4 bg-white">
      {cards.map((card) => {
        console.log('Rendering card:', card);
        return (
          <div
            key={card.id}
            className="group relative aspect-[2.5/3.5] rounded-lg overflow-hidden shadow-lg transition-all duration-300 hover:shadow-xl"
            onClick={() => onCardClick?.(card)}
          >
            <img
              src={card.imageUrl}
              alt={card.name}
              className={`w-full h-full object-cover transition-opacity duration-300 ${
                collectionCardIds.has(card.id) ? 'opacity-100' : 'opacity-50 group-hover:opacity-100'
              }`}
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
              <div className="absolute bottom-0 left-0 right-0 p-2 text-white">
                <p className="text-sm font-medium truncate">{card.name}</p>
                <p className="text-xs">{card.setCode} {card.number}</p>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
} 