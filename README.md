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

### Prerequisites
Install Devbox:
```bash
curl -fsSL https://get.jetpack.io/devbox | bash
```

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/shiftregister-vg/card-craft.git
   cd card-craft
   ```

2. Start the development environment:
   ```bash
   devbox shell
   ```

3. Start the development stack:
   ```bash
   devbox services up
   ```

The development stack includes:
- PostgreSQL database
- Backend server
- Frontend development server
- Database migrations
- Test data seeding

Once running, you can access:
- Frontend: http://localhost:5173
- GraphQL Playground: http://localhost:8080

## Project Structure

```
.
├── cmd/                # Application entry points
│   ├── seed/         # Database seeding tool
│   └── server/       # Main server application
├── docs/             # Documentation files
├── internal/         # Private application code
│   ├── auth/        # Authentication and authorization
│   ├── cards/       # Card management logic
│   ├── config/      # Application configuration
│   ├── database/    # Database connections and utilities
│   ├── graph/       # GraphQL schema and resolvers
│   ├── handlers/    # HTTP request handlers
│   ├── importers/   # Card import implementations
│   ├── middleware/  # HTTP/GraphQL middleware
│   ├── models/      # Database models
│   ├── seed/        # Seeding utilities
│   ├── server/      # Server setup and configuration
│   ├── services/    # Business logic services
│   ├── types/       # Common type definitions
│   └── utils/       # Shared utilities
├── migrations/      # Database migration files
├── pkg/            # Public packages
├── scripts/        # Utility scripts
├── tools/          # Development tools
└── web/           # Frontend application (Remix)
    ├── app/       # Application code
    │   ├── components/  # React components
    │   ├── context/    # React context providers
    │   ├── graphql/    # GraphQL queries and mutations
    │   ├── lib/        # Utility functions and configs
    │   ├── routes/     # Remix routes
    │   └── styles/     # CSS and styling
    └── public/    # Static assets
```

## Development Guide

### Managing the Development Stack

The project uses Devbox to manage all required services (PostgreSQL, etc.). Start the entire development stack with:

```bash
# From project root
devbox services up
```

This will start all necessary services in the foreground with logs visible. Use Ctrl+C to stop all services when done.

For development, it's recommended to keep this running in a dedicated terminal while you work in other terminals for running the application, tests, etc.

### Working with Migrations

1. Migration Files Location:
   ```
   migrations/
   ├── 000000_create_updated_at_function.up.sql   # Base function for updated_at
   ├── 000000_create_updated_at_function.down.sql
   ├── 000001_create_users.up.sql                 # Users table
   ├── 000001_create_users.down.sql
   └── ... other migrations
   ```

2. Creating New Migrations:
   ```bash
   # Create a new migration
   devbox run migrate:create "add_user_preferences"

   # This creates two files:
   # - migrations/YYYYMMDDHHMMSS_add_user_preferences.up.sql
   # - migrations/YYYYMMDDHHMMSS_add_user_preferences.down.sql
   ```

3. Migration Commands:
   ```bash
   # Apply all pending migrations
   devbox run migrate:up

   # Rollback last migration
   devbox run migrate:down
   ```

### Database Management

1. Connect to PostgreSQL (ensure devbox services are running):
   ```bash
   # Using psql (from devbox shell)
   psql -U postgres -d card_craft

   # Using connection string
   psql postgresql://postgres:postgres@localhost:5432/card_craft
   ```

2. Common Database Tasks:
   ```bash
   # Initialize a fresh database
   devbox run init-db

   # Seed test data
   devbox run seed
   ```

### Common Development Tasks

1. Generate GraphQL Code:
   ```bash
   # From project root
   go generate ./...
   ```

2. Update Frontend Dependencies:
   ```bash
   # From web directory
   pnpm update
   ```

3. Run Tests:
   ```bash
   # Backend tests
   go test ./...

   # Frontend tests
   cd web && pnpm test
   ```

4. Code Formatting:
   ```bash
   # Format Go code
   go fmt ./...

   # Format TypeScript/JavaScript
   cd web && pnpm format
   ```

5. Linting:
   ```bash
   # Lint Go code
   golangci-lint run

   # Lint TypeScript/JavaScript
   cd web && pnpm lint
   ```

### Troubleshooting

1. Reset Development Environment:
   ```bash
   # Stop the development stack (Ctrl+C if running in foreground)
   
   # Clean devbox shell
   exit  # if in devbox shell
   devbox clean

   # Start fresh
   devbox shell
   devbox services up

   # Reinitialize database if needed
   devbox run init-db
   devbox run migrate:up
   devbox run seed
   ```

2. Common Issues:
   - Port conflicts: Check if ports 5173, 5432, or 8080 are in use
   - Database connection: Ensure the development stack is running (`devbox services up`)
   - Migration errors: Check migration status and try rolling back with `devbox run migrate:down`
   - JWT issues: Clear browser cookies and try logging in again

## License

[License information to be added] 