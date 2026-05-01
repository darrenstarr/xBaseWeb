# db4web Architecture

## Overview

db4web is a web-based database development platform in the xBase tradition
(dBase/FoxPro/Clipper). Developers write `.prg` files and design forms as
structured JSON. The platform interprets these in real time to generate
web applications backed by SQLite3.

## Principles

1. **xBase drives everything.** The UI is not hand-coded. All screens,
   forms, lists, and navigation derive from xBase artifacts (`.prg` files
   and form JSON definitions). The generic React renderer has zero domain
   knowledge — it reads the workspace definition and renders accordingly.

2. **Forms are structured data.** A form is a JSON document describing
   fields (with masks, types, bindings, validation), buttons, and event
   handlers that reference `.prg` procedures. There is no absolute
   positioning — layout is flow-based with field width derived from
   the format mask (e.g., `"(XXX)XXX-XXXX"` → width 14).

3. **The .prg runtime is an interpreter.** Rather than compiling xBase
   to Go or WASM, the runtime parses and executes `.prg` files on each
   request. This keeps the development loop tight: edit a `.prg`, reload
   the browser, see the change. The interpreter walks the AST directly.

4. **SQLite3 is the only backend.** No separate database server. The
   platform is self-contained.

## Source Map

```
cmd/db4web/main.go         HTTP server entrypoint (port 8080)
internal/
  compiler/                .prg lexer, parser, AST
    token.go               100+ token types
    lexer.go               Character-by-character tokenizer
    ast.go                 25+ AST node types
    parser.go              Recursive-descent parser
  runtime/                 .prg interpreter
    value.go               Value types (N, C, L, D, A, O) with operations
    workarea.go            Work area manager (225 areas, aliased)
    runtime.go             AST walker, expression evaluator
  form/                    Form definitions engine
    mask.go                Field mask parsing, validation, formatting
    form.go                Form/field/button definition types
  sqlite/                  SQLite3 wrapper
    db.go                  Open, Exec, Query, Tables, Columns, transactions
  server/                  HTTP layer
    server.go              Routes, middleware, JSON helpers
    crud.go                CRUD handlers + interpreter endpoint
examples/                  Application artifacts (the "source code")
  app.json                 Workspace definition (nav, lists, dashboards, theme)
  schema.sql               Database schema (6 tables, 1 view)
  main.prg                 Entry point, main menu loop
  customer.prg             Customer CRUD, search, risk assessment
  appointment.prg          Scheduling, service catalog
  invoice.prg              Billing, payments, collections, dunning
  forms/                   Form definitions (JSON)
    customer.json          12 fields, 4 buttons
    appointment.json       13 fields, 4 buttons
    invoice.json           13 fields, 4 buttons
    service.json           7 fields, 2 buttons
frontend/                  Generic React renderer (no domain logic)
  src/main.tsx             Single ~400-line file — reads workspace, renders UI
```

## Data Flow

```
User Browser
    │
    ▼
React SPA (main.tsx)          ◄── Generic renderer. No hard-coded pages.
    │
    ├── GET /api/workspace    ──►  examples/app.json
    │                              (nav, lists, dashboards, theme)
    │
    ├── GET /api/forms/:name  ──►  examples/forms/{name}.json
    │                              (field definitions, buttons, events)
    │
    ├── GET /api/query?sql=   ──►  SQLite3 (list data, dashboard stats)
    │
    ├── POST /api/:resource   ──►  SQLite3 INSERT (create record)
    ├── PUT  /api/:resource/id──►  SQLite3 UPDATE (update record)
    ├── DELETE /api/:resource/id──► SQLite3 DELETE (delete record)
    │
    └── POST /api/execute     ──►  .prg interpreter
                                       │
                                       ├── Lexer → Tokens
                                       ├── Parser → AST
                                       └── Runtime → walk AST, execute
```

## The Interpreter

`POST /api/execute` accepts:

```json
{
  "program": "examples/main.prg",
  "procedure": "MainMenu",
  "state": { "mChoice": "1" }
}
```

The runtime:
1. Reads and parses the `.prg` file
2. Finds the requested procedure
3. Walks the AST executing each statement
4. `@ SAY...GET` produces form field definitions
5. `DO WHILE .T.` menus produce navigation options
6. `REPLACE`, `APPEND BLANK`, etc. produce SQL operations
7. Returns execution result plus updated state

The frontend interprets the result:
- `{ "screen": "menu", "items": [...] }` → render menu
- `{ "screen": "form", "fields": [...] }` → render form
- `{ "screen": "done", "result": "..." }` → show result

This allows `.prg` files to generate the UI dynamically, just like
Clipper did — the `@ 10,5 SAY "Name:" GET mName` commands define
the form, and the runtime tells the frontend what to render.

## Current State

- **Workspace definition**: Fully functional. Navigation, lists, dashboards,
  and theme all driven by `app.json`.
- **Form rendering**: Fully functional. Generic FormView renders any form
  JSON definition. CRUD operations work via API endpoints.
- **List views**: Fully functional. Queries + column definitions drive
  table rendering. Row click opens edit form.
- **Dashboard**: Fully functional. Stat widgets and table widgets driven
  by SQL queries.
- **Service CRUD**: Complete — create, read, update, delete via form UI.
- **.prg interpreter**: Initial implementation. Can parse and execute
  basic xBase programs. `@ SAY...GET`, control flow, expressions work.
  Not yet wired to the frontend as the primary UI driver.

## Roadmap

1. Wire interpreter output to frontend — menus and forms generated
   directly from `.prg` execution instead of from `app.json`/form JSON
2. Full Clipper-compatible expression evaluation (all functions)
3. Report engine (xBase `REPORT FORM` equivalent)
4. Visual form designer (drag-and-drop fields)
5. Multi-user with row-level locking

## Key Constraints

- SQLite3 is the only data backend. No PostgreSQL, no MySQL, no network DB.
- Forms use flow layout with smart field sizing from masks. No absolute positioning.
- The platform is self-contained: a developer builds the entire application
  within the tool using `.prg` files and the form designer.
- All UI must be generatable from xBase artifacts. No hand-coded frontend pages.
