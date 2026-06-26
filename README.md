# Ticket System

A small backend service for a ticket system: users register, log in, create tickets,
and can only view/update tickets they own.

- **Language:** Go 1.26 (stdlib `net/http` only -- no web framework)
- **Auth:** JWT (`github.com/golang-jwt/jwt/v5`), passwords hashed with bcrypt
- **Storage:** in-memory (mutex-protected)
- **Port:** 8080

---

## Table of Contents

- [Project Structure](#project-structure)
- [Run Locally](#run-locally-without-docker)
- [Run with Docker](#run-locally-with-docker)
- [API Reference](#api-reference)
- [Assumptions](#assumptions)
- [Deployment](#deployment)
- [Manual End-to-End Test](#manual-end-to-end-test)

---

## Project Structure

```
.
├── main.go         # routes + server startup
├── handlers.go     # HTTP handlers, request/response shapes, validation
├── auth.go         # JWT issue/verify, bcrypt, auth middleware
├── store.go        # in-memory data store (mutex-protected maps)
├── models.go       # User/Ticket structs + status transition rules
├── Dockerfile
├── .env.example
└── go.mod / go.sum
```

---

## Run Locally (without Docker)

Requires Go 1.22+ (tested on Go 1.26.4)

```bash
go mod download
go run .
```

The server listens on `:8080`. Verify it:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Run Locally with Docker

```bash
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
```

Expected response:

```json
{
    "status": "ok"
}
```

To set a custom JWT secret instead of the built-in dev default:

```bash
docker run -p 8080:8080 -e JWT_SECRET=some-long-random-string ticket-system
```

---

## API Reference

All responses are JSON. Errors are `{"error": "message"}` with a meaningful HTTP status.

### `GET /health` (public)
Returns `{"status": "ok"}`.

### `POST /auth/register` (public)
Creates a user.

Request:
```json
{
    "email": "anakin@tatooine.jedi",
    "password": "sandislife"
}
```
- `email` must be non-empty and contain `@`.
- `password` must be at least 6 characters.

Responses:
- `201` -- user created, returns `{ "id", "email", "created_at" }` (no password fields, ever)
- `400` -- invalid/missing email or password
- `409` -- email already registered

### `POST /auth/login` (public)
Returns a JWT.

Request:
```json
{
    "email": "anakin@tatooine.jedi",
    "password": "sandislife"
}
```

Responses:
- `200` -- `{ "token": "<jwt>" }`
- `400` -- missing email/password
- `401` -- wrong email or password (deliberately generic -- doesn't reveal which)

Use the token on every protected call:
```
Authorization: Bearer <token>
```
Tokens expire after 72 hours.

### `POST /tickets` (protected)
Creates a ticket owned by the caller. New tickets always start as `open`.

Request:
```json
{
    "title": "Execute Order 66",
    "description": "Clone units have turned against the Jedi. Immediate action required."
}
```
- `title` is required (non-empty after trimming).
- `description` is optional.

Responses:
- `201` -- the created ticket
- `400` -- missing title
- `401` -- missing/invalid/expired token

### `GET /tickets` (protected)
Returns only the caller's own tickets, newest first, as a JSON array (`[]` if none).

### `GET /tickets/{id}` (protected)
Returns one ticket.

Responses:
- `200` -- the ticket (only if the caller owns it)
- `403` -- the ticket exists but belongs to someone else
- `404` -- no ticket with that id exists at all

### `PATCH /tickets/{id}/status` (protected)
Updates a ticket's status.

Request:
```json
{ 
    "status": "in_progress" 
}
```

Status flow:
```
open ──> in_progress ──> closed
 └──────────────────────> closed   (direct close is allowed)
```
`closed` is terminal -- it can never move to `open` or `in_progress` again.

Responses:
- `200` -- the updated ticket
- `400` -- `status` is not one of `open` / `in_progress` / `closed`
- `403` -- caller doesn't own the ticket
- `404` -- ticket doesn't exist
- `409` -- the transition isn't allowed (e.g. reopening a closed ticket, or any
  backward/no-op move)

---

## Ticket JSON Shape

```json
{
    "id": "54210dd82267e39b02f23a9c",
    "user_id": "d16c30a9706c23274acc42ac",
    "title": "Execute Order 66",
    "description": "Clone units have turned against the Jedi. Immediate action required.",
    "status": "in_progress",
    "created_at": "2026-06-25T07:40:45.79Z",
    "updated_at": "2026-06-25T07:40:45.79Z"
}
```

---

## Assumptions

The brief leaves a few details open to interpretation. Here's what was decided and why:

1. **Storage is in-memory**, explicitly allowed by the brief ("in-memory storage,
   SQLite, PostgreSQL, or any simple persistent store"). This keeps the service simple
   and dependency-free. Data resets on restart.
   `store.go` is a small, isolated interface, so swapping it for SQLite/Postgres later
   wouldn't touch any handler logic.
2. **Ownership violations return `403 Forbidden`** (ticket exists, but isn't yours)
   while a genuinely missing ticket returns `404`. This makes the ownership check
   explicit and separately testable from "not found."
3. **Login response field is `token`**: `{"token": "<jwt>"}`.
4. **Status transitions are forward-only.** `open -> in_progress -> closed` is the
   intended path. Skipping straight from `open -> closed` is also allowed (e.g.
   closing a duplicate/spam ticket immediately) since the brief never forbids it.
   Any backward move (`in_progress -> open`, `closed -> open`, `closed -> in_progress`)
   and any no-op move (setting a ticket to its current status) return `409 Conflict`.
5. **`409 Conflict`** is used for illegal status transitions (valid status value,
   wrong state to apply it from); **`400 Bad Request`** is reserved for malformed
   input (e.g. `status: "foo"`).
6. **Email is lower-cased and trimmed** before storage/lookup so `Alice@Example.com`
   and `alice@example.com` are treated as the same account.
7. **IDs** are random 24-character hex strings (12 random bytes), generated server-side.

---

## Deployment

**Deployed URL:** https://ticket-system-xkod.onrender.com

Test it:
```bash
curl https://ticket-system-xkod.onrender.com/health
```

Note: Render's free web services spin down after 15 minutes of inactivity; the
first request after idling takes ~30-60 seconds to wake up. This is normal and fine
for an assignment review -- just give the first request a moment.

---

## Manual End-to-End Test

```bash
BASE=http://localhost:8080

# Register user (Anakin Skywalker)
curl -s -X POST $BASE/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"anakin@tatooine.jedi","password":"sandislife"}'

# Login
TOKEN=$(curl -s -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"anakin@tatooine.jedi","password":"sandislife"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Create ticket
TICKET=$(curl -s -X POST $BASE/tickets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Execute Order 66","description":"Clone units have turned against the Jedi. Immediate action required."}')

echo $TICKET

# Extract ticket ID
TICKET_ID=$(echo $TICKET | grep -o '"id":"[^"]*' | cut -d'"' -f4)

# List tickets
curl -s $BASE/tickets \
  -H "Authorization: Bearer $TOKEN"

# Get single ticket
curl -s $BASE/tickets/$TICKET_ID \
  -H "Authorization: Bearer $TOKEN"

# Move ticket to in_progress
curl -s -X PATCH $BASE/tickets/$TICKET_ID/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'

# Close ticket
curl -s -X PATCH $BASE/tickets/$TICKET_ID/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"closed"}'