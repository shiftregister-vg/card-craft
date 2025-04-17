# Card Craft

A web-based application for managing and building custom decks for various Trading Card Games (TCGs) including Pokémon, Star Wars: Unlimited, and Disney Lorcana.

## Features

- User authentication and account management
- Deck building and management
- Card collection import from various TCG services:
  - Pokémon TCG (tcgcollector.com)
  - Disney Lorcana (dreamborn.ink)
  - Star Wars: Unlimited (starwarsunlimited.com)
- Responsive web interface
- PostgreSQL database for data storage

## Tech Stack

- Backend: Go 1.24.1
- Database: PostgreSQL
- Package Management: Devbox
- Version Control: Git

## Getting Started

1. Install Devbox:
   ```bash
   curl -fsSL https://get.jetpack.io/devbox | bash
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/shiftregister-vg/card-craft.git
   cd card-craft
   ```

3. Start the development environment:
   ```bash
   devbox shell
   ```

4. Install dependencies:
   ```bash
   go mod tidy
   ```

5. Set up the database:
   ```bash
   # Start PostgreSQL
   devbox services start postgresql
   
   # Create database and run migrations
   # (Instructions will be added as the project develops)
   ```

6. Start the development server:
   ```bash
   go run cmd/server/main.go
   ```

## Project Structure

```
.
├── cmd/                # Application entry points
├── internal/          # Private application code
│   ├── auth/         # Authentication logic
│   ├── database/     # Database models and migrations
│   ├── handlers/     # HTTP handlers
│   ├── importers/    # Card import logic
│   └── services/     # Business logic
├── pkg/              # Public packages
├── web/              # Frontend code
└── migrations/       # Database migrations
```

## License

[License information to be added] 