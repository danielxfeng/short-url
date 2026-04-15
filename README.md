# short-url

A full-stack URL shortener with OAuth login, JWT-protected link management, click tracking, custom aliases, and notes.

The project is organized as a monorepo:

- `apps/backend-chi`: Go + Chi backend API
- `apps/backend-ktor`: Kotlin + Ktor backend API [experimental](#experimental)
- `apps/frontend`: React + Vite frontend

The project is hosted on: [https://s.danielslab.dev](https://s.danielslab.dev)

## Features

- Create short URLs for any valid target URL
- Optional custom short code
- Click statistics for each short link
- Optional note field for internal context
- Soft delete, restore, and permanent delete
- OAuth login with Google and GitHub
- JWT-based authenticated API access
- Frontend redirect handling for unknown or invalid short codes
- Sentry hooks on both frontend and backend

## How It Works

1. A signed-in user creates a short link from the frontend.
2. The backend stores the original URL, generated or custom code, click count, and optional note.
3. Visiting `/{code}` redirects to the target URL.
4. The backend increments the click count for successful redirects.
5. Authenticated users can list, delete, restore, and permanently remove their own links.

## Auth Model

- OAuth providers: Google and GitHub
- Backend handles OAuth exchange and user upsert
- Backend returns a JWT after successful auth callback
- Frontend stores the authenticated user and token in local state persisted with Zustand
- Protected API routes require `Authorization: Bearer <token>`

## Stack

### Backend

- Go
- Chi router
- PostgreSQL
- sqlc
- golang-migrate
- go-playground/validator
- OAuth2
- JWT
- Sentry Go SDK

### Frontend

- React 19
- Vite
- TypeScript
- React Router
- TanStack Query
- TanStack Form
- Zustand
- Tailwind CSS v4
- Shadcn/ui
- Vitest + Testing Library
- Sentry React SDK

## Project Structure

```text
.
├── apps
│   ├── backend-chi
│   │   ├── package.json                 # Backend scripts
│   │   ├── cmd
│   │   │   ├── server/main.go           # Backend application entrypoint and HTTP server bootstrap
│   │   │   └── migrate/main.go          # Migration runner
│   │   └── internal
│   │       ├── dep
│   │       │   ├── config.go            # Backend environment config loading
│   │       │   ├── dep.go               # Dependency initialization, validator setup, DB pool setup
│   │       │   ├── logger.go            # Structured logger configuration
│   │       │   └── sentryinit.go        # Backend Sentry initialization
│   │       └── api
│   │           ├── auth
│   │           │   ├── oauth.go         # OAuth provider config and user info fetch logic
│   │           │   └── token.go         # JWT generation and validation
│   │           ├── dto/dto.go           # Request/response DTOs and validation rules
│   │           ├── mymiddleware
│   │           │   ├── auth.go          # JWT auth middleware
│   │           │   └── helmet.go        # Security headers middleware
│   │           ├── router
│   │           │   ├── router.go        # Base router, middleware, and health endpoint
│   │           │   ├── shorturl.go      # Short URL redirect and authenticated link management routes
│   │           │   └── user.go          # OAuth routes and authenticated user endpoints
│   │           ├── util/util.go         # Helper functions
│   │           └── repository
│   │               ├── inmemory/state.go # In-memory OAuth state/verifier store
│   │               └── db
│   │                   ├── query/query.sql # SQL source used by sqlc
│   │                   ├── query.sql.go  # Generated sqlc queries
│   │                   └── query/migrations # Database migrations
│   └── frontend
│       ├── package.json                 # Frontend scripts
│       ├── vercel.json                  # Vercel rewrites for SPA routes and short-link handling
│       └── src
│           ├── main.tsx                 # Frontend bootstrap, QueryClient setup, and Sentry error boundary
│           ├── App.tsx                  # Top-level route definitions
│           ├── instrument.ts            # Frontend Sentry initialization
│           ├── config/config.ts         # Frontend runtime config such as API base URL
│           ├── components
│           │   ├── AuthCallback.tsx     # OAuth callback handling on the frontend
│           │   ├── NotFound.tsx         # Invalid route handling and short-code redirect fallback
│           │   └── homepage
│           │       ├── AuthGuard.tsx    # Login prompt and provider links for unauthenticated users
│           │       ├── CreateLinkForm.tsx # Create short link UI with optional custom code and note
│           │       └── LinksTable.tsx   # Link list, clicks, and delete/restore/permanent delete actions
│           ├── hooks
│           │   ├── useAddLinkForm.ts    # Form logic for creating links
│           │   ├── useLinks.ts          # Paginated link query logic
│           │   ├── useMutateLink.ts     # Create/delete/restore/permanent-delete mutations
│           │   └── useUser.ts           # Persisted auth state with Zustand
│           └── services
│               ├── service.ts           # Shared fetch wrapper for API calls
│               ├── linkService.ts       # Link-specific API requests
│               └── userService.ts       # User/auth-related API requests
```

## Prerequisites

- Go 1.25+
- PostgreSQL
- pnpm 10+
- Node.js 20+

## Installation

Install workspace dependencies:

```bash
pnpm install
```

## Environment Variables

Copy the sample env files and adjust the values for your local setup:

```bash
cp apps/backend-chi/.env.sample apps/backend-chi/.env
cp apps/frontend/.env.sample apps/frontend/.env
```

Notes:

- `JWT_EXPIRY` is in hours.
- `BACKEND_PUBLIC_URL` must match the public backend base URL used by OAuth callbacks.
- `FRONTEND_REDIRECT_URL` must match the frontend auth callback URL.
- `NOT_FOUND_PAGE` is where invalid short codes are redirected.

## Database Setup

Run backend migrations:

```bash
pnpm --filter backend-chi db:migrate
```

If you change SQL in `query.sql`, regenerate sqlc output:

```bash
pnpm --filter backend-chi db:generate
```

## Running Locally

### Backend

```bash
pnpm --filter backend-chi dev
```

The API starts on `http://localhost:8080` by default.

### Frontend

```bash
pnpm --filter frontend dev
```

The frontend starts on `http://localhost:5173` by default.

## Testing

Run all tests:

```bash
pnpm -r run test
```

## API Overview

### Public

- `GET /api/v1/health`
- `GET /api/v1/short-urls/{code}` redirect to original URL

### Auth

- `GET /api/v1/user/auth/{provider}`
- `GET /api/v1/user/auth/{provider}/callback`

### Protected

- `GET /api/v1/user/me`
- `DELETE /api/v1/user/me`
- `GET /api/v1/short-urls`
- `POST /api/v1/short-urls`
- `PUT /api/v1/short-urls/{code}/restore`
- `DELETE /api/v1/short-urls/{code}`
- `DELETE /api/v1/short-urls/{code}/permanent`

## Development Notes

- Redirect lookups are handled by the backend and the frontend fallback route.
- Click counts are updated after successful redirect.
- Link records support soft deletion via `deleted_at`.
- Notes are optional and stored with each link.
- Pagination on managed links is cursor-based and backed by database indexes.

## Observability

- Frontend uses Sentry React integration and an error boundary.
- Backend initializes Sentry and integrates it with structured logging.

## Experimental

### Ktor Backend

The Ktor backend in `apps/backend-ktor` is the Kotlin implementation of the same short-url API. Its goal is to replicate the Go server's behavior while running as a separate app on `http://localhost:8081`.

Current status:

- Replicates the Go backend's main behavior: OAuth login, JWT auth, short-link CRUD, redirect handling, soft delete, restore, permanent delete, and click tracking.
- Uses port `8081` by default, while the Go backend uses `8080`.
- Exposes workspace scripts through `pnpm --filter backend-ktor ...` and local Gradle entrypoints through `./gradlew ...`.
- Is still experimental and should be treated as a parallel implementation, not the primary source of database setup.

Important limitation:

- `apps/backend-ktor` does not currently provide its own migration command or schema bootstrap.
- Before running Ktor against a fresh database, initialize the database with the Go backend migration flow first:

```bash
pnpm --filter backend-chi db:migrate
```

- This is required because Ktor expects the existing PostgreSQL schema, including `provider_enum`, `users`, `links`, and related indexes.

Ktor local setup:

```bash
cp apps/backend-ktor/.env.sample apps/backend-ktor/.env
pnpm --filter backend-ktor dev
```

Local config notes:

- `BACKEND_PUBLIC_URL` should match the Ktor server URL used by OAuth callbacks, usually `http://localhost:8081`.
- `FRONTEND_REDIRECT_URL` should point to the frontend callback page, usually `http://localhost:5173/auth/callback`.
- `JWT_EXPIRY` is parsed as an integer hour value. Do not append inline comments to the value in `.env`.
- The frontend should point to the Ktor API with `VITE_API_BASE_URL=http://localhost:8081/api/v1/` when you want to test the Kotlin backend.

Ktor stack:

- Kotlin 2.3
- Ktor 3
- Gradle
- PostgreSQL
- Exposed
- kotlinx.serialization
- Ktor OAuth and JWT auth
- Apache5 Ktor HTTP client
- java-dotenv
- ktlint

Ktor project structure:

```text
apps/backend-ktor
├── package.json                         # pnpm workspace scripts for Gradle tasks
├── build.gradle.kts                    # Ktor, Kotlin, Exposed, and tooling dependencies
├── src/main/resources/application.yaml # Ktor app config, port, JWT metadata, OAuth endpoints
└── src/main/kotlin/dev/danielslab/shorturl
    ├── Application.kt                  # Application bootstrap and dependency wiring
    ├── config                          # Environment-backed runtime config
    ├── db                              # Exposed database wiring and table schema definitions
    ├── domain                          # Core domain models
    ├── dto                             # Request/response and OAuth DTOs
    ├── plugins                         # Ktor plugin setup such as auth, CORS, logging, status pages
    ├── repository                      # PostgreSQL repository implementation
    ├── routes                          # HTTP routes for users, auth, links, and redirects
    ├── service                         # Service-layer helpers
    └── utils                           # Shared helpers such as OAuth HTTP client utilities
```

Useful commands:

```bash
pnpm --filter backend-ktor dev
pnpm --filter backend-ktor test
pnpm --filter backend-ktor build
pnpm --filter backend-ktor jar
pnpm --filter backend-ktor docker:build
```

## License

MIT
