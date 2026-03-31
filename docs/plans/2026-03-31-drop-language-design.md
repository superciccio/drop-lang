# Drop Language Design

> A zero-ceremony prototyping language. If you're typing anything that isn't your idea, Drop has failed.

## Overview

Drop is an interpreted language built in Go. Its only job is to take an idea from your head to a running prototype in seconds. No project setup, no imports, no package managers, no boilerplate.

Write a `.drop` file. Run `drop file.drop`. Done.

## Goals

- **Zero setup**: no init, no config, no dependencies
- **Zero cognitive overhead**: no types, no imports, no error handling ceremony
- **Batteries included**: HTTP, data storage, UI — built into the language
- **Readable as pseudocode**: if you can read English, you can read Drop
- **Single binary**: download one file, you have the whole language

## Non-goals

- Performance (it's for prototypes, not production)
- Completeness (missing a feature? you've graduated to a real language)
- Extensibility (no plugin system, no package manager)

## Syntax

### Variables

No declarations, no types. Everything is inferred.

```
name = "Andrea"
age = 30
tags = ["fast", "simple", "fun"]
user = {name: "Andrea", role: "builder"}
```

### String interpolation

Curly braces inside quotes. No prefix, no template syntax.

```
msg = "Hello {name}, you are {age}"
```

### Functions

The `do` keyword. Minimal.

```
do greet(name)
  return "Hello, {name}!"
```

### Conditionals and loops

```
if age > 18
  respond "Welcome"
else
  respond "Too young"

for user in users
  print user.name
```

### Pipes

Chain transformations left-to-right.

```
fetch "https://api.example.com/users"
  | filter .active
  | map .name
  | sort
  | respond
```

### Comments

Double dash.

```
-- this is a comment
```

## Web

HTTP is built into the language. No imports, no framework.

```
serve 8080

get "/hello"
  respond "Hello, World!"

post "/todos"
  save "todos" body
  respond body 201

get "/secret"
  respond "Nope" 403
```

- `serve <port>` starts the server
- `get`, `post`, `put`, `delete` define routes
- `respond <data> [status]` sends a response. Default status is 200.
- `body` is a magic keyword for the request payload (auto-parsed JSON)
- Route params via `:param` syntax: `get "/users/:id"`
- No headers, no content-type. Drop figures it out:
  - String? `text/html`
  - List/object? `application/json`
  - Error? Rendered error page
- `fetch <url>` makes HTTP GET requests and returns parsed data

## Data

Zero-config persistence. JSON files on disk. No database, no ORM, no SQL.

```
store "todos"

post "/todos"
  save "todos" body
  respond body 201

get "/todos"
  respond load "todos"

delete "/todos/:id"
  remove "todos" id
  respond "Deleted" 200
```

- `store "name"` declares a data store (creates a JSON file)
- `save "store" data` appends to the store
- `load "store"` returns all items
- `remove "store" id` deletes by auto-assigned ID

Every saved item gets an auto-generated `id` field.

## UI

### Auto-render (default)

When a browser hits a route, Drop auto-renders data as a clean HTML page:
- Lists become tables
- Objects become key-value cards
- Strings become text blocks

No configuration. Just works.

### Page blocks (custom UI)

The `page` keyword for when you want control:

```
get "/"
  page "My Todos"
    each todo in load "todos"
      row
        text todo.name
        button "Done" -> delete "todos" todo.id
    form "Add Todo"
      input "name"
      submit "Add" -> post "/todos"
```

UI keywords: `page`, `text`, `row`, `each`, `button`, `form`, `input`, `submit`, `link`, `image`.

The `->` operator connects UI elements to route actions. No JavaScript, no event handlers.

### Default styling

Every Drop app gets a built-in stylesheet. No opt-in, no config.

- System font stack (clean sans-serif)
- Dark theme: `#0a0a0a` background, `#ededed` text
- Centered content, max 720px width
- Styled tables with subtle borders and alternating rows
- Clean form inputs with focus states
- Solid accent-color buttons with hover states
- Card layouts for auto-rendered objects
- Responsive: stacks on mobile automatically

The goal: looks like a clean dev tool. Not beautiful, not ugly. Professional enough to show a colleague.

## Implementation

- **Language**: Go
- **Architecture**: Lexer -> Parser -> AST -> Tree-walk interpreter
- **Phases**:
  1. Lexer & parser (core syntax: variables, functions, if/for, pipes)
  2. Built-in web server (serve, get, post, respond)
  3. Data layer (store, save, load, remove)
  4. UI rendering (auto-render + page blocks)
  5. Default stylesheet
  6. Pipe operators (filter, map, sort)
  7. Fetch (external HTTP)
  8. CLI polish (error messages, help, hot reload)

## Example: Full Todo App

```
serve 8080
store "todos"

get "/"
  page "My Todos"
    each todo in load "todos"
      row
        text todo.name
        button "Delete" -> delete "/todos/{todo.id}"
    form "New Todo"
      input "name"
      submit "Add" -> post "/todos"

post "/todos"
  save "todos" body
  respond body 201

delete "/todos/:id"
  remove "todos" id
  respond "Deleted"

get "/api/todos"
  respond load "todos"
```

That's a full CRUD app with a UI. ~20 lines. No setup. No dependencies.
