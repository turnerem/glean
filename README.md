# Glean

Grab text from anywhere and save it as notes.

Use a keyboard shortcut (`Cmd+Shift+G`), and any text in your clipboard gets captured to a note on your desktop. Add tags, and the app automatically detects the source (URL from browser, etc.).

## Features

- **Global shortcut** (`Cmd+Shift+G`) to capture clipboard text
- **Origin detection** - Automatically detects URL from Safari, Chrome, Arc browsers
- **Tagging** - Add tags from existing list or create new ones
- **Local storage** - SQLite database, works offline
- **Sync** (optional) - Sync notes across devices via the API server

## Prerequisites

- **Node.js** 20+ and **pnpm** 9+
- **Rust** (install via [rustup](https://rustup.rs/))
- **Go** 1.22+ (for the sync server)
- **PostgreSQL** (optional, for sync server persistence)

## Installation

```bash
# Clone the repository
git clone https://github.com/your-org/glean.git
cd glean

# Install pnpm if not already installed
npm install -g pnpm

# Install dependencies
pnpm install

# Install Rust if not already installed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

## Running the Desktop App

```bash
# Development mode (hot reload)
cd apps/desktop
pnpm tauri dev

# Build for production
pnpm tauri build
```

The app will open and you can:

1. Copy some text to your clipboard
2. Press `Cmd+Shift+G` (or click the capture button)
3. Add tags and edit the source if needed
4. Press `Cmd+Enter` or click Save

## Running the Sync Server (Optional)

The sync server allows you to sync notes across devices.

### Quick Start (In-Memory Storage)

For testing without PostgreSQL:

```bash
cd apps/api
go run ./cmd/server
```

Server runs at `http://localhost:8080`. Note: Data is lost when server restarts.

### Production (PostgreSQL)

1. **Set up PostgreSQL database:**

```bash
# Create database
createdb glean

# Run migrations
psql -d glean -f migrations/001_init.sql
```

2. **Run server with DATABASE_URL:**

```bash
cd apps/api

# Set environment variables
export DATABASE_URL="postgres://username:password@localhost:5432/glean?sslmode=disable"
export JWT_SECRET="your-secure-secret-here"  # Change in production!
export PORT=8080

go run ./cmd/server
```

### Environment Variables

| Variable       | Description                     | Default                |
| -------------- | ------------------------------- | ---------------------- |
| `DATABASE_URL` | PostgreSQL connection string    | (in-memory if not set) |
| `JWT_SECRET`   | Secret for JWT signing          | `glean-dev-secret...`  |
| `PORT`         | Server port                     | `8080`                 |
| `GIN_MODE`     | Gin mode (`debug` or `release`) | `release`              |

### API Endpoints

| Method | Endpoint                | Description       | Auth |
| ------ | ----------------------- | ----------------- | ---- |
| POST   | `/api/v1/auth/register` | Register new user | No   |
| POST   | `/api/v1/auth/login`    | Login             | No   |
| GET    | `/api/v1/notes`         | Get all notes     | Yes  |
| POST   | `/api/v1/notes`         | Create note       | Yes  |
| PUT    | `/api/v1/notes/:id`     | Update note       | Yes  |
| DELETE | `/api/v1/notes/:id`     | Delete note       | Yes  |
| GET    | `/api/v1/tags`          | Get all tags      | Yes  |
| POST   | `/api/v1/tags`          | Create tag        | Yes  |
| POST   | `/api/v1/sync`          | Sync notes        | Yes  |
| GET    | `/health`               | Health check      | No   |

## Building

### Desktop App

```bash
cd apps/desktop
pnpm tauri build
```

Built app will be in `apps/desktop/src-tauri/target/release/bundle/`.

### API Server

```bash
cd apps/api
go build -o bin/server ./cmd/server
```

## Testing

```bash
# Rust tests (desktop app)
cd apps/desktop/src-tauri
cargo test

# Go tests (API server)
cd apps/api
go test ./...

# Svelte type checking
cd apps/desktop
pnpm check
```

## Project Structure

```
glean/
├── apps/
│   ├── desktop/              # Tauri desktop app
│   │   ├── src/              # Svelte frontend
│   │   │   ├── lib/
│   │   │   │   ├── components/   # UI components
│   │   │   │   ├── stores/       # Svelte stores
│   │   │   │   └── types.ts      # TypeScript types
│   │   │   └── App.svelte
│   │   ├── src-tauri/        # Rust backend
│   │   │   └── src/
│   │   │       ├── commands/     # Tauri commands
│   │   │       ├── db/           # SQLite operations
│   │   │       └── origin/       # Origin detection
│   │   └── package.json
│   └── api/                  # Go sync server
│       ├── cmd/server/       # Main entry point
│       ├── internal/
│       │   ├── handlers/     # HTTP handlers
│       │   ├── middleware/   # Auth, logging, CORS
│       │   ├── models/       # Data models
│       │   └── repository/   # Database layer
│       ├── migrations/       # SQL migrations
│       └── go.mod
├── package.json              # pnpm workspace root
└── pnpm-workspace.yaml
```

## Tech Stack

- **Desktop**: [Tauri](https://tauri.app/) v2 (Rust) + [Svelte](https://svelte.dev/) 5
- **API**: [Go](https://go.dev/) with [Gin](https://gin-gonic.com/)
- **Databases**: SQLite (local), PostgreSQL (sync server)
- **Auth**: JWT tokens

## License

MIT
