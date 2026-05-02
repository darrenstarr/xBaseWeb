# db4web — Agent Instructions

## Vision

A complete web-based database development platform in the xBase (dBase/FoxPro/Clipper) tradition. Developers write `.prg` files and design forms visually. The platform interprets/runs these as web applications backed by SQLite3.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (`cmd/db4web/main.go`) |
| Frontend | React SPA (`frontend/`) via Vite |
| Database | SQLite3 via `mattn/go-sqlite3` |
| Language | xBase-like `.prg` files (interpreted, not compiled) |
| Forms | Defined in `.prg` via `@ SAY...GET` or form JSON |

## Directory Layout

```
cmd/db4web/main.go       — server entrypoint
internal/
  compiler/              — .prg lexer, parser, AST (378 tests)
  runtime/               — interpreter, value system, screen protocol
    runtime.go           — AST walker, expression evaluator, DB operations
    value.go             — Value types (N, C, L, D, A, O) with arithmetic
    workarea.go          — Work area manager (225 areas, aliased)
    screen.go            — Screen protocol (lines, fields, table, nav, confirm)
  form/                  — Field mask parsing, validation, formatting
  sqlite/                — SQLite3 wrapper (WAL, foreign keys, transactions)
  server/                — HTTP routes, middleware, generic CRUD
    server.go            — Routes, middleware, generic /api/data/{table} CRUD
    crud.go              — Interpreter execution, RUNSQL, pagination
frontend/
  src/main.tsx           — Single-file terminal renderer (no app logic)
  index.html
examples/cureforwoke/    — Demo app: "The DeSantis Cure for Woke"
  app.prg                — All app logic (menus, CRUD, forms, lists)
  app.json               — Theme config only
  schema.sql             — Database schema (6 tables, 1 view)
  forms/                 — Form JSON definitions (optional)
docs/
  architecture.md        — Full architecture document
```

## Commands

### Go Backend

```bash
go run ./cmd/db4web          # start dev server (port 8080)
go build ./cmd/db4web        # build binary
go test ./...                # all 378+ tests
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

`go vet ./...` → `go test ./...` → `npm run typecheck` → `npm run lint`.

## Architecture

- **Stateless interpreter**: Each `POST /api/execute` call restarts the procedure. Two-phase forms use `IF var == "" → show form → READ → RETURN / ELSE → save` pattern.
- **Screen protocol**: Interpreter returns a `Screen` JSON with `lines`, `fields`, `table`, `nav`, `confirm`, `prompt`, `wait`, `done` — the frontend renders generically.
- **Navigation**: `NAV "1" -> "CustomerMenu"` in `.prg` defines menu choice → procedure mapping. Frontend reads `screen.nav`. A `procStack` tracks call history for back/cancel.
- **CONFIRM**: `CONFIRM "message"` pauses execution, shows a confirmation banner, sets `_confirm` variable in the next call.
- **RUNSQL**: Executes a SQL query, returns structured `TableData` with columns, rows, actions, pagination info. Frontend renders it as a scrollable HTML table.
- **ACTIONS**: `RUNSQL ... ACTIONS "Edit" -> "EditProc"` adds per-row buttons. Frontend calls `runInterpreter(procedure, { mId: key, ... })` on click.
- **Generic CRUD**: `GET/POST/PUT/DELETE /api/data/{table}` and `/api/data/{table}/{id}` work on any SQLite table — no hardcoded endpoints.
- **Browser history**: Each `runInterpreter` call pushes `window.history.state`. `popstate` handler navigates within app history, not out of the SPA.
- **Field masks**: `X` = any char, `9` = digit, `A` = alpha, `!` = uppercase, others = literals. Mask formatting strips/cleans on each keystroke to prevent double-formatting.

## Language Extensions (beyond standard Clipper)

| Statement | Purpose |
|-----------|---------|
| `RUNSQL "SELECT ..." COLUMNS "A","B" ACTIONS "E"->"P"` | Execute SQL, return table with actions |
| `NAV "1" -> "Proc", "0" -> "BACK"` | Define menu navigation map |
| `CONFIRM "message"` | Show confirmation dialog, sets `_confirm` |
| `SET TITLE TO "..."` | Set application title in the browser header |
| `SET TAGLINE TO "..."` | Set application tagline |
| `GO expr` | Go to record (evaluates expression like `VAL(mId)`) |

## Built-in Functions

`VAL()`, `STR()`, `UPPER()`, `EMPTY()`, `FOUND()`, `EOF()`, `BOF()`, `RECNO()`, `DATE()`, `DTOC()`, `LEFT()`, `IIF()`, `INT()`, `ALLTRIM()`, `TRIM()`

## Key Constraints

- SQLite3 is the only backend. No network DB layer.
- Forms use flow layout with smart field sizing from masks. No absolute positioning.
- All application logic lives in `.prg` files. Frontend has zero domain knowledge.
- Go backend has only infrastructure routes — no application-specific CRUD handlers.
- The platform is self-contained: a developer builds everything in `.prg` files within the tool.
