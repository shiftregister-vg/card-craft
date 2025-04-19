import type { Card } from './card.js';

export interface CollectionCard {
  id: string;
  collectionId: string;
  cardId: string;
  card: Card;
  quantity: number;
  condition: string | null;
  isFoil: boolean | null;
  notes: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Collection {
  id: string;
  userId: string;
  name: string;
  description: string | null;
  game: string;
  cards: CollectionCard[];
  createdAt: string;
  updatedAt: string;
} 