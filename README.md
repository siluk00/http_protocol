# http_protocol

A from-scratch HTTP/1.1 server implemented in Go over raw TCP — no `net/http`, no frameworks, no shortcuts.

This project exists to answer a question most developers never ask: *what actually happens between a TCP connection and an HTTP response?* Every byte of the wire format — request line, headers, body, chunked encoding — is parsed and written by hand.

-----

## Why this exists

Most Go developers reach for `net/http` without ever understanding what it does. This project is an exercise in **protocol fluency**: reading RFC 7230, understanding the wire format, and building the machinery from scratch. The result is a server that handles real HTTP traffic using nothing but `net.Listener` and raw byte parsing.

-----

## Architecture

```
cmd/
  httpserver/       ← entry point, signal handling, route definition
internal/
  server/           ← TCP listener, goroutine-per-connection, graceful shutdown
  request/          ← stateful request parser (request line → headers → body)
  response/         ← response writer with state machine enforcement
  headers/          ← RFC-compliant header parsing and storage
```

### How a request flows through the system

```
TCP connection
    │
    ▼
server.listen()          ← accepts connections, spawns goroutines
    │
    ▼
request.RequestFromReader()  ← reads bytes from conn, feeds stateful parser
    │  ┌─────────────────────────────────────┐
    │  │  state machine:                     │
    │  │  initialized → request line parsed  │
    │  │  → headers parsed                   │
    │  │  → body consumed (Content-Length)   │
    │  │  → done                             │
    │  └─────────────────────────────────────┘
    ▼
handler(writer, req)     ← user-defined handler receives parsed request
    │
    ▼
response.Writer          ← enforces write order: status → headers → body
    │
    ▼
TCP connection (raw bytes written back)
```

### Key design decisions

**Stateful parser with a byte buffer** — `RequestFromReader` reads from the connection in small chunks (starting at 8 bytes, doubling when needed). This simulates real network conditions where data arrives in fragments. The parser drives a state machine: it won’t accept headers before the request line, and won’t read the body before headers are complete. This is the same approach used in production HTTP parsers like `llhttp` (Node.js).

**Response writer as state machine** — `WriteStatusLine`, `WriteHeaders`, and `WriteBody` must be called in order. Call them out of order and you get an explicit error. This turns a class of bugs — missing headers, double-writes — into compile-time-adjacent errors rather than silent corruption.

**RFC 7230-compliant header parsing** — The `headers` package validates token characters per the spec, normalizes keys to lowercase, and implements the comma-folding rule for duplicate headers. This isn’t just string splitting: it rejects headers with spaces in the key, missing colons, or illegal characters.

**Goroutine-per-connection with WaitGroup** — Each connection gets its own goroutine. `sync.WaitGroup` tracks all active connections, and `atomic.Bool` coordinates shutdown so `Close()` waits for in-flight requests before returning. This is the standard Go concurrency pattern for servers.

-----

## What’s implemented

- [x] TCP listener with goroutine-per-connection concurrency
- [x] Request line parsing (method, target, HTTP version validation)
- [x] RFC 7230 header parsing (token validation, lowercase normalization, comma-folding)
- [x] Body parsing via `Content-Length`
- [x] Response writer with enforced write ordering
- [x] Graceful shutdown (`SIGINT`/`SIGTERM` → wait for active connections)
- [x] Dynamic read buffer (doubles when full, handles fragmented TCP delivery)

-----

## Roadmap

- [ ] **Chunked Transfer Encoding** — `Transfer-Encoding: chunked` for streaming responses without known `Content-Length`
- [ ] **Keep-Alive / persistent connections** — handle multiple requests per TCP connection (HTTP/1.1 default)
- [ ] **Router** — path and method-based dispatch with path parameters (e.g. `/users/{id}`)
- [ ] **TLS** — wrap the listener with `tls.Listen` for HTTPS
- [ ] **Request body streaming** — expose `Body` as `io.Reader` instead of buffering fully into memory
- [ ] **Timeout handling** — read deadlines per connection to prevent resource exhaustion
- [ ] **Full method support** — PUT, DELETE, PATCH, OPTIONS, HEAD

-----

## Running

```bash
git clone https://github.com/siluk00/http_protocol
cd http_protocol
go run ./cmd/httpserver
```

The server starts on port `42069`. Test it with curl:

```bash
curl -v http://localhost:42069/
curl -v http://localhost:42069/yourproblem
curl -v http://localhost:42069/myproblem
```

Stop with `Ctrl+C` — the server waits for in-flight requests to finish before exiting.

-----

## Testing

```bash
go test ./internal/request/...
```

The request package has unit tests covering partial reads, malformed headers, and body length validation — the edge cases that matter in a byte-level parser.

-----

## What I learned

Building this forced me to think at a level most HTTP work never reaches:

- HTTP is a text protocol over a byte stream. There are no message boundaries — you have to find them yourself using `\r\n` delimiters and `Content-Length`.
- TCP delivers data in arbitrary chunks. A parser that assumes it gets one complete request per `Read()` call is wrong. The stateful buffer approach here handles the realistic case.
- The RFC isn’t just bureaucratic overhead. The header token rules exist because spaces in header names caused real security vulnerabilities (HTTP request smuggling).
- Go’s `io.Reader` interface is the right abstraction for this: the parser doesn’t know or care whether it’s reading from a TCP connection, a file, or a test string — and the tests take advantage of that.
