# db4web — Agent Instructions

## Vision

A complete web-based database development platform in the xBase (dBase/FoxPro/Clipper) tradition. Developers write `.prg` files and design forms visually. The platform compiles/runs these as web applications backed by SQLite3.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (`cmd/db4web/main.go`) |
| Frontend | React SPA (`frontend/`) via Vite |
| Database | SQLite3 via `mattn/go-sqlite3` |
| Language | xBase-like `.prg` files |
| Forms | Visual designer, stored as structured data |

## Directory Layout

```
cmd/db4web/main.go       — server entrypoint
internal/
  compiler/              — .prg lexer, parser, codegen
  runtime/               — executes compiled xBase programs
  form/                  — form definitions engine
  sqlite/                — SQLite3 connection/wrapper
  server/                — HTTP routes, middleware
frontend/
  src/                   — React SPA (Vite + TypeScript)
  index.html
examples/                — sample .prg files and form defs
  schema.sql             — full database schema (6 tables, 1 view)
  main.prg               — entry point, main menu, DB init
  customer.prg           — CRUD, search, risk assessment
  appointment.prg        — scheduling, service catalog
  invoice.prg            — billing, payment, collections, dunning
  forms/
    customer.json        — customer entry form (12 fields, 4 buttons)
    appointment.json     — appointment scheduler (13 fields, 4 buttons)
    invoice.json         — invoice view (13 fields, 4 buttons)
docs/                    — design docs
```

## Commands

### Go Backend

```bash
go run ./cmd/db4web          # start dev server (port 8080, $PORT)
go build ./cmd/db4web        # build binary
go test ./...                # all tests
go vet ./...                 # static analysis
```

Environment: `PORT` (default 8080), `DB4WEB_DB` (default `./data.db`).

### Frontend

```bash
cd frontend && npm run dev      # Vite dev server, proxies /api -> :8080
cd frontend && npm run build    # production build -> frontend/dist/
cd frontend && npm run lint     # ESLint
cd frontend && npm run typecheck  # tsc --noEmit
```

### Verification Order

For any change: `go vet ./...` → `go test ./...` → `npm run typecheck` → `npm run lint`.

## Architecture

- **Stateless runtime**: xBase program execution state is per-request isolated.
- **Form field sizing**: field width derives from the format mask string (e.g. `"(XXX)XXX-XXXX"` → width 14). No absolute positioning — flow layout. Format mask characters: `X` = any char, `9` = digit, `A` = alpha, `!` = uppercase, other chars are literal.
- **`.prg` files** contain xBase procedures/functions, form event handlers, and data logic.
- **Forms** are designed visually, stored as JSON, and rendered as React components.
- **Data binding**: form fields bind via `ALIAS->FIELD` syntax (e.g. `Customer->Phone`).

## Conventions

- Go: standard `internal/` package visibility. `error` values returned, no panics.
- Frontend: TypeScript strict mode. React functional components, hooks.
- No generated code committed. Artifacts in `frontend/dist/` are gitignored.
- `.prg` files use Clipper-compatible syntax where possible.

## Key Constraints

- SQLite3 is the only backend. No network DB layer.
- Forms must not use absolute positioning. Layout is flow-based with smart field sizing from masks.
- The platform must be self-contained: a developer builds forms and logic entirely within the tool, like Clipper.
