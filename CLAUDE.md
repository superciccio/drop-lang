# Drop

A zero-ceremony prototyping language. Write a `.drop` file, run `drop file.drop`, get a running prototype.

## Stack

- **Language**: Go
- **Architecture**: Lexer -> Parser -> AST -> Tree-walk interpreter
- **Distribution**: Single binary, no runtime dependencies

## Project Structure

```
drop/
├── cmd/drop/          -- CLI entrypoint
├── internal/
│   ├── lexer/         -- Tokenizer
│   ├── parser/        -- AST generation
│   ├── ast/           -- Node types
│   ├── interpreter/   -- Tree-walk evaluator
│   ├── web/           -- Built-in HTTP server + routing
│   ├── store/         -- JSON-backed data persistence
│   └── ui/            -- Auto-render + page block HTML generation + default CSS
├── examples/          -- Example .drop files
└── docs/plans/        -- Design documents
```

## Build Phases

1. Lexer & parser (variables, functions, if/for, pipes, comments)
2. Built-in web server (serve, get, post, respond, body, route params)
3. Data layer (store, save, load, remove — JSON files on disk)
4. UI rendering (auto-render for data + page blocks for custom UI)
5. Default stylesheet (dark theme, system fonts, responsive)
6. Pipe operators (filter, map, sort)
7. Fetch (external HTTP requests)
8. CLI polish (error messages, help, hot reload)

## Design Doc

Full language spec: [docs/plans/2026-03-31-drop-language-design.md](docs/plans/2026-03-31-drop-language-design.md)
