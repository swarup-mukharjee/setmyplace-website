# Set My Place — backend (PostgreSQL edition)

A small Go server with one external dependency: `github.com/lib/pq`, the
Postgres driver. It does three things:

1. Serves the website (the HTML files in the sibling `frontend` folder).
2. Exposes an API so the contact form on `contact.html` saves submissions
   into a Postgres database.
3. Creates its own database table automatically on first run — no separate
   migration tool needed.

## Requirements

- Go 1.21 or newer. Check with `go version`. Install from https://go.dev/dl/
- A Postgres database (14+ recommended). Either:
  - a local install (`postgres.app`, Homebrew, apt, etc.), or
  - Docker, using the included `docker-compose.yml` (easiest), or
  - a managed database (Supabase, Neon, Railway, RDS, etc.)

## Folder layout expected

```
your-project/
├── assets/  (style.css, main.js, logo.png, icon.png)
├── frontend/
│   ├── index.html
│   ├── services.html / how-it-works.html / why-us.html / partner.html / contact.html
│   └── setmyplaceadminsuper.html
└── backend/
    ├── main.go
    ├── go.mod
    ├── go.sum          <- generated the first time you run `go mod tidy`
    ├── docker-compose.yml
    ├── .env.example
    └── README.md
```

The server serves website files from `../frontend` by default, relative to
the `backend` folder. Set `SITE_DIR` if your deployment uses another path.

## 1. Get a Postgres database running

**Option A — Docker (fastest for local dev):**

```bash
cd backend
docker compose up -d
```

This starts Postgres on `localhost:5432` with user `setmyplace`, password
`setmyplace`, database `setmyplace` — matching the example connection
string below, so no further config needed for local testing.

**Option B — your own Postgres install or a managed service:**
create an empty database and note its connection string. It looks like:

```
postgres://USER:PASSWORD@HOST:5432/DATABASE?sslmode=disable
```

(Managed providers like Supabase/Neon usually give you this string directly
in their dashboard — for those, use `sslmode=require` instead of `disable`.)

## 2. Set your environment variables

Copy `.env.example` for reference, then export the values (there's no
`.env`-loading library wired in — these are read with plain `os.Getenv`):

```bash
export DATABASE_URL="postgres://setmyplace:setmyplace@localhost:5432/setmyplace?sslmode=disable"
export ADMIN_KEY="pick-a-long-random-string-here"
```

On Windows (PowerShell):

```powershell
$env:DATABASE_URL="postgres://setmyplace:setmyplace@localhost:5432/setmyplace?sslmode=disable"
$env:ADMIN_KEY="pick-a-long-random-string-here"
```

If `DATABASE_URL` isn't set, the server refuses to start and tells you so.
If `ADMIN_KEY` isn't set, it falls back to `changeme` and prints a warning —
fine for local testing, not fine once this is live on the internet.

## 3. Fetch the Postgres driver and run it

```bash
cd backend
go mod tidy      # downloads github.com/lib/pq and writes go.sum
go run main.go
```

You should see:

```
Connected to Postgres.
Set My Place backend running at http://localhost:8080
Serving website files from: ..
```

Open **http://localhost:8080** — that's your whole website, now with a
contact form that writes to Postgres.

The `contacts` table is created automatically the first time the server
connects, via `CREATE TABLE IF NOT EXISTS` — nothing to run by hand.

## Viewing submissions

Open **http://localhost:8080/admin.html**, paste in your `ADMIN_KEY`, click
"Load submissions" — a table of every enquiry, newest first.

Or query the database directly:

```sql
SELECT * FROM contacts ORDER BY created_at DESC;
```

Or hit the API:

```bash
curl -H "X-Admin-Key: your-key" http://localhost:8080/api/contacts
```

## Health check

`GET /api/health` returns `{"status":"ok"}` if the server can reach
Postgres — useful for uptime monitors or container health checks.

## Building a single binary (for deploying to a server)

```bash
cd backend
go build -o setmyplace-server main.go
```

Copy that binary plus the whole project folder (so `SITE_DIR` still finds
your frontend files) to your server:

```bash
DATABASE_URL="postgres://..." ADMIN_KEY="your-real-key" PORT=8080 ./setmyplace-server
```

## Deploying

- **A small VPS** — run the binary behind `systemd`, put Nginx/Caddy in
  front for HTTPS, point `DATABASE_URL` at either a local Postgres or a
  managed one.
- **Render, Railway, Fly.io** — most of these offer a managed Postgres
  add-on with one click, which gives you a `DATABASE_URL` to paste straight
  into the app's environment variables.
- **Docker** — example multi-stage Dockerfile:

  ```dockerfile
  FROM golang:1.21 AS build
  WORKDIR /app
  COPY backend/ ./backend/
  COPY . .
  RUN cd backend && go mod download && go build -o /setmyplace-server main.go

  FROM debian:bookworm-slim
  COPY --from=build /setmyplace-server /setmyplace-server
  COPY --from=build /app /app
  WORKDIR /app/backend
  ENV PORT=8080
  EXPOSE 8080
  CMD ["/setmyplace-server"]
  ```

  Run alongside the `postgres` service from `docker-compose.yml`, or point
  `DATABASE_URL` at a managed database instead.

## API reference

### `POST /api/contact`

```json
{
  "name": "Priya Sharma",
  "phone": "+91 98765 43210",
  "service": "Full ready-to-live setup",
  "date": "2026-09-20",
  "message": "2BHK, need cleaning + furniture setup"
}
```

`name` and `phone` are required. Response: `{"success": true}` on success,
or `{"error": "..."}` with a 4xx/5xx status.

### `GET /api/contacts`

Requires `X-Admin-Key: <key>` header or `?key=<key>` query param. Returns a
JSON array of all submissions, newest first.

### `GET /api/health`

Returns `{"status":"ok"}` if Postgres is reachable, `503` otherwise.

## Database schema

```sql
CREATE TABLE contacts (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    phone          TEXT NOT NULL,
    service        TEXT,
    preferred_date TEXT,
    message        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip             TEXT
);
```

This is created for you automatically — the SQL above is just for
reference if you want to inspect or back up the table manually.

**Back up your Postgres database regularly** — most managed providers do
this automatically; for self-hosted Postgres, `pg_dump` on a schedule is
the simplest option.
