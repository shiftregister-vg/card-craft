import { Link } from '@remix-run/react';
import type { Card } from '../types/card.js';
import type { CollectionCard } from '../types/collection.js';

interface CollectionCardGridProps {
  cards: CollectionCard[];
  onCardClick?: (collectionCard: CollectionCard) => void;
}

export function CollectionCardGrid({ cards, onCardClick }: CollectionCardGridProps) {
  console.log('CollectionCardGrid received cards:', cards.length);

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 p-4 bg-white">
      {cards.map((collectionCard) => {
        const card = collectionCard.card;
        console.log('Rendering collection card:', card);
        return (
          <div
            key={collectionCard.id}
            className="group relative aspect-[2.5/3.5] rounded-lg overflow-hidden shadow-lg transition-all duration-300 hover:shadow-xl"
            onClick={() => onCardClick?.(collectionCard)}
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
                <p className="text-xs">Quantity: {collectionCard.quantity}</p>
                {collectionCard.condition && (
                  <p className="text-xs">Condition: {collectionCard.condition}</p>
                )}
                {collectionCard.isFoil && (
                  <p className="text-xs">Foil</p>
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
} 